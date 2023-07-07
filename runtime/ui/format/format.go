package format

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/lunixbochs/vtclean"
)

const (
	// selectedLeftBracketStr = " "
	// selectedRightBracketStr = " "
	// selectedFillStr = " "
	//
	//leftBracketStr = "▏"
	//rightBracketStr = "▕"
	//fillStr = "─"

	// selectedLeftBracketStr = " "
	// selectedRightBracketStr = " "
	// selectedFillStr = "━"
	//
	//leftBracketStr = "▏"
	//rightBracketStr = "▕"
	//fillStr = "─"

	selectedLeftBracketStr  = "┃"
	selectedRightBracketStr = "┣"
	selectedFillStr         = "━"

	leftBracketStr  = "│"
	rightBracketStr = "├"
	fillStr         = "─"

	selectStr = " ● "
	// selectStr = " "
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

func RenderNoHeader(width int, selected bool) string {
	if selected {
		return strings.Repeat(selectedFillStr, width)
	}
	return strings.Repeat(fillStr, width)
}

func RenderHeader(title string, width int, selected bool) string {
	if selected {
		body := Header(fmt.Sprintf("%s%s ", selectStr, title))
		bodyLen := len(vtclean.Clean(body, false))
		repeatCount := width - bodyLen - 2
		if repeatCount < 0 {
			repeatCount = 0
		}
		return fmt.Sprintf("%s%s%s%s\n", selectedLeftBracketStr, body, selectedRightBracketStr, strings.Repeat(selectedFillStr, repeatCount))
		// return fmt.Sprintf("%s%s%s%s\n", Selected(selectedLeftBracketStr), body, Selected(selectedRightBracketStr), Selected(strings.Repeat(selectedFillStr, width-bodyLen-2)))
		// return fmt.Sprintf("%s%s%s%s\n", Selected(selectedLeftBracketStr), body, Selected(selectedRightBracketStr), strings.Repeat(selectedFillStr, width-bodyLen-2))
	}
	body := Header(fmt.Sprintf(" %s ", title))
	bodyLen := len(vtclean.Clean(body, false))
	repeatCount := width - bodyLen - 2
	if repeatCount < 0 {
		repeatCount = 0
	}
	return fmt.Sprintf("%s%s%s%s\n", leftBracketStr, body, rightBracketStr, strings.Repeat(fillStr, repeatCount))
}

func RenderHelpKey(control, title string, selected bool) string {
	if selected {
		return StatusSelected("▏") + StatusControlSelected(control) + StatusSelected(" "+title+" ")
	} else {
		return StatusNormal("▏") + StatusControlNormal(control) + StatusNormal(" "+title+" ")
	}
}
