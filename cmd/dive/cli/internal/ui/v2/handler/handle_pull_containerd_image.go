package handler

import (
	"fmt"
	"github.com/wagoodman/dive/internal/log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"

	"github.com/anchore/bubbly/bubbles/taskprogress"
	stereoscopeParsers "github.com/anchore/stereoscope/pkg/event/parsers"
	"github.com/anchore/stereoscope/pkg/image/containerd"
)

var _ interface {
	progress.Stager
	progress.Progressable
} = (*containerdPullProgressAdapter)(nil)

type containerdPullStatus interface {
	Complete() bool
	Layers() []containerd.LayerID
	Current(containerd.LayerID) progress.Progressable
}

type containerdPullProgressAdapter struct {
	status    containerdPullStatus
	formatter containerdPullStatusFormatter
}

type containerdPullStatusFormatter struct {
	auxInfoStyle       lipgloss.Style
	pullCompletedStyle lipgloss.Style
	pullDownloadStyle  lipgloss.Style
	pullStageChars     []string
	layerCaps          []string
}

func (m *Handler) handlePullContainerdImage(e partybus.Event) []tea.Model {
	_, pullStatus, err := stereoscopeParsers.ParsePullContainerdImage(e)
	if err != nil {
		log.WithFields("error", err).Debug("unable to parse event")
		return nil
	}

	if pullStatus == nil {
		return nil
	}

	tsk := m.newTaskProgress(
		taskprogress.Title{
			Default: "Pull image",
			Running: "Pulling image",
			Success: "Pulled image",
		},
		taskprogress.WithStagedProgressable(
			newContainerdPullProgressAdapter(pullStatus),
		),
	)

	tsk.HintStyle = lipgloss.NewStyle()
	tsk.HintEndCaps = nil

	return []tea.Model{tsk}
}

func newContainerdPullProgressAdapter(status *containerd.PullStatus) *containerdPullProgressAdapter {
	return &containerdPullProgressAdapter{
		status:    status,
		formatter: newContainerdPullStatusFormatter(),
	}
}

func newContainerdPullStatusFormatter() containerdPullStatusFormatter {
	return containerdPullStatusFormatter{
		auxInfoStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")),
		pullCompletedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#fcba03")),
		pullDownloadStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")),
		pullStageChars:     strings.Split("▁▃▄▅▆▇█", ""),
		layerCaps:          strings.Split("▕▏", ""),
	}
}

func (d containerdPullProgressAdapter) Size() int64 {
	return -1
}

func (d containerdPullProgressAdapter) Current() int64 {
	return 1
}

func (d containerdPullProgressAdapter) Error() error {
	if d.status.Complete() {
		return progress.ErrCompleted
	}
	// TODO: return intermediate error indications
	return nil
}

func (d containerdPullProgressAdapter) Stage() string {
	return d.formatter.Render(d.status)
}

// Render crafts the given docker image pull status summarized into a single line.
func (f containerdPullStatusFormatter) Render(pullStatus containerdPullStatus) string {
	var size, current uint64

	layers := pullStatus.Layers()
	status := make(map[containerd.LayerID]progress.Progressable)
	completed := make([]string, len(layers))

	// fetch the current state
	for idx, layer := range layers {
		completed[idx] = " "
		status[layer] = pullStatus.Current(layer)
	}

	numCompleted := 0
	for idx, layer := range layers {
		prog := status[layer]
		curN := prog.Current()
		curSize := prog.Size()

		if progress.IsCompleted(prog) {
			input := f.pullStageChars[len(f.pullStageChars)-1]
			completed[idx] = f.formatPullPhase(prog.Error() != nil, input)
		} else if curN != 0 {
			var ratio float64
			switch {
			case curN == 0 || curSize < 0:
				ratio = 0
			case curN >= curSize:
				ratio = 1
			default:
				ratio = float64(curN) / float64(curSize)
			}

			i := int(ratio * float64(len(f.pullStageChars)-1))
			input := f.pullStageChars[i]
			completed[idx] = f.formatPullPhase(status[layer].Error() != nil, input)
		}

		if progress.IsErrCompleted(status[layer].Error()) {
			numCompleted++
		}
	}

	for _, layer := range layers {
		prog := status[layer]
		size += uint64(prog.Size())
		current += uint64(prog.Current())
	}

	var progStr, auxInfo string
	if len(layers) > 0 {
		render := strings.Join(completed, "")
		prefix := f.pullCompletedStyle.Render(fmt.Sprintf("%d Layers", len(layers)))
		auxInfo = f.auxInfoStyle.Render(fmt.Sprintf("[%s / %s]", humanize.Bytes(current), humanize.Bytes(size)))
		if len(layers) == numCompleted {
			auxInfo = f.auxInfoStyle.Render(fmt.Sprintf("[%s] Extracting...", humanize.Bytes(size)))
		}

		progStr = fmt.Sprintf("%s%s%s%s", prefix, f.layerCap(false), render, f.layerCap(true))
	}

	return progStr + auxInfo
}

// formatPullPhase returns a single character that represents the status of a layer pull.
func (f containerdPullStatusFormatter) formatPullPhase(completed bool, inputStr string) string {
	if completed {
		return f.pullCompletedStyle.Render(f.pullStageChars[len(f.pullStageChars)-1])
	}
	return f.pullDownloadStyle.Render(inputStr)
}

func (f containerdPullStatusFormatter) layerCap(end bool) string {
	l := len(f.layerCaps)
	if l == 0 {
		return ""
	}
	if end {
		return f.layerCaps[l-1]
	}
	return f.layerCaps[0]
}
