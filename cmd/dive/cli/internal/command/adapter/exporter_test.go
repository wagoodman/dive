package adapter

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/wagoodman/dive/dive/image"
)

func TestJSONExporterTruncatesExistingFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	const path = "dive.json"

	if err := afero.WriteFile(fs, path, []byte(strings.Repeat("x", 4096)), 0o644); err != nil {
		t.Fatalf("failed to seed export file: %v", err)
	}

	analysis := &image.Analysis{Image: "example:latest"}
	if err := NewExporter(fs).ExportTo(context.Background(), analysis, path); err != nil {
		t.Fatalf("failed to export analysis: %v", err)
	}

	contents, err := afero.ReadFile(fs, path)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}
	if !json.Valid(contents) {
		t.Fatalf("export file contains trailing data: %q", contents)
	}
}
