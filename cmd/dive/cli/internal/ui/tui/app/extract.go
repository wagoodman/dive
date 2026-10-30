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

	"github.com/wagoodman/dive/internal/layertree"
)

// extract writes the regular files that layer idx shows at or under p (its squashed tree: the file as it is at that
// layer, whichever layer added it) into dst, each at its image path relative to dst (/etc/hosts lands at
// dst/etc/hosts). Nothing is written outside dst: entries whose tar name is absolute or climbs out with "..", and
// paths that would go through a symlink pointing out of dst, are refused. An existing path is never removed,
// overwritten or written through (a symlink included), it is skipped instead. Refused and skipped entries are
// reported, the rest are still written.
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
	var found bool
	err = layertree.Walk(img.Layers[idx].SquashedTree, nil, func(n *filenode.FileNode, _ bool) error {
		realPath := string(n.RealPath)
		if (realPath != target && !strings.HasPrefix(realPath, prefix)) || n.RealPath.IsWhiteout() {
			return nil
		}
		hdr, ok := layertree.Header(img, n.Reference)
		if !ok || hdr.Typeflag != tar.TypeReg {
			return nil
		}
		found = true
		if !filepath.IsLocal(filepath.FromSlash(hdr.Name)) {
			errs = append(errs, fmt.Errorf("refusing to extract %q: not a relative path", hdr.Name))
			return nil
		}
		if err := extractFile(img, root, *n.Reference, strings.TrimPrefix(realPath, "/")); err != nil {
			errs = append(errs, fmt.Errorf("unable to extract %q: %w", realPath, err))
		}
		return nil
	})
	if err == nil && !found {
		err = fmt.Errorf("no regular files to extract at %q", target)
	}
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
	// O_EXCL fails on anything already at the path, a symlink included (it is never followed), so nobody's file is
	// replaced
	out, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("%q exists, not overwritten", rel)
	}
	if err != nil {
		return err
	}
	_, err = io.Copy(out, contents)
	return errors.Join(err, out.Close())
}
