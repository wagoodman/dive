package format

import (
	"github.com/fatih/color"
)

var (
	Header                func(...interface{}) string
	Selected              func(...interface{}) string
	StatusSelected        func(...interface{}) string
	StatusNormal          func(...interface{}) string
	StatusControlSelected func(...interface{}) string
	StatusControlNormal   func(...interface{}) string
	CompareTop            func(...interface{}) string
	CompareBottom         func(...interface{}) string
)

func init() {
	Selected = color.New(color.ReverseVideo, color.Bold).SprintFunc()
	Header = color.New(color.Bold).SprintFunc()
	StatusSelected = color.New(color.BgMagenta, color.FgWhite).SprintFunc()
	StatusNormal = color.New(color.ReverseVideo).SprintFunc()
	StatusControlSelected = color.New(color.BgMagenta, color.FgWhite, color.Bold).SprintFunc()
	StatusControlNormal = color.New(color.ReverseVideo, color.Bold).SprintFunc()
	CompareTop = color.New(color.BgMagenta).SprintFunc()
	CompareBottom = color.New(color.BgGreen).SprintFunc()
}

func RenderHelpKey(control, title string, selected bool) string {
	if selected {
		return StatusSelected("▏") + StatusControlSelected(control) + StatusSelected(" "+title+" ")
	} else {
		return StatusNormal("▏") + StatusControlNormal(control) + StatusNormal(" "+title+" ")
	}
}
