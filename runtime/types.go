package runtime

import (
	"github.com/spf13/viper"
)

type Options struct {
	Ci         bool
	ImageId    string
	ExportFile string
	CiConfig   *viper.Viper
	BuildArgs  []string
}

type export struct {
	Layer []exportLayer `json:"layer"`
	Image exportImage   `json:"image"`
}

type exportLayer struct {
	Index     int    `json:"index"`
	DigestID  string `json:"digestId"`
	SizeBytes uint64 `json:"sizeBytes"`
	Command   string `json:"command"`
}

type exportImage struct {
	SizeBytes        uint64             `json:"sizeBytes"`
	InefficientBytes uint64             `json:"inefficientBytes"`
	EfficiencyScore  float64            `json:"efficiencyScore"`
	InefficientFiles []inefficientFiles `json:"inefficientFiles"`
}

type inefficientFiles struct {
	Count     int    `json:"count"`
	SizeBytes uint64 `json:"sizeBytes"`
	File      string `json:"file"`
}
