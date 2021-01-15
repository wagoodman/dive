package format

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	//selectedLeftBracketStr = " "
	//selectedRightBracketStr = " "
	//selectedFillStr = " "
	//
	//leftBracketStr = "▏"
	//rightBracketStr = "▕"
	//fillStr = "─"

	//selectedLeftBracketStr = " "
	//selectedRightBracketStr = " "
	//selectedFillStr = "━"
	//
	//leftBracketStr = "▏"
	//rightBracketStr = "▕"
	//fillStr = "─"

	selectedLeftBracketStr  = "┃"
	selectedRightBracketStr = "┣"
	selectedEndBracketStr   = "┗"
	selectedFillStr         = "━"

	leftBracketStr  = "│"
	rightBracketStr = "├"
	endBracketStr   = "└"
	fillStr         = "─"

	selectStr = " ● "
	//selectStr = " "
)

type Formatter func(s string) string

func GenerateFormatter(fg, bg, flags string) Formatter {
	bold := strings.Contains(flags, "b")
	return func(s string) string {
		if bold {
		}
		return fmt.Sprintf("[%s:%s:%s]%s[-:-:-]", fg, bg, flags, s)
	}
}

func GenerateWholeLineFormatter(fg, bg, flags string) Formatter {
	bold := strings.Contains(flags, "b")
	return func(s string) string {
		if bold {
		}
		return fmt.Sprintf("[%s:%s:%s]%s", fg, bg, flags, s)
	}
}

var (
	// Bolds text
	Header                Formatter = GenerateFormatter("", "", "b")
	Normal                Formatter = GenerateFormatter("", "", "")
	Selected              Formatter = GenerateFormatter("", "", "rb")
	StatusSelected        Formatter = GenerateFormatter(colorTranslate(tcell.ColorWhite), colorTranslate(tcell.ColorDarkMagenta), "")
	StatusNormal          Formatter = GenerateFormatter("", "", "r")
	StatusControlSelected Formatter = GenerateFormatter(colorTranslate(tcell.ColorWhite), colorTranslate(tcell.ColorDarkMagenta), "b")
	StatusControlNormal   Formatter = GenerateFormatter("", "", "rb")
	CompareTop            Formatter = GenerateFormatter("", colorTranslate(tcell.ColorDarkMagenta), "")
	CompareBottom         Formatter = GenerateFormatter("", colorTranslate(tcell.ColorDarkGreen), "")
	FileTreeSelected      Formatter = func(s string) string { return boldReplace(GenerateWholeLineFormatter("", "", "rb")(s)) }

	// filediff types
	Added    Formatter = GenerateFormatter(colorTranslate(tcell.ColorGreen), "", "")
	Removed  Formatter = GenerateFormatter(colorTranslate(tcell.ColorRed), "", "")
	Modified Formatter = GenerateFormatter(colorTranslate(tcell.ColorYellow), "", "")

	// Styles these are needed to completely color a line
	SelectedStyle tcell.Style = tcell.Style{}.Bold(true).Reverse(true)
)

func PrintLine(screen tcell.Screen, text string, x, y, maxWidth, align int, style tcell.Style) (int, int) {
	totalWidth, totalHeight := screen.Size()
	if maxWidth <= 0 || len(text) == 0 || y < 0 || y >= totalHeight {
		return 0, 0
	}
	b, w := tview.PrintWithStyle(screen, text, x, y, maxWidth, align, style)
	maxWidth = intMin(totalWidth, maxWidth)
	for ; w < maxWidth; w++ {
		b++
		screen.SetContent(x+w, y, rune(' '), nil, style)
	}
	return b, w
}

func colorTranslate(c tcell.Color) string {
	return fmt.Sprintf("#%06x", c.Hex())
}

func boldReplace(s string) string {
	s = strings.ReplaceAll(s, leftBracketStr, selectedLeftBracketStr)
	s = strings.ReplaceAll(s, rightBracketStr, selectedRightBracketStr)
	s = strings.ReplaceAll(s, fillStr, selectedFillStr)
	s = strings.ReplaceAll(s, endBracketStr, selectedEndBracketStr)

	return s
}

// TODO factor me out into a utils package along with my usage in the componenets package
func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
