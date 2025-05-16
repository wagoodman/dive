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
	"github.com/anchore/stereoscope/pkg/image/docker"
)

var _ interface {
	progress.Stager
	progress.Progressable
} = (*dockerPullProgressAdapter)(nil)

type dockerPullStatus interface {
	Complete() bool
	Layers() []docker.LayerID
	Current(docker.LayerID) docker.LayerState
}

type dockerPullProgressAdapter struct {
	status    dockerPullStatus
	formatter dockerPullStatusFormatter
}

type dockerPullStatusFormatter struct {
	auxInfoStyle             lipgloss.Style
	dockerPullCompletedStyle lipgloss.Style
	dockerPullDownloadStyle  lipgloss.Style
	dockerPullExtractStyle   lipgloss.Style
	dockerPullStageChars     []string
	layerCaps                []string
}

func (m *Handler) handlePullDockerImage(e partybus.Event) []tea.Model {
	_, pullStatus, err := stereoscopeParsers.ParsePullDockerImage(e)
	if err != nil {
		log.WithFields("error", err).Debug("unable to parse event")
		return nil
	}

	tsk := m.newTaskProgress(
		taskprogress.Title{
			Default: "Pull image",
			Running: "Pulling image",
			Success: "Pulled image",
		},
		taskprogress.WithStagedProgressable(
			newDockerPullProgressAdapter(pullStatus),
		),
	)

	tsk.HintStyle = lipgloss.NewStyle()
	tsk.HintEndCaps = nil

	return []tea.Model{tsk}
}

func newDockerPullProgressAdapter(status dockerPullStatus) *dockerPullProgressAdapter {
	return &dockerPullProgressAdapter{
		status:    status,
		formatter: newDockerPullStatusFormatter(),
	}
}

func newDockerPullStatusFormatter() dockerPullStatusFormatter {
	return dockerPullStatusFormatter{
		auxInfoStyle:             lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")),
		dockerPullCompletedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#fcba03")),
		dockerPullDownloadStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")),
		dockerPullExtractStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
		dockerPullStageChars:     strings.Split("▁▃▄▅▆▇█", ""),
		layerCaps:                strings.Split("▕▏", ""),
	}
}

func (d dockerPullProgressAdapter) Size() int64 {
	return -1
}

func (d dockerPullProgressAdapter) Current() int64 {
	return 1
}

func (d dockerPullProgressAdapter) Error() error {
	if d.status.Complete() {
		return progress.ErrCompleted
	}
	// TODO: return intermediate error indications
	return nil
}

func (d dockerPullProgressAdapter) Stage() string {
	return d.formatter.Render(d.status)
}

// Render crafts the given docker image pull status summarized into a single line.
func (f dockerPullStatusFormatter) Render(pullStatus dockerPullStatus) string {
	var size, current uint64

	layers := pullStatus.Layers()
	status := make(map[docker.LayerID]docker.LayerState)
	completed := make([]string, len(layers))

	// fetch the current state
	for idx, layer := range layers {
		completed[idx] = " "
		status[layer] = pullStatus.Current(layer)
	}

	numCompleted := 0
	for idx, layer := range layers {
		prog := status[layer].PhaseProgress
		curN := prog.Current()
		curSize := prog.Size()

		if progress.IsCompleted(prog) {
			input := f.dockerPullStageChars[len(f.dockerPullStageChars)-1]
			completed[idx] = f.formatDockerPullPhase(status[layer].Phase, input)
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

			i := int(ratio * float64(len(f.dockerPullStageChars)-1))
			input := f.dockerPullStageChars[i]
			completed[idx] = f.formatDockerPullPhase(status[layer].Phase, input)
		}

		if progress.IsErrCompleted(status[layer].DownloadProgress.Error()) {
			numCompleted++
		}
	}

	for _, layer := range layers {
		prog := status[layer].DownloadProgress
		size += uint64(prog.Size())
		current += uint64(prog.Current())
	}

	var progStr, auxInfo string
	if len(layers) > 0 {
		render := strings.Join(completed, "")
		prefix := f.dockerPullCompletedStyle.Render(fmt.Sprintf("%d Layers", len(layers)))
		auxInfo = f.auxInfoStyle.Render(fmt.Sprintf("[%s / %s]", humanize.Bytes(current), humanize.Bytes(size)))
		if len(layers) == numCompleted {
			auxInfo = f.auxInfoStyle.Render(fmt.Sprintf("[%s] Extracting...", humanize.Bytes(size)))
		}

		progStr = fmt.Sprintf("%s%s%s%s", prefix, f.layerCap(false), render, f.layerCap(true))
	}

	return progStr + auxInfo
}

// formatDockerPullPhase returns a single character that represents the status of a layer pull.
func (f dockerPullStatusFormatter) formatDockerPullPhase(phase docker.PullPhase, inputStr string) string {
	switch phase {
	case docker.WaitingPhase:
		// ignore any progress related to waiting
		return " "
	case docker.PullingFsPhase, docker.DownloadingPhase:
		return f.dockerPullDownloadStyle.Render(inputStr)
	case docker.DownloadCompletePhase:
		return f.dockerPullDownloadStyle.Render(f.dockerPullStageChars[len(f.dockerPullStageChars)-1])
	case docker.ExtractingPhase:
		return f.dockerPullExtractStyle.Render(inputStr)
	case docker.VerifyingChecksumPhase, docker.PullCompletePhase:
		return f.dockerPullCompletedStyle.Render(inputStr)
	case docker.AlreadyExistsPhase:
		return f.dockerPullCompletedStyle.Render(f.dockerPullStageChars[len(f.dockerPullStageChars)-1])
	default:
		return inputStr
	}
}

func (f dockerPullStatusFormatter) layerCap(end bool) string {
	l := len(f.layerCaps)
	if l == 0 {
		return ""
	}
	if end {
		return f.layerCaps[l-1]
	}
	return f.layerCaps[0]
}
