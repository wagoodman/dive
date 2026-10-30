package handler

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/wagoodman/go-progress"

	"github.com/anchore/stereoscope/pkg/image/containerd"
)

var _ containerdPullStatus = (*mockContainerdPullStatus)(nil)

type mockContainerdPullStatus struct {
	complete bool
	layers   []containerd.LayerID
	current  map[containerd.LayerID]progress.Progressable
}

func (m mockContainerdPullStatus) Complete() bool {
	return m.complete
}

func (m mockContainerdPullStatus) Layers() []containerd.LayerID {
	return m.layers
}

func (m mockContainerdPullStatus) Current(id containerd.LayerID) progress.Progressable {
	return m.current[id]
}

func Test_containerdPullStatusFormatter_Render(t *testing.T) {
	newProg := func(size, current int64, done bool) progress.Progressable {
		p := progress.NewManual(size)
		p.Set(current)
		if done {
			p.SetCompleted()
		}
		return p
	}

	tests := []struct {
		name   string
		status containerdPullStatus
	}{
		{
			name: "pulling",
			status: mockContainerdPullStatus{
				layers: []containerd.LayerID{"sha256:1", "sha256:2", "sha256:3"},
				current: map[containerd.LayerID]progress.Progressable{
					"sha256:1": newProg(10, 10, true),
					"sha256:2": newProg(10, 6, false),
					"sha256:3": newProg(10, 0, false),
				},
			},
		},
		{
			name: "complete",
			status: mockContainerdPullStatus{
				complete: true,
				layers:   []containerd.LayerID{"sha256:1", "sha256:2"},
				current: map[containerd.LayerID]progress.Progressable{
					"sha256:1": newProg(10, 10, true),
					"sha256:2": newProg(10, 10, true),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newContainerdPullStatusFormatter()
			snaps.MatchSnapshot(t, f.Render(tt.status))
		})
	}
}
