package format

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var colorAttrMap = map[string]color.Attribute{
	// Foreground basic
	"black":      color.FgBlack,
	"red":        color.FgRed,
	"green":      color.FgGreen,
	"yellow":     color.FgYellow,
	"blue":       color.FgBlue,
	"magenta":    color.FgMagenta,
	"cyan":       color.FgCyan,
	"white":      color.FgWhite,
	"fg-black":   color.FgBlack,
	"fg-red":     color.FgRed,
	"fg-green":   color.FgGreen,
	"fg-yellow":  color.FgYellow,
	"fg-blue":    color.FgBlue,
	"fg-magenta": color.FgMagenta,
	"fg-cyan":    color.FgCyan,
	"fg-white":   color.FgWhite,

	// Foreground bright / high intensity
	"bright-black":      color.FgHiBlack,
	"bright-red":        color.FgHiRed,
	"bright-green":      color.FgHiGreen,
	"bright-yellow":     color.FgHiYellow,
	"bright-blue":       color.FgHiBlue,
	"bright-magenta":    color.FgHiMagenta,
	"bright-cyan":       color.FgHiCyan,
	"bright-white":      color.FgHiWhite,
	"hi-black":          color.FgHiBlack,
	"hi-red":            color.FgHiRed,
	"hi-green":          color.FgHiGreen,
	"hi-yellow":         color.FgHiYellow,
	"hi-blue":           color.FgHiBlue,
	"hi-magenta":        color.FgHiMagenta,
	"hi-cyan":           color.FgHiCyan,
	"hi-white":          color.FgHiWhite,
	"fg-bright-black":   color.FgHiBlack,
	"fg-bright-red":     color.FgHiRed,
	"fg-bright-green":   color.FgHiGreen,
	"fg-bright-yellow":  color.FgHiYellow,
	"fg-bright-blue":    color.FgHiBlue,
	"fg-bright-magenta": color.FgHiMagenta,
	"fg-bright-cyan":    color.FgHiCyan,
	"fg-bright-white":   color.FgHiWhite,
	"fg-hi-black":       color.FgHiBlack,
	"fg-hi-red":         color.FgHiRed,
	"fg-hi-green":       color.FgHiGreen,
	"fg-hi-yellow":      color.FgHiYellow,
	"fg-hi-blue":        color.FgHiBlue,
	"fg-hi-magenta":     color.FgHiMagenta,
	"fg-hi-cyan":        color.FgHiCyan,
	"fg-hi-white":       color.FgHiWhite,

	// Background basic
	"bg-black":   color.BgBlack,
	"bg-red":     color.BgRed,
	"bg-green":   color.BgGreen,
	"bg-yellow":  color.BgYellow,
	"bg-blue":    color.BgBlue,
	"bg-magenta": color.BgMagenta,
	"bg-cyan":    color.BgCyan,
	"bg-white":   color.BgWhite,

	// Background bright / high intensity
	"bg-bright-black":   color.BgHiBlack,
	"bg-bright-red":     color.BgHiRed,
	"bg-bright-green":   color.BgHiGreen,
	"bg-bright-yellow":  color.BgHiYellow,
	"bg-bright-blue":    color.BgHiBlue,
	"bg-bright-magenta": color.BgHiMagenta,
	"bg-bright-cyan":    color.BgHiCyan,
	"bg-bright-white":   color.BgHiWhite,
	"bg-hi-black":       color.BgHiBlack,
	"bg-hi-red":         color.BgHiRed,
	"bg-hi-green":       color.BgHiGreen,
	"bg-hi-yellow":      color.BgHiYellow,
	"bg-hi-blue":        color.BgHiBlue,
	"bg-hi-magenta":     color.BgHiMagenta,
	"bg-hi-cyan":        color.BgHiCyan,
	"bg-hi-white":       color.BgHiWhite,

	// Text attributes
	"bold":          color.Bold,
	"faint":         color.Faint,
	"italic":        color.Italic,
	"underline":     color.Underline,
	"reverse":       color.ReverseVideo,
	"reverse-video": color.ReverseVideo,
	"reversevideo":  color.ReverseVideo,
	"crossed-out":   color.CrossedOut,
	"crossedout":    color.CrossedOut,
	"strikethrough": color.CrossedOut,
	"reset":         color.Reset,
	"normal":        color.Reset,
	"default":       color.Reset,
}

// ParseColor parses a human-readable color string (such as "bright-white", "green", "bold, red") into a color.Color.
func ParseColor(input string) (*color.Color, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, nil
	}

	fields := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})

	var attrs []color.Attribute
	for _, f := range fields {
		f = strings.ToLower(strings.TrimSpace(f))
		if f == "" {
			continue
		}
		norm := strings.ReplaceAll(f, "_", "-")
		attr, ok := colorAttrMap[norm]
		if !ok {
			return nil, fmt.Errorf("unknown color or attribute %q", f)
		}
		attrs = append(attrs, attr)
	}

	if len(attrs) == 0 {
		return nil, nil
	}

	return color.New(attrs...), nil
}

// SetSelected overrides the Selected styling function.
func SetSelected(c *color.Color) {
	if c != nil {
		Selected = wrapper(c.SprintFunc())
	}
}

// SetLayer overrides the Layer styling function.
func SetLayer(c *color.Color) {
	if c != nil {
		Layer = wrapper(c.SprintFunc())
	}
}

// SetHeader overrides the Header styling function.
func SetHeader(c *color.Color) {
	if c != nil {
		Header = wrapper(c.SprintFunc())
	}
}
