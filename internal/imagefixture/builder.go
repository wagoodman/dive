// Package imagefixture writes small, deterministic container image tarballs from handwritten layer specs, so tests
// can exercise image parsing and analysis without a docker daemon.
package imagefixture

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/klauspost/compress/zstd"
)

// Format is the on-disk layout of the image tarball.
type Format int

const (
	// LegacySave is the `docker save` layout before docker 25: manifest.json, <id>/layer.tar, <hex>.json config.
	LegacySave Format = iota
	// OCISave is the `docker save` layout from docker 25 on: an OCI layout (index.json, oci-layout, blobs/sha256/...)
	// plus docker's manifest.json and repositories files.
	OCISave
	// OCIArchive is a plain OCI layout tarball (no manifest.json), as written by `buildx --output type=oci`.
	OCIArchive
)

// Compression applies to layer blobs of the OCI formats. LegacySave layers are always uncompressed.
type Compression int

const (
	Uncompressed Compression = iota
	Gzip
	Zstd
)

// Entry is a single tar entry within a layer.
type Entry struct {
	Name     string
	Type     byte // tar typeflag, zero means tar.TypeReg
	Mode     int64
	UID, GID int
	Linkname string
	Content  string
}

// History is one image config history entry.
type History struct {
	CreatedBy  string
	EmptyLayer bool
}

// Spec describes a whole image.
type Spec struct {
	Format      Format
	Compression Compression
	Layers      [][]Entry
	// nil means the config has no history key at all
	History []History
}

// fixedTime is used for every mtime and timestamp so output is byte for byte stable.
var fixedTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

const repoTag = "dive-fixture:latest"

type blob struct {
	digest string // "sha256:<hex>"
	data   []byte
}

func newBlob(data []byte) blob {
	sum := sha256.Sum256(data)
	return blob{digest: "sha256:" + hex.EncodeToString(sum[:]), data: data}
}

func (b blob) hex() string { return b.digest[len("sha256:"):] }

type descriptor struct {
	MediaType   string            `json:"mediaType"`
	Digest      string            `json:"digest"`
	Size        int               `json:"size"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type dockerManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// Write writes the image tarball for the spec to w.
func (s Spec) Write(w io.Writer) error {
	var diffIDs []string
	var layerTars [][]byte
	for i, entries := range s.Layers {
		data, err := layerTar(entries)
		if err != nil {
			return fmt.Errorf("layer %d: %w", i, err)
		}
		layerTars = append(layerTars, data)
		diffIDs = append(diffIDs, newBlob(data).digest)
	}

	config, err := s.config(diffIDs)
	if err != nil {
		return err
	}

	out := tar.NewWriter(w)
	switch s.Format {
	case LegacySave:
		err = writeLegacy(out, config, layerTars, diffIDs)
	case OCISave, OCIArchive:
		err = s.writeOCI(out, config, layerTars)
	default:
		err = fmt.Errorf("unknown format %d", s.Format)
	}
	if err != nil {
		return err
	}
	return out.Close()
}

func (s Spec) config(diffIDs []string) (blob, error) {
	type historyJSON struct {
		Created    string `json:"created"`
		CreatedBy  string `json:"created_by,omitempty"`
		EmptyLayer bool   `json:"empty_layer,omitempty"`
	}
	type rootfs struct {
		Type    string   `json:"type"`
		DiffIDs []string `json:"diff_ids"`
	}
	cfg := struct {
		Architecture string         `json:"architecture"`
		OS           string         `json:"os"`
		Created      string         `json:"created"`
		Config       map[string]any `json:"config"`
		History      []historyJSON  `json:"history,omitempty"`
		RootFS       rootfs         `json:"rootfs"`
	}{
		Architecture: "amd64",
		OS:           "linux",
		Created:      fixedTime.Format(time.RFC3339),
		Config:       map[string]any{"Env": []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}},
		RootFS:       rootfs{Type: "layers", DiffIDs: diffIDs},
	}
	for _, h := range s.History {
		cfg.History = append(cfg.History, historyJSON{
			Created:    fixedTime.Format(time.RFC3339),
			CreatedBy:  h.CreatedBy,
			EmptyLayer: h.EmptyLayer,
		})
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return blob{}, err
	}
	return newBlob(data), nil
}

func writeLegacy(out *tar.Writer, config blob, layerTars [][]byte, diffIDs []string) error {
	files := map[string][]byte{}
	var layerPaths []string
	parent := ""
	for i, data := range layerTars {
		// ponytail: v1 layer ids are just a stable hash of parent + diff id, docker's real chain math isn't needed
		sum := sha256.Sum256([]byte(parent + " " + diffIDs[i]))
		id := hex.EncodeToString(sum[:])
		v1, err := json.Marshal(map[string]any{"id": id, "parent": parent, "created": fixedTime.Format(time.RFC3339)})
		if err != nil {
			return err
		}
		files[id+"/"] = nil
		files[id+"/VERSION"] = []byte("1.0")
		files[id+"/json"] = v1
		files[id+"/layer.tar"] = data
		layerPaths = append(layerPaths, id+"/layer.tar")
		parent = id
	}
	configName := config.hex() + ".json"
	files[configName] = config.data

	manifest, err := json.Marshal([]dockerManifest{{Config: configName, RepoTags: []string{repoTag}, Layers: layerPaths}})
	if err != nil {
		return err
	}
	files["manifest.json"] = manifest
	if parent != "" {
		files["repositories"] = []byte(fmt.Sprintf(`{"dive-fixture":{"latest":%q}}`, parent))
	}
	return writeSorted(out, files)
}

func (s Spec) writeOCI(out *tar.Writer, config blob, layerTars [][]byte) error {
	mediaType := "application/vnd.oci.image.layer.v1.tar"
	switch s.Compression {
	case Gzip:
		mediaType += "+gzip"
	case Zstd:
		mediaType += "+zstd"
	}

	files := map[string][]byte{"blobs/": nil, "blobs/sha256/": nil}
	add := func(b blob) string {
		name := "blobs/sha256/" + b.hex()
		files[name] = b.data
		return name
	}

	var layerDescs []descriptor
	var layerPaths []string
	for _, data := range layerTars {
		compressed, err := compress(data, s.Compression)
		if err != nil {
			return err
		}
		b := newBlob(compressed)
		layerPaths = append(layerPaths, add(b))
		layerDescs = append(layerDescs, descriptor{MediaType: mediaType, Digest: b.digest, Size: len(b.data)})
	}
	configPath := add(config)

	manifest, err := json.Marshal(struct {
		SchemaVersion int          `json:"schemaVersion"`
		MediaType     string       `json:"mediaType"`
		Config        descriptor   `json:"config"`
		Layers        []descriptor `json:"layers"`
	}{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config:        descriptor{MediaType: "application/vnd.oci.image.config.v1+json", Digest: config.digest, Size: len(config.data)},
		Layers:        layerDescs,
	})
	if err != nil {
		return err
	}
	manifestBlob := newBlob(manifest)
	add(manifestBlob)

	index, err := json.Marshal(map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.index.v1+json",
		"manifests": []descriptor{{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Digest:    manifestBlob.digest,
			Size:      len(manifestBlob.data),
			Annotations: map[string]string{
				"io.containerd.image.name":          "docker.io/library/" + repoTag,
				"org.opencontainers.image.ref.name": "latest",
			},
		}},
	})
	if err != nil {
		return err
	}
	files["index.json"] = index
	files["oci-layout"] = []byte(`{"imageLayoutVersion":"1.0.0"}`)

	if s.Format == OCISave {
		dm, err := json.Marshal([]dockerManifest{{Config: configPath, RepoTags: []string{repoTag}, Layers: layerPaths}})
		if err != nil {
			return err
		}
		files["manifest.json"] = dm
		files["repositories"] = []byte(fmt.Sprintf(`{"dive-fixture":{"latest":%q}}`, config.hex()))
	}
	return writeSorted(out, files)
}

// writeSorted writes files in name order; a nil value is written as a directory.
func writeSorted(out *tar.Writer, files map[string][]byte) error {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		data := files[name]
		hdr := &tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(data)), ModTime: fixedTime}
		if data == nil {
			hdr.Typeflag, hdr.Mode, hdr.Size = tar.TypeDir, 0o755, 0
		}
		if err := out.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := out.Write(data); err != nil {
			return err
		}
	}
	return nil
}

// layerTar writes the entries in the given order (order is meaningful within a layer, e.g. hardlink targets).
func layerTar(entries []Entry) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.Name,
			Typeflag: e.Type,
			Mode:     e.Mode,
			Uid:      e.UID,
			Gid:      e.GID,
			Linkname: e.Linkname,
			ModTime:  fixedTime,
			Format:   tar.FormatGNU,
		}
		if hdr.Typeflag == 0 {
			hdr.Typeflag = tar.TypeReg
		}
		if hdr.Typeflag == tar.TypeReg {
			hdr.Size = int64(len(e.Content))
		}
		if hdr.Mode == 0 {
			switch hdr.Typeflag {
			case tar.TypeDir:
				hdr.Mode = 0o755
			case tar.TypeSymlink:
				hdr.Mode = 0o777
			default:
				hdr.Mode = 0o644
			}
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, fmt.Errorf("entry %q: %w", e.Name, err)
		}
		if hdr.Size > 0 {
			if _, err := io.WriteString(tw, e.Content); err != nil {
				return nil, err
			}
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func compress(data []byte, c Compression) ([]byte, error) {
	var buf bytes.Buffer
	var w io.WriteCloser
	var err error
	switch c {
	case Uncompressed:
		return data, nil
	case Gzip:
		// zero header mtime and no name keep gzip output stable
		w = gzip.NewWriter(&buf)
	case Zstd:
		w, err = zstd.NewWriter(&buf, zstd.WithEncoderConcurrency(1))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown compression %d", c)
	}
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
