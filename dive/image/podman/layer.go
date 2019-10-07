package podman

import (
	"context"
	"fmt"
	podmanImage "github.com/containers/libpod/libpod/image"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"strings"
)

// Layer represents a Docker image layer and metadata
type layer struct {
	obj       *podmanImage.Image
	history   *podmanImage.History
	index     int
	tree      *filetree.FileTree
}

func (l *layer) getHistory() (*podmanImage.History, error) {
	if l.history != nil {
		return l.history, nil
	}
	history, err := l.obj.History(context.TODO())
	if err != nil {
		return nil, err
	}
	if len(history) > 0 {
		l.history = history[0]
		return history[0], nil
	}
	return nil, fmt.Errorf("could not find history")
}

func (l *layer) Size() uint64 {
	history, err := l.getHistory()
	if err != nil {
		// todo: what should be done here???
		panic(err)
	}
	return uint64(history.Size)
}

// ShortId returns the truncated id of the current layer.
func (l *layer) Command() string {
	history, err := l.getHistory()
	if err != nil {
		return "error: " + err.Error()
	}
	return strings.TrimPrefix(history.CreatedBy, "/bin/sh -c ")
}

// ShortId returns the truncated id of the current layer.
func (l *layer) ShortId() string {
	rangeBound := 15
	id := l.obj.ID()
	if length := len(id); length < 15 {
		rangeBound = length
	}
	id = id[0:rangeBound]

	return id
}

// String represents a layer in a columnar format.
func (l *layer) ToLayer() *image.Layer {
	return &image.Layer{
		Id:      l.obj.ID(),
		Index:   l.index,
		Command: l.Command(),
		Size:    l.Size(),
		Tree:    l.tree,
		Names:   l.obj.Names(),
		Digest:  l.obj.Digest().String(),
	}
}
