package view

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileTree_RequestedSizeUsesPaneWidth(t *testing.T) {
	v := &FileTree{requestedWidthRatio: 0.7}

	width := v.RequestedSize(200)
	require.NotNil(t, width)
	require.Equal(t, 140, *width)
}
