package docker

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanTarPath(t *testing.T) {
	cases := map[string]string{
		"etc/motd":                         "etc/motd",
		"/etc/motd":                        "etc/motd",
		"etc/motd/../../../home/x/.bashrc": "home/x/.bashrc",
		"./etc/motd":                       "etc/motd",
		"etc/motd/../motd":                 "etc/motd",
	}
	for in, want := range cases {
		if got := cleanTarPath(in); got != want {
			t.Errorf("cleanTarPath(%q)=%q want %q", in, got, want)
		}
	}
}

func TestSafeExtractPathRejectsEscape(t *testing.T) {
	root := t.TempDir()
	if _, err := safeExtractPath(root, "etc/motd"); err != nil {
		t.Fatalf("expected ok for nested path: %v", err)
	}
	if _, err := safeExtractPath(root, "../outside"); err == nil {
		t.Fatal("expected error for ../outside")
	}
	// cleaned slip that leaves root
	if _, err := safeExtractPath(root, "a/../../outside"); err == nil {
		t.Fatal("expected error for a/../../outside")
	}
}

func TestExtractInnerRejectsTarSlip(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "dive-slip-should-not-exist")
	_ = os.Remove(outside)

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	slipName := "etc/motd/../../../" + filepath.Base(outside)
	// Use forward-slash form as tar archives do
	slipName = "etc/motd/../../../" + filepath.ToSlash(filepath.Base(outside))
	content := []byte("pwned\n")
	hdr := &tar.Header{
		Name:     slipName,
		Mode:     0644,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	// Also include a legitimate sibling that should extract when requested
	legit := []byte("ok\n")
	if err := tw.WriteHeader(&tar.Header{Name: "etc/motd", Mode: 0644, Size: int64(len(legit)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(legit); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	err = extractInner(tar.NewReader(bytes.NewReader(buf.Bytes())), "etc/motd")
	// Slip entry cleans to basename outside root match → skipped; legit etc/motd written.
	// If an implementation tried to write the raw slip name, safeExtractPath would error.
	if err != nil {
		// Accept either skip-by-clean-match or explicit refuse; must not write outside.
		if !strings.Contains(err.Error(), "path escape") {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if _, err := os.Stat(outside); err == nil {
		t.Fatalf("tar-slip wrote outside extract root: %s", outside)
		_ = os.Remove(outside)
	}

	want := filepath.Join(root, "etc", "motd")
	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("expected legitimate extract at %s: %v", want, err)
	}
	if string(data) != "ok\n" {
		t.Fatalf("legit content=%q", data)
	}
}

func TestExtractInnerWritesMatchedFile(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	body := []byte("hello\n")
	if err := tw.WriteHeader(&tar.Header{Name: "usr/share/doc/hello.txt", Mode: 0644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	if err := extractInner(tar.NewReader(bytes.NewReader(buf.Bytes())), "usr/share/doc/hello.txt"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(root, "usr", "share", "doc", "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello\n" {
		t.Fatalf("got %q", got)
	}
}

// Ensure unused import does not linger when io is needed by package already.
