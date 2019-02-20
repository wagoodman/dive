package ui

import (
	"github.com/wagoodman/dive/filetree"
	"github.com/wagoodman/dive/image"
	"testing"
)

func TestFileTreeTest(t *testing.T) {
	result, err := image.TestLoadDockerImageTar("../.data/test-docker-image.tar")
	if err != nil {
		t.Fatalf("Test_Export: unable to fetch analysis: %v", err)
	}
	cache := filetree.NewFileTreeCache(result.RefTrees)
	cache.Build()



}
