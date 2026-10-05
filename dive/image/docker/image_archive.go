package docker

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"

	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
)

type ImageArchive struct {
	manifest  manifest
	manifests []manifest
	config    config
	layerMap  map[string]*filetree.FileTree
}

func normalizeLayerPath(p string) string {
	cleaned := path.Clean(strings.TrimPrefix(p, "/"))
	return strings.TrimPrefix(cleaned, "./")
}

func extractLayerID(p string) string {
	p = normalizeLayerPath(p)
	if strings.HasPrefix(p, "blobs/sha256/") {
		return strings.TrimPrefix(p, "blobs/sha256/")
	}
	if strings.HasPrefix(p, "blobs/") {
		return path.Base(p)
	}
	if strings.HasPrefix(p, "sha256:") {
		return strings.TrimPrefix(p, "sha256:")
	}
	if strings.HasSuffix(p, "/layer.tar") {
		return strings.TrimSuffix(p, "/layer.tar")
	}
	if strings.HasSuffix(p, ".tar.gz") {
		return strings.TrimSuffix(path.Base(p), ".tar.gz")
	}
	if strings.HasSuffix(p, ".tgz") {
		return strings.TrimSuffix(path.Base(p), ".tgz")
	}
	if strings.HasSuffix(p, ".tar") {
		return strings.TrimSuffix(path.Base(p), ".tar")
	}
	if strings.HasSuffix(p, ".json") {
		return strings.TrimSuffix(path.Base(p), ".json")
	}
	return path.Base(p)
}

func getConfig(jsonFiles map[string][]byte, configPath string) ([]byte, bool) {
	if content, ok := jsonFiles[configPath]; ok {
		return content, true
	}
	norm := normalizeLayerPath(configPath)
	if content, ok := jsonFiles[norm]; ok {
		return content, true
	}
	id := extractLayerID(configPath)
	if id != "" && id != "." {
		candidates := []string{
			id,
			id + ".json",
			"blobs/sha256/" + id,
			"sha256:" + id,
			"blobs/" + id,
		}
		for _, c := range candidates {
			if content, ok := jsonFiles[c]; ok {
				return content, true
			}
		}
		for k, content := range jsonFiles {
			if extractLayerID(k) == id || strings.TrimSuffix(path.Base(k), ".json") == id {
				return content, true
			}
		}
	}
	return nil, false
}

func (img *ImageArchive) addLayer(tree *filetree.FileTree) {
	img.addLayerWithName(tree.Name, tree)
}

func (img *ImageArchive) addLayerWithName(name string, tree *filetree.FileTree) {
	if tree.Name != name {
		namedTree := *tree
		namedTree.Name = name
		tree = &namedTree
	}

	img.layerMap[name] = tree

	norm := normalizeLayerPath(name)
	img.layerMap[norm] = tree

	id := extractLayerID(name)
	if id != "" && id != "." {
		img.layerMap[id] = tree
		img.layerMap["blobs/sha256/"+id] = tree
		img.layerMap[id+"/layer.tar"] = tree
		img.layerMap["sha256:"+id] = tree
	}
}

func (img *ImageArchive) getLayer(treeName string) (*filetree.FileTree, bool) {
	if tr, exists := img.layerMap[treeName]; exists {
		return tr, true
	}

	norm := normalizeLayerPath(treeName)
	if tr, exists := img.layerMap[norm]; exists {
		return tr, true
	}

	id := extractLayerID(treeName)
	if id != "" && id != "." {
		candidates := []string{
			"blobs/sha256/" + id,
			id + "/layer.tar",
			id,
			"sha256:" + id,
			id + ".tar",
			id + ".tar.gz",
			id + ".tgz",
			"blobs/" + id,
		}
		for _, cand := range candidates {
			if tr, exists := img.layerMap[cand]; exists {
				return tr, true
			}
		}

		for k, tr := range img.layerMap {
			if extractLayerID(k) == id {
				return tr, true
			}
		}
	}

	return nil, false
}

func (img *ImageArchive) selectManifest(id string) manifest {
	if len(img.manifests) == 0 {
		return img.manifest
	}
	if len(img.manifests) == 1 || id == "" {
		return img.manifests[0]
	}
	for _, m := range img.manifests {
		for _, tag := range m.RepoTags {
			if tag == id || strings.HasPrefix(tag, id+":") || strings.TrimPrefix(tag, "docker.io/library/") == id {
				return m
			}
		}
	}
	for _, m := range img.manifests {
		if extractLayerID(m.ConfigPath) == extractLayerID(id) {
			return m
		}
	}
	return img.manifests[0]
}

func NewImageArchive(tarFile io.ReadCloser) (*ImageArchive, error) {
	img := &ImageArchive{
		layerMap: make(map[string]*filetree.FileTree),
	}

	tarReader := tar.NewReader(tarFile)

	// store discovered json files in a map so we can read the image in one pass
	jsonFiles := make(map[string][]byte)
	layerLinks := make(map[string]string)

	var currentLayer uint
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return img, err
		}

		name := path.Clean(strings.TrimPrefix(header.Name, "/"))
		name = strings.TrimPrefix(name, "./")
		if name == "." || name == "" {
			continue
		}

		if header.Typeflag == tar.TypeLink || header.Typeflag == tar.TypeSymlink {
			target := path.Clean(strings.TrimPrefix(header.Linkname, "/"))
			target = strings.TrimPrefix(target, "./")
			if !path.IsAbs(header.Linkname) && !strings.HasPrefix(header.Linkname, "blobs/") {
				resolved := path.Clean(path.Join(path.Dir(name), header.Linkname))
				resolved = strings.TrimPrefix(resolved, "/")
				resolved = strings.TrimPrefix(resolved, "./")
				if resolved != "." && resolved != "" {
					target = resolved
				}
			}
			layerLinks[name] = target
			continue
		}

		// some layer tars can be relative layer symlinks to other layer tars
		if header.Typeflag == tar.TypeReg {
			// For the Docker image format, use file name conventions
			if strings.HasSuffix(name, ".tar") {
				currentLayer++
				layerReader := tar.NewReader(tarReader)
				tree, err := processLayerTar(name, layerReader)
				if err != nil {
					return img, err
				}

				// add the layer to the image
				img.addLayer(tree)
			} else if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, "tgz") {
				currentLayer++

				// Add gzip reader
				gz, err := gzip.NewReader(tarReader)
				if err != nil {
					return img, err
				}

				// Add tar reader
				layerReader := tar.NewReader(gz)

				// Process layer
				tree, err := processLayerTar(name, layerReader)
				if err != nil {
					return img, err
				}

				// add the layer to the image
				img.addLayer(tree)
			} else if strings.HasSuffix(name, ".json") || strings.HasPrefix(name, "sha256:") {
				fileBuffer, err := io.ReadAll(tarReader)
				if err != nil {
					return img, err
				}
				jsonFiles[name] = fileBuffer
			} else if strings.HasPrefix(name, "blobs/") {
				// For the OCI-compatible image format (used since Docker 25), use mime sniffing
				// but limit this to only the blobs/ (containing the config, and the layers)

				// The idea here is that we try various formats in turn, and those tries should
				// never consume more bytes than this buffer contains so we can start again.

				// 512 bytes ought to be enough (as that's the size of a TAR entry header),
				// but play it safe with 1024 bytes. This should also include very small layers.
				buffer := make([]byte, 1024)
				n, err := io.ReadFull(tarReader, buffer)
				if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
					return img, err
				}

				if n == 0 {
					tree := filetree.NewFileTree()
					tree.Name = name
					currentLayer++
					img.addLayer(tree)
					continue
				}

				originalReader := func() io.Reader {
					return io.MultiReader(bytes.NewReader(buffer[:n]), tarReader)
				}

				// Try reading a gzip/estargz compressed layer
				gzipReader, err := gzip.NewReader(originalReader())
				if err == nil {
					layerReader := tar.NewReader(gzipReader)
					tree, err := processLayerTar(name, layerReader)
					if err == nil {
						currentLayer++
						// add the layer to the image
						img.addLayer(tree)
						continue
					}
				}

				// Try reading a zstd compressed layer
				zstdReader, err := zstd.NewReader(originalReader())
				if err == nil {
					layerReader := tar.NewReader(zstdReader)
					tree, err := processLayerTar(name, layerReader)
					if err == nil {
						currentLayer++
						// add the layer to the image
						img.addLayer(tree)
						continue
					}
				}

				// Try reading a plain tar layer
				layerReader := tar.NewReader(originalReader())
				tree, err := processLayerTar(name, layerReader)
				if err == nil {
					currentLayer++
					// add the layer to the image
					img.addLayer(tree)
					continue
				}

				// Not a TAR/GZIP/ZSTD, might be a JSON file
				decoder := json.NewDecoder(bytes.NewReader(buffer[:n]))
				token, err := decoder.Token()
				if _, ok := token.(json.Delim); err == nil && ok {
					// Looks like a JSON object (or array)
					// XXX: should we add a header.Size check too?
					fileBuffer, err := io.ReadAll(originalReader())
					if err != nil {
						return img, err
					}
					jsonFiles[name] = fileBuffer
				}
				// Ignore every other unknown file type
			}
		}
	}

	// Resolve layer and config links (hardlinks/symlinks)
	for pass := 0; pass < 5; pass++ {
		progress := false
		for linkName, targetName := range layerLinks {
			if _, exists := img.layerMap[linkName]; !exists {
				if tree, found := img.getLayer(targetName); found {
					img.addLayerWithName(linkName, tree)
					progress = true
				}
			}
			if _, exists := jsonFiles[linkName]; !exists {
				if content, found := getConfig(jsonFiles, targetName); found {
					jsonFiles[linkName] = content
					progress = true
				}
			}
		}
		if !progress {
			break
		}
	}

	manifestContent, exists := jsonFiles["manifest.json"]
	if !exists {
		for k, v := range jsonFiles {
			if path.Base(k) == "manifest.json" {
				manifestContent = v
				exists = true
				break
			}
		}
	}

	if exists {
		img.manifest = newManifest(manifestContent)
		img.manifests = newManifests(manifestContent)
	} else {
		// manifest.json is not part of the OCI spec, docker includes it for compatibility
		// Provide compatibility by finding the config and using our layerMap
		var configPath string
		for path, content := range jsonFiles {
			if isConfig(content) {
				configPath = path
				break
			}
		}
		if len(configPath) == 0 {
			return img, fmt.Errorf("could not find image manifest")
		}

		var layerPaths []string
		for k := range img.layerMap {
			layerPaths = append(layerPaths, k)
		}
		img.manifest = manifest{
			ConfigPath:    configPath,
			LayerTarPaths: layerPaths,
		}
		img.manifests = []manifest{img.manifest}
	}

	configContent, exists := getConfig(jsonFiles, img.manifest.ConfigPath)
	if !exists {
		return img, fmt.Errorf("could not find image config")
	}

	img.config = newConfig(configContent)

	return img, nil
}

func processLayerTar(name string, reader *tar.Reader) (*filetree.FileTree, error) {
	tree := filetree.NewFileTree()
	tree.Name = name

	fileInfos, err := getFileList(reader)
	if err != nil {
		return nil, err
	}

	for _, element := range fileInfos {
		tree.FileSize += uint64(element.Size)

		_, _, err := tree.AddPath(element.Path, element)
		if err != nil {
			return nil, err
		}
	}

	return tree, nil
}

func getFileList(tarReader *tar.Reader) ([]filetree.FileInfo, error) {
	var files []filetree.FileInfo

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// always ensure relative path notations are not parsed as part of the filename
		name := path.Clean(header.Name)
		if name == "." {
			continue
		}

		switch header.Typeflag {
		case tar.TypeXGlobalHeader:
			return nil, fmt.Errorf("unexpected tar file: (XGlobalHeader): type=%v name=%s", header.Typeflag, name)
		case tar.TypeXHeader:
			return nil, fmt.Errorf("unexpected tar file (XHeader): type=%v name=%s", header.Typeflag, name)
		default:
			files = append(files, filetree.NewFileInfoFromTarHeader(tarReader, header, name))
		}
	}
	return files, nil
}

func (img *ImageArchive) ToImage(id string) (*image.Image, error) {
	if len(img.manifests) > 1 {
		img.manifest = img.selectManifest(id)
	}

	trees := make([]*filetree.FileTree, 0)

	// build the content tree
	for _, treeName := range img.manifest.LayerTarPaths {
		tr, exists := img.getLayer(treeName)
		if exists {
			if tr.Name != treeName {
				namedTree := *tr
				namedTree.Name = treeName
				tr = &namedTree
			}
			trees = append(trees, tr)
			continue
		}
		return nil, fmt.Errorf("could not find '%s' in parsed layers", treeName)
	}

	// build the layers array
	layers := make([]*image.Layer, 0)

	// note that the engineResolver config stores images in reverse chronological order, so iterate backwards through layers
	// as you iterate chronologically through history (ignoring history items that have no layer contents)
	// Note: history is not required metadata in a docker image!
	histIdx := 0
	for idx, tree := range trees {
		// ignore empty layers, we are only observing layers with content
		historyObj := historyEntry{
			CreatedBy: "(missing)",
		}
		for nextHistIdx := histIdx; nextHistIdx < len(img.config.History); nextHistIdx++ {
			if !img.config.History[nextHistIdx].EmptyLayer {
				histIdx = nextHistIdx
				break
			}
		}
		if histIdx < len(img.config.History) && !img.config.History[histIdx].EmptyLayer {
			historyObj = img.config.History[histIdx]
			histIdx++
		}

		historyObj.Size = tree.FileSize

		dockerLayer := layer{
			history: historyObj,
			index:   idx,
			tree:    tree,
		}
		layers = append(layers, dockerLayer.ToLayer())
	}

	return &image.Image{
		Request: id,
		Trees:   trees,
		Layers:  layers,
	}, nil
}

func layerMatches(name, target string) bool {
	if name == target {
		return true
	}
	normName := normalizeLayerPath(name)
	normTarget := normalizeLayerPath(target)
	if normName == normTarget {
		return true
	}
	idName := extractLayerID(normName)
	idTarget := extractLayerID(normTarget)
	return idName != "" && idName == idTarget
}

func getDecompressedReader(r io.Reader) (io.Reader, error) {
	buffer := make([]byte, 1024)
	n, err := io.ReadFull(r, buffer)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, err
	}
	prefix := buffer[:n]
	fullReader := io.MultiReader(bytes.NewReader(prefix), r)

	if len(prefix) >= 2 && prefix[0] == 0x1f && prefix[1] == 0x8b {
		return gzip.NewReader(fullReader)
	}

	if len(prefix) >= 4 && prefix[0] == 0x28 && prefix[1] == 0xb5 && prefix[2] == 0x2f && prefix[3] == 0xfd {
		return zstd.NewReader(fullReader)
	}

	return fullReader, nil
}

func ExtractFromImage(tarFile io.ReadCloser, l string, p string) error {
	tarReader := tar.NewReader(tarFile)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		name := path.Clean(strings.TrimPrefix(header.Name, "/"))
		name = strings.TrimPrefix(name, "./")

		if header.Typeflag == tar.TypeReg {
			if layerMatches(name, l) {
				reader, err := getDecompressedReader(tarReader)
				if err != nil {
					return err
				}
				err = extractInner(tar.NewReader(reader), p)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}

	return nil
}

func extractInner(reader *tar.Reader, p string) error {
	target := strings.TrimPrefix(p, "/")

	for {
		header, err := reader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeReg:
			if strings.HasPrefix(name, target) {
				err := os.MkdirAll(filepath.Dir(name), 0755)
				if err != nil {
					return err
				}

				out, err := os.Create(name)
				if err != nil {
					return err
				}

				_, err = io.Copy(out, reader)
				if err != nil {
					return err
				}
			}
		default:
			continue
		}
	}

	return nil
}
