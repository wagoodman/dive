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

	//selectStr = " ● "
	//selectStr = " "
)

func SyncWithTermColors() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.PrimaryTextColor = tcell.ColorDefault
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault    // Main background color for primitives.
	tview.Styles.ContrastBackgroundColor = tcell.ColorDefault     // Background color for contrasting elements.
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorDefault // Background color for even more contrasting elements.
	tview.Styles.BorderColor = tcell.ColorDefault                 // Box borders.
	tview.Styles.TitleColor = tcell.ColorDefault                  // Box titles.
	tview.Styles.GraphicsColor = tcell.ColorDefault               // Graphics.
	tview.Styles.PrimaryTextColor = tcell.ColorDefault            // Primary text.
	tview.Styles.SecondaryTextColor = tcell.ColorDefault          // Secondary text (e.g. labels).
	tview.Styles.TertiaryTextColor = tcell.ColorDefault           // Tertiary text (e.g. subtitles, notes).
	tview.Styles.InverseTextColor = tcell.ColorDefault            // Text on primary-colored backgrounds.
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorDefault  // Secondary text on ContrastBackgroundColor-colored backgrounds.
}

type Formatter func(s string) string

func GenerateFormatter(fg, bg, flags string) Formatter {
	return func(s string) string {
		return fmt.Sprintf("[%s:%s:%s]%s[-:-:-]", fg, bg, flags, s)
	}
}

func GenerateWholeLineFormatter(fg, bg, flags string) Formatter {
	return func(s string) string {
		return fmt.Sprintf("[%s:%s:%s]%s", fg, bg, flags, s)
	}
}

var (
	// Bolds text
	Header                    Formatter = GenerateFormatter("", "", "b")
	Normal                    Formatter = GenerateFormatter("", "", "")
	None                      Formatter = func(s string) string { return s }
	Selected                  Formatter = GenerateFormatter("", "", "rb")
	StatusSelected            Formatter = GenerateFormatter("white", "purple", "")
	StatusNormal              Formatter = GenerateFormatter("", "", "r")
	StatusControlSelected     Formatter = GenerateFormatter("white", "purple", "")
	StatusControlSelectedBold Formatter = GenerateFormatter("white", "purple", "b")
	StatusControlNormal       Formatter = GenerateFormatter("", "", "r")
	StatusControlNormalBold   Formatter = GenerateFormatter("", "", "rb")
	CompareTop                Formatter = GenerateFormatter("", "purple", "")
	CompareBottom             Formatter = GenerateFormatter("", "green", "")

	// filediff types
	Added    Formatter = GenerateFormatter("green", "", "")
	Removed  Formatter = GenerateFormatter("red", "", "")
	Modified Formatter = GenerateFormatter("yellow", "", "")

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

func BoldReplace(s string) string {
	s = strings.ReplaceAll(s, leftBracketStr, selectedLeftBracketStr)
	s = strings.ReplaceAll(s, rightBracketStr, selectedRightBracketStr)
	s = strings.ReplaceAll(s, fillStr, selectedFillStr)
	s = strings.ReplaceAll(s, endBracketStr, selectedEndBracketStr)

	return s
}

// TODO factor this out into a utils package along with my usage in the componenets package
func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
