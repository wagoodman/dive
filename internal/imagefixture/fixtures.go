package imagefixture

import (
	"archive/tar"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// diskFixtures are prebuilt image tarballs checked into the repo, relative to the repo root.
var diskFixtures = map[string]string{
	// legacy `docker save` of .data/Dockerfile.test-image
	"test-docker-image": ".data/test-docker-image.tar",
	// docker 25+ (OCI layout) `docker save` of the same image
	"test-oci-docker-image": ".data/test-oci-docker-image.tar",
}

func dir(name string) Entry { return Entry{Name: name, Type: tar.TypeDir} }

func file(name, content string) Entry { return Entry{Name: name, Content: content} }

func whiteout(name string) Entry { return Entry{Name: name} }

func run(cmd string) History { return History{CreatedBy: "/bin/sh -c " + cmd} }

func nop(cmd string) History { return History{CreatedBy: "/bin/sh -c #(nop) " + cmd, EmptyLayer: true} }

// baseLayers is a small three layer image shared by the format fixtures, so their goldens are directly comparable.
func baseLayers() ([][]Entry, []History) {
	layers := [][]Entry{
		{
			dir("bin/"),
			file("bin/sh", strings.Repeat("sh", 1024)),
			dir("etc/"),
			file("etc/os-release", "ID=fixture\n"),
			dir("usr/"),
			dir("usr/lib/"),
			file("usr/lib/libc.so", strings.Repeat("c", 3000)),
		},
		{
			dir("app/"),
			file("app/main.txt", "hello\n"),
			file("app/config.yaml", "key: value\n"),
		},
		{
			dir("app/"),
			file("app/main.txt", "hello, world\n"),
			dir("etc/"),
			whiteout("etc/.wh.os-release"),
		},
	}
	history := []History{
		{CreatedBy: "/bin/sh -c #(nop) ADD file:0123456789abcdef in / "},
		nop(`CMD ["/bin/sh"]`),
		{CreatedBy: "COPY app /app # buildkit"},
		run("echo 'hello, world' > /app/main.txt && rm /etc/os-release"),
	}
	return layers, history
}

func withFormat(f Format, c Compression) Spec {
	layers, history := baseLayers()
	return Spec{Format: f, Compression: c, Layers: layers, History: history}
}

// Fixtures returns the handwritten fixture specs by name.
func Fixtures() map[string]Spec {
	return map[string]Spec{
		"legacy-save":   withFormat(LegacySave, Uncompressed),
		"oci-save":      withFormat(OCISave, Uncompressed),
		"oci-save-gzip": withFormat(OCISave, Gzip),
		"oci-save-zstd": withFormat(OCISave, Zstd),
		"oci-archive-gzip": {
			// single layer: the old parser orders layers of a manifest-less archive by map iteration
			Format:      OCIArchive,
			Compression: Gzip,
			Layers: [][]Entry{{
				dir("bin/"),
				file("bin/busybox", strings.Repeat("b", 2048)),
				Entry{Name: "bin/ls", Type: tar.TypeSymlink, Linkname: "busybox"},
			}},
			History: []History{{CreatedBy: "/bin/sh -c #(nop) ADD file:fedcba9876543210 in / "}},
		},
		"opaque-whiteout-dir": {
			Format: OCISave,
			Layers: [][]Entry{
				{
					dir("etc/"),
					dir("etc/app/"),
					file("etc/app/a.conf", strings.Repeat("a", 100)),
					file("etc/app/b.conf", strings.Repeat("b", 200)),
					dir("etc/app/sub/"),
					file("etc/app/sub/c.conf", strings.Repeat("c", 300)),
					file("etc/keep.conf", "keep\n"),
				},
				{
					dir("etc/"),
					dir("etc/app/"),
					whiteout("etc/app/.wh..wh..opq"),
					file("etc/app/new.conf", strings.Repeat("n", 50)),
				},
			},
			History: []History{run("make etc"), run("rm -rf /etc/app && mkdir /etc/app && touch /etc/app/new.conf")},
		},
		"file-and-dir-whiteouts": {
			Format: OCISave,
			Layers: [][]Entry{
				{
					dir("data/"),
					file("data/keep.txt", "keep\n"),
					file("data/remove.txt", strings.Repeat("r", 400)),
					dir("data/olddir/"),
					file("data/olddir/x", strings.Repeat("x", 500)),
					dir("data/olddir/nested/"),
					file("data/olddir/nested/y", strings.Repeat("y", 600)),
				},
				{
					dir("data/"),
					whiteout("data/.wh.remove.txt"),
					whiteout("data/.wh.olddir"),
				},
				{
					dir("data/"),
					dir("data/olddir/"),
					file("data/olddir/z", strings.Repeat("z", 70)),
				},
			},
			History: []History{run("make data"), run("rm -rf /data/remove.txt /data/olddir"), run("mkdir /data/olddir && touch /data/olddir/z")},
		},
		"whiteout-missing-target": {
			// a whiteout for a path no lower layer has (seen in real images after squashing)
			Format: OCISave,
			Layers: [][]Entry{
				{dir("opt/"), file("opt/a", strings.Repeat("a", 100))},
				{
					dir("opt/"),
					whiteout("opt/.wh.never-existed"),
					file("opt/a", strings.Repeat("A", 150)),
					file("opt/b", strings.Repeat("b", 50)),
				},
			},
			History: []History{run("touch /opt/a"), run("rm -f /opt/never-existed && touch /opt/a /opt/b")},
		},
		"hardlink-and-symlink": {
			Format: OCISave,
			Layers: [][]Entry{
				{
					dir("usr/"),
					dir("usr/bin/"),
					file("usr/bin/tool", strings.Repeat("t", 1000)),
					{Name: "usr/bin/tool-hard", Type: tar.TypeLink, Linkname: "usr/bin/tool"},
					{Name: "usr/bin/tool-rel", Type: tar.TypeSymlink, Linkname: "tool"},
					{Name: "usr/bin/tool-abs", Type: tar.TypeSymlink, Linkname: "/usr/bin/tool"},
					{Name: "usr/bin/dangling", Type: tar.TypeSymlink, Linkname: "/nowhere"},
					dir("usr/lib/"),
					file("usr/lib/libx.so", strings.Repeat("l", 400)),
					{Name: "lib", Type: tar.TypeSymlink, Linkname: "usr/lib"},
				},
				{
					dir("usr/"),
					dir("usr/bin/"),
					// replace a symlink with a regular file, and hardlink to a file from the lower layer
					file("usr/bin/tool-rel", "now a file\n"),
					{Name: "usr/bin/libx-hard", Type: tar.TypeLink, Linkname: "usr/lib/libx.so"},
				},
			},
			History: []History{run("install tool"), run("relink tool")},
		},
		"empty-layer-history": {
			Format: OCISave,
			Layers: [][]Entry{
				{dir("a/"), file("a/one", "1\n")},
				{dir("a/"), file("a/two", "22\n")},
				{dir("a/"), file("a/three", "333\n")},
			},
			History: []History{
				nop("ARG VERSION=1"),
				{CreatedBy: "/bin/sh -c #(nop) ADD file:aaaa in / "},
				nop("ENV A=1"),
				nop("LABEL x=y"),
				run("echo 22 > /a/two"),
				nop("WORKDIR /a"),
				{CreatedBy: "RUN |1 VERSION=1 /bin/sh -c echo 333 > /a/three # buildkit"},
				nop(`ENTRYPOINT ["/bin/sh"]`),
			},
		},
		"no-history": {
			Format: OCISave,
			Layers: [][]Entry{
				{dir("a/"), file("a/one", "1\n")},
				{dir("a/"), file("a/two", "22\n")},
			},
		},
		"history-fewer-than-layers": {
			Format: OCISave,
			Layers: [][]Entry{
				{dir("a/"), file("a/one", "1\n")},
				{dir("a/"), file("a/two", "22\n")},
				{dir("a/"), file("a/three", "333\n")},
			},
			History: []History{nop("ENV A=1"), run("touch /a/one")},
		},
		"mode-uid-only-change": {
			Format: OCISave,
			Layers: [][]Entry{
				{
					dir("app/"),
					file("app/run.sh", "#!/bin/sh\necho hi\n"),
					file("app/owned", "owned\n"),
					file("app/same", "same\n"),
				},
				{
					dir("app/"),
					// same content, only metadata changes
					{Name: "app/run.sh", Mode: 0o755, Content: "#!/bin/sh\necho hi\n"},
					{Name: "app/owned", UID: 1000, GID: 1000, Content: "owned\n"},
					// byte for byte identical copy
					file("app/same", "same\n"),
				},
			},
			History: []History{run("make app"), run("chmod +x /app/run.sh && chown 1000:1000 /app/owned")},
		},
		"duplicate-files-across-layers": {
			Format: OCISave,
			Layers: [][]Entry{
				{
					dir("var/"),
					dir("var/cache/"),
					file("var/cache/big.bin", strings.Repeat("1", 10000)),
					file("var/cache/medium.bin", strings.Repeat("m", 3000)),
					file("var/cache/tie-a", strings.Repeat("a", 500)),
					file("var/cache/tie-b", strings.Repeat("b", 500)),
				},
				{
					dir("var/"),
					dir("var/cache/"),
					file("var/cache/big.bin", strings.Repeat("2", 12000)),
					file("var/cache/medium.bin", strings.Repeat("M", 3000)),
					file("var/cache/tie-a", strings.Repeat("A", 500)),
					file("var/cache/tie-b", strings.Repeat("B", 500)),
				},
				{
					dir("var/"),
					dir("var/cache/"),
					file("var/cache/big.bin", strings.Repeat("3", 8000)),
				},
				{
					dir("var/"),
					dir("var/cache/"),
					whiteout("var/cache/.wh.big.bin"),
				},
			},
			History: []History{run("fill cache"), run("refill cache"), run("refill big"), run("rm /var/cache/big.bin")},
		},
		"implicit-parent-dirs": {
			// no directory entries at all, parents only exist implicitly
			Format: OCISave,
			Layers: [][]Entry{
				{file("a/b/c/deep.txt", "deep\n"), file("top.txt", "top\n")},
				{file("a/b/other.txt", "other\n")},
			},
			History: []History{run("touch deep"), run("touch other")},
		},
	}
}

// Names returns every fixture name (handwritten and on-disk), sorted.
func Names() []string {
	var names []string
	for name := range Fixtures() {
		names = append(names, name)
	}
	for name := range diskFixtures {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Path returns the path of the named fixture's image tarball, writing it to a temp dir first if it is handwritten.
func Path(t testing.TB, name string) string {
	t.Helper()
	if rel, ok := diskFixtures[name]; ok {
		return RepoPath(rel)
	}
	spec, ok := Fixtures()[name]
	if !ok {
		t.Fatalf("unknown image fixture %q", name)
	}
	path := filepath.Join(t.TempDir(), name+".tar")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := spec.Write(f); err != nil {
		t.Fatalf("writing fixture %q: %v", name, err)
	}
	return path
}

// GoldenPath returns where the golden of the given kind ("export" or "ci") for a fixture lives:
// internal/imagefixture/testdata/golden/<name>.<kind>.json
func GoldenPath(name, kind string) string {
	return filepath.Join(packageDir(), "testdata", "golden", name+"."+kind+".json")
}

// UpdateGoldens reports whether golden files should be rewritten instead of compared (UPDATE_GOLDEN=true).
func UpdateGoldens() bool {
	return os.Getenv("UPDATE_GOLDEN") == "true"
}

func packageDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

// RepoPath resolves a slash separated path relative to the repo root.
func RepoPath(rel string) string {
	return filepath.Join(packageDir(), "..", "..", filepath.FromSlash(rel))
}

// AssertGolden compares got against the golden file at path, or rewrites the file when UpdateGoldens is set.
func AssertGolden(t testing.TB, path string, got []byte) {
	t.Helper()
	if UpdateGoldens() {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden (set UPDATE_GOLDEN=true to create it): %v", err)
	}
	require.Equal(t, string(want), string(got), "golden mismatch for %s (set UPDATE_GOLDEN=true to update)", path)
}
