package image

import (
	"fmt"
	"github.com/buildpacks/lifecycle"
	"github.com/buildpacks/pack"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/dive/filetree"
)


func (img *Image) CNBAnalyze(layerAnalysis *AnalysisResult) (*AnalysisResult, error) {
	client, err := pack.NewClient()
	if err != nil {
		return nil, fmt.Errorf("unable to create pack client: %s", err)
	}

	logrus.Debugf("Inspecting image: %s", img.Name)
	imageInfo, err := client.InspectImage(img.Name, true)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve %s image info: %s", img.Name, err)
	}
	empty := lifecycle.RunImageMetadata{}
	if imageInfo.Base == empty {
		return nil, fmt.Errorf("missing metadata")
	}
	logrus.Debugf("image info %#v", imageInfo)

	newLayers := []*Layer{}
	newRefTree := []*filetree.FileTree{}

	if len(layerAnalysis.Layers) != len(layerAnalysis.RefTrees) {
		return nil, fmt.Errorf("mismatched lengths %s vs %s", len(layerAnalysis.Layers), len(layerAnalysis.RefTrees))
	}

	var curLayer *Layer = nil
	var curRefTree *filetree.FileTree = nil
	var isStack = true
	for layerIdx, layer := range layerAnalysis.Layers {
		rTree := layerAnalysis.RefTrees[layerIdx]
		if curLayer == nil {
			curLayer = layer
			curRefTree = rTree
			continue
		}
		if isStack { // in stack still
			curLayer.Size += layer.Size
			_, err = curRefTree.Stack(rTree)
			if err != nil {
				return nil, fmt.Errorf("error to stacking trees")
			}
		}
		if layer.Digest == imageInfo.Base.TopLayer { // end of stack
			newLayers = append(newLayers, curLayer)
			newRefTree = append(newRefTree, curRefTree)
			isStack = false
		}
		if !isStack {
			newLayers = append(newLayers, layer)
			newRefTree = append(newRefTree, rTree)
		}
	}

	layerAnalysis.RefTrees = newRefTree
	layerAnalysis.Layers = newLayers
	layerAnalysis.BOMMapping = buildBOMMapping(layerAnalysis.Layers, imageInfo)

	return layerAnalysis, nil
}

func buildBOMMapping(layers []*Layer, labelMetadata *pack.ImageInfo) map[string]lifecycle.BOMEntry {
	result := make(map[string]lifecycle.BOMEntry, 0)
	for layerIndex, layer := range layers {
		result[layer.Digest] = lifecycle.BOMEntry{
			Require:   lifecycle.Require{},
			Buildpack: lifecycle.Buildpack{
				ID: fmt.Sprintf("buildpack metadata for layer: %d", layerIndex),
				Version: "1.2.3",
			},
		}
	}

	return result
}
