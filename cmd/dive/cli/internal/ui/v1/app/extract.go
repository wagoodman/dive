package app

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/filetree/filenode"
	"github.com/anchore/stereoscope/pkg/image"

	"github.com/wagoodman/dive/dive"
)

// extract writes the regular files that layer idx has at or under p into dst, each at its image path relative to
// dst (/etc/hosts lands at dst/etc/hosts). Nothing is written outside dst: entries whose tar name is absolute or
// climbs out with "..", and paths that would go through a symlink pointing out of dst, are refused, and an
// existing symlink at a file's own path is never written through. Refused entries are reported, the rest are
// still written.
func extract(img *image.Image, idx int, p, dst string) error {
	if idx < 0 || idx >= len(img.Layers) {
		return fmt.Errorf("invalid layer index: %d", idx)
	}

	// os.Root refuses any path (including through symlinks) that resolves outside of dst
	root, err := os.OpenRoot(dst)
	if err != nil {
		return err
	}
	defer root.Close()

	target := path.Clean("/" + p)
	prefix := strings.TrimSuffix(target, "/") + "/"

	var errs []error
	err = dive.WalkTree(img.Layers[idx].Tree, func(n *filenode.FileNode, _ bool) error {
		realPath := string(n.RealPath)
		if (realPath != target && !strings.HasPrefix(realPath, prefix)) || n.Reference == nil {
			return nil
		}
		entry, err := img.FileCatalog.Get(*n.Reference)
		if err != nil {
			return nil
		}
		hdr := dive.TarHeader(entry.Metadata)
		if hdr == nil || hdr.Typeflag != tar.TypeReg {
			return nil
		}
		if !filepath.IsLocal(filepath.FromSlash(hdr.Name)) {
			errs = append(errs, fmt.Errorf("refusing to extract %q: not a relative path", hdr.Name))
			return nil
		}
		if err := extractFile(img, root, *n.Reference, strings.TrimPrefix(realPath, "/")); err != nil {
			errs = append(errs, fmt.Errorf("unable to extract %q: %w", realPath, err))
		}
		return nil
	})
	return errors.Join(append(errs, err)...)
}

func extractFile(img *image.Image, root *os.Root, ref file.Reference, rel string) error {
	contents, err := img.OpenReference(ref)
	if err != nil {
		return err
	}
	defer contents.Close()

	name := filepath.FromSlash(rel)
	if err := root.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}
	// never follow a symlink that is already at this path, even one that stays inside the root
	if fi, err := root.Lstat(name); err == nil && fi.Mode()&fs.ModeSymlink != 0 {
		return fmt.Errorf("%q is a symlink", rel)
	}
	out, err := root.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, contents)
	return errors.Join(err, out.Close())
}
