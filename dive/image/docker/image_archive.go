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
	"github.com/wagoodman/dive/internal/log"
)

type ImageArchive struct {
	manifest manifest
	config   config
	layerMap map[string]*filetree.FileTree
}

func NewImageArchive(tarFile io.ReadCloser) (*ImageArchive, error) {
	img := &ImageArchive{
		layerMap: make(map[string]*filetree.FileTree),
	}

	tarReader := tar.NewReader(tarFile)

	// store discovered json files in a map so we can read the image in one pass
	jsonFiles := make(map[string][]byte)

	var currentLayer uint
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := header.Name

		// some layer tars can be relative layer symlinks to other layer tars
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeReg {
			// For the Docker image format, use file name conventions
			if strings.HasSuffix(name, ".tar") {
				currentLayer++
				layerReader := tar.NewReader(tarReader)
				tree, err := processLayerTar(name, layerReader)
				if err != nil {
					return img, err
				}

				// add the layer to the image
				img.layerMap[tree.Name] = tree
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
				img.layerMap[tree.Name] = tree
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
				if err != nil && err != io.ErrUnexpectedEOF {
					return img, err
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
						img.layerMap[tree.Name] = tree
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
						img.layerMap[tree.Name] = tree
						continue
					}
				}

				// Try reading a plain tar layer
				layerReader := tar.NewReader(originalReader())
				tree, err := processLayerTar(name, layerReader)
				if err == nil {
					currentLayer++
					// add the layer to the image
					img.layerMap[tree.Name] = tree
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

	manifestContent, exists := jsonFiles["manifest.json"]
	if exists {
		img.manifest = newManifest(manifestContent)
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
	}

	configContent, exists := jsonFiles[img.manifest.ConfigPath]
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
	trees := make([]*filetree.FileTree, 0)

	// build the content tree
	for _, treeName := range img.manifest.LayerTarPaths {
		tr, exists := img.layerMap[treeName]
		if exists {
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

func ExtractFromImage(tarFile io.ReadCloser, l string, p string) error {
	defer tarFile.Close()
	tarReader := tar.NewReader(tarFile)

	// sanity check
	log.Debugf("ExtractFromImage: searching for layer '%s'", l)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("error reading tar entry: %w", err)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeReg:
			switch name {
			// Docker format: layers stored as <layer-id>/layer.tar
			case l + "/layer.tar":
				log.Debugf("ExtractFromImage: found docker format layer '%s'", name)
				return extractFromTar(tar.NewReader(tarReader), strings.TrimPrefix(p, "/"))
			// OCI format: layers stored as blobs/<layer-id>
			case "blobs/" + l:
				log.Debugf("ExtractFromImage: found OCI format layer '%s'", name)
				return extractCompressedLayer(tarReader, strings.TrimPrefix(p, "/"))
			default:
				continue
			}
		case tar.TypeSymlink:
			log.Debugf("ExtractFromImage: found symlink '%s'", name)
			continue
		default:
			continue
		}
	}

	return fmt.Errorf("layer '%s' not found in image archive", l)
}

func extractCompressedLayer(reader io.Reader, target string) error {
	if target == "" {
		target = "."
	}

	// Read a buffer to figure out the compression type
	buf := make([]byte, 1024)
	n, err := io.ReadFull(reader, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("failed to read layer header: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("empty layer entry")
	}

	layerReader := io.MultiReader(bytes.NewReader(buf[:n]), reader)

	// Check magic bytes to determine compression type
	if n >= 2 && buf[0] == 0x1f && buf[1] == 0x8b {
		// gzip magic bytes
		gzReader, err := gzip.NewReader(layerReader)
		if err != nil {
			return fmt.Errorf("failed to decompress gzip layer: %w", err)
		}
		defer gzReader.Close()
		return extractFromTar(tar.NewReader(gzReader), target)
	}

	if n >= 4 && buf[0] == 0x28 && buf[1] == 0xb5 && buf[2] == 0x2f && buf[3] == 0xfd {
		// zstd magic bytes (0xFD2FB528 little-endian)
		zstdReader, err := zstd.NewReader(layerReader)
		if err != nil {
			return fmt.Errorf("failed to decompress zstd layer: %w", err)
		}
		defer zstdReader.Close()
		return extractFromTar(tar.NewReader(zstdReader), target)
	}

	// No compression — plain tar archive
	return extractFromTar(tar.NewReader(layerReader), target)
}

func extractFromTar(tarReader *tar.Reader, target string) error {
	if target == "" {
		target = "."
	}

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("error reading layer tar entry: %w", err)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeReg:
			if strings.HasPrefix(name, target) {
				outPath := filepath.Clean(name)
				if filepath.IsAbs(outPath) {
					outPath = outPath[1:]
				}

				if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(outPath), err)
				}

				out, err := os.Create(outPath)
				if err != nil {
					return fmt.Errorf("failed to create file %s: %w", outPath, err)
				}

				if _, err := io.Copy(out, tarReader); err != nil {
					out.Close()
					return fmt.Errorf("failed to write file %s: %w", outPath, err)
				}
				out.Close()
			}
		default:
			continue
		}
	}

	return nil
}
