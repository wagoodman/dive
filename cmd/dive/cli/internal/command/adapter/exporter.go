package adapter

import (
	"context"
	"fmt"
	"github.com/spf13/afero"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/export"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/internal/log"
	"os"
)

type Exporter interface {
	ExportTo(ctx context.Context, img *image.Analysis, path string) error
}

type jsonExporter struct {
	filesystem afero.Fs
}

func NewExporter(fs afero.Fs) Exporter {
	return &jsonExporter{
		filesystem: fs,
	}
}

func (e *jsonExporter) ExportTo(ctx context.Context, analysis *image.Analysis, path string) error {
	log.WithFields("path", path).Infof("exporting analysis")

	bytes, err := export.NewExport(analysis).Marshal()
	if err != nil {
		return fmt.Errorf("cannot marshal export payload: %w", err)
	}

	file, err := e.filesystem.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("cannot open export file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("cannot write to export file: %w", err)
	}
	return nil
}
