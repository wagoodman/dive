package handler

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/wagoodman/go-progress"

	"github.com/anchore/stereoscope/pkg/image/docker"
)

var _ dockerPullStatus = (*mockDockerPullStatus)(nil)

type mockDockerPullStatus struct {
	complete bool
	layers   []docker.LayerID
	current  map[docker.LayerID]docker.LayerState
}

func (m mockDockerPullStatus) Complete() bool {
	return m.complete
}

func (m mockDockerPullStatus) Layers() []docker.LayerID {
	return m.layers
}

func (m mockDockerPullStatus) Current(id docker.LayerID) docker.LayerState {
	return m.current[id]
}

func Test_dockerPullStatusFormatter_Render(t *testing.T) {

	tests := []struct {
		name   string
		status dockerPullStatus
	}{
		{
			name: "pulling",
			status: func() dockerPullStatus {
				complete := progress.NewManual(10)
				complete.Set(10)
				complete.SetCompleted()

				quarter := progress.NewManual(10)
				quarter.Set(2)

				half := progress.NewManual(10)
				half.Set(6)

				empty := progress.NewManual(10)

				return mockDockerPullStatus{
					complete: false,
					layers: []docker.LayerID{
						"sha256:1",
						"sha256:2",
						"sha256:3",
					},
					current: map[docker.LayerID]docker.LayerState{
						"sha256:1": {
							Phase:            docker.ExtractingPhase,
							PhaseProgress:    half,
							DownloadProgress: complete,
						},
						"sha256:2": {
							Phase:            docker.DownloadingPhase,
							PhaseProgress:    quarter,
							DownloadProgress: quarter,
						},
						"sha256:3": {
							Phase:            docker.WaitingPhase,
							PhaseProgress:    empty,
							DownloadProgress: empty,
						},
					},
				}
			}(),
		},
		{
			name: "download complete",
			status: func() dockerPullStatus {
				complete := progress.NewManual(10)
				complete.Set(10)
				complete.SetCompleted()

				quarter := progress.NewManual(10)
				quarter.Set(2)

				half := progress.NewManual(10)
				half.Set(6)

				empty := progress.NewManual(10)

				return mockDockerPullStatus{
					complete: false,
					layers: []docker.LayerID{
						"sha256:1",
						"sha256:2",
						"sha256:3",
					},
					current: map[docker.LayerID]docker.LayerState{
						"sha256:1": {
							Phase:            docker.ExtractingPhase,
							PhaseProgress:    complete,
							DownloadProgress: complete,
						},
						"sha256:2": {
							Phase:            docker.ExtractingPhase,
							PhaseProgress:    quarter,
							DownloadProgress: complete,
						},
						"sha256:3": {
							Phase:            docker.ExtractingPhase,
							PhaseProgress:    empty,
							DownloadProgress: complete,
						},
					},
				}
			}(),
		},
		{
			name: "complete",
			status: func() dockerPullStatus {
				complete := progress.NewManual(10)
				complete.Set(10)
				complete.SetCompleted()

				return mockDockerPullStatus{
					complete: true,
					layers: []docker.LayerID{
						"sha256:1",
						"sha256:2",
						"sha256:3",
					},
					current: map[docker.LayerID]docker.LayerState{
						"sha256:1": {
							Phase:            docker.PullCompletePhase,
							PhaseProgress:    complete,
							DownloadProgress: complete,
						},
						"sha256:2": {
							Phase:            docker.PullCompletePhase,
							PhaseProgress:    complete,
							DownloadProgress: complete,
						},
						"sha256:3": {
							Phase:            docker.PullCompletePhase,
							PhaseProgress:    complete,
							DownloadProgress: complete,
						},
					},
				}
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newDockerPullStatusFormatter()
			snaps.MatchSnapshot(t, f.Render(tt.status))
		})
	}
}
