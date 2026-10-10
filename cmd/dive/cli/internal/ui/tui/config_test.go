package tui

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/filetree"
)

func TestColors_Apply_And_ResetDefaultColors(t *testing.T) {
	defer ResetDefaultColors()

	customSelected := color.New(color.FgCyan)
	customAdded := color.New(color.FgHiBlue)
	customRemoved := color.New(color.FgHiMagenta)
	customModified := color.New(color.FgHiYellow)
	customUnmodified := color.New(color.FgHiWhite)
	customLayer := color.New(color.FgGreen)
	customHeader := color.New(color.Underline)

	colors := Colors{
		Selected:   customSelected,
		Added:      customAdded,
		Removed:    customRemoved,
		Modified:   customModified,
		Unmodified: customUnmodified,
		Layer:      customLayer,
		Header:     customHeader,
	}

	colors.Apply()

	assert.Equal(t, customAdded, filetree.DiffTypeColor(filetree.Added))
	assert.Equal(t, customRemoved, filetree.DiffTypeColor(filetree.Removed))
	assert.Equal(t, customModified, filetree.DiffTypeColor(filetree.Modified))
	assert.Equal(t, customUnmodified, filetree.DiffTypeColor(filetree.Unmodified))

	ResetDefaultColors()

	assert.NotEqual(t, customAdded, filetree.DiffTypeColor(filetree.Added))
	assert.NotEqual(t, customRemoved, filetree.DiffTypeColor(filetree.Removed))
}

func TestColors_Apply_Fallback(t *testing.T) {
	defer ResetDefaultColors()

	ResetDefaultColors()
	defaultAddedColor := filetree.DiffTypeColor(filetree.Added)

	// Applying an empty Colors struct should not change any diff colors (fallback)
	emptyColors := Colors{}
	emptyColors.Apply()

	assert.Equal(t, defaultAddedColor, filetree.DiffTypeColor(filetree.Added))
}

func TestDefaultPreferences_Colors(t *testing.T) {
	prefs := DefaultPreferences()
	require.NotNil(t, prefs.Colors)
	assert.Nil(t, prefs.Colors.Selected)
	assert.Nil(t, prefs.Colors.Added)
}
