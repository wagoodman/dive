package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFieldDescriptionSet struct {
	descriptions map[any]string
}

func (m *mockFieldDescriptionSet) Add(ptr any, description string) {
	if m.descriptions == nil {
		m.descriptions = make(map[any]string)
	}
	m.descriptions[ptr] = description
}

func TestDefaultUIColors(t *testing.T) {
	colors := DefaultUIColors()
	assert.Empty(t, colors.Selected)
	assert.Empty(t, colors.Added)
	assert.Empty(t, colors.Removed)
	assert.Empty(t, colors.Modified)
	assert.Empty(t, colors.Unmodified)
	assert.Empty(t, colors.Layer)
	assert.Empty(t, colors.Header)
}

func TestUIColors_DescribeFields(t *testing.T) {
	colors := DefaultUIColors()
	mock := &mockFieldDescriptionSet{}
	colors.DescribeFields(mock)
	assert.Contains(t, mock.descriptions[&colors.Selected], "selected")
	assert.Contains(t, mock.descriptions[&colors.Added], "added")
	assert.Contains(t, mock.descriptions[&colors.Removed], "removed")
	assert.Contains(t, mock.descriptions[&colors.Modified], "modified")
	assert.Contains(t, mock.descriptions[&colors.Unmodified], "unmodified")
	assert.Contains(t, mock.descriptions[&colors.Layer], "layer")
	assert.Contains(t, mock.descriptions[&colors.Header], "header")
}

func TestUIColors_PostLoad(t *testing.T) {
	t.Run("valid colors", func(t *testing.T) {
		colors := UIColors{
			Selected:   "bright-white",
			Added:      "green",
			Removed:    "red",
			Modified:   "yellow",
			Unmodified: "white",
			Layer:      "blue",
			Header:     "bold",
		}
		require.NoError(t, colors.PostLoad())
	})

	t.Run("empty colors", func(t *testing.T) {
		colors := DefaultUIColors()
		require.NoError(t, colors.PostLoad())
	})

	t.Run("invalid colors log warning but do not fail", func(t *testing.T) {
		colors := UIColors{
			Selected: "not-a-valid-color",
		}
		require.NoError(t, colors.PostLoad())
	})
}

func TestUIColors_ToTUIColors(t *testing.T) {
	t.Run("valid colors mapping", func(t *testing.T) {
		colors := UIColors{
			Selected:   "bright-white",
			Added:      "green",
			Removed:    "red",
			Modified:   "yellow",
			Unmodified: "cyan",
			Layer:      "blue",
			Header:     "bold",
		}

		tuiColors := colors.ToTUIColors()
		assert.NotNil(t, tuiColors.Selected)
		assert.NotNil(t, tuiColors.Added)
		assert.NotNil(t, tuiColors.Removed)
		assert.NotNil(t, tuiColors.Modified)
		assert.NotNil(t, tuiColors.Unmodified)
		assert.NotNil(t, tuiColors.Layer)
		assert.NotNil(t, tuiColors.Header)
	})

	t.Run("fallback handling for empty fields", func(t *testing.T) {
		colors := UIColors{
			Selected: "bright-white",
			// other fields empty
		}

		tuiColors := colors.ToTUIColors()
		assert.NotNil(t, tuiColors.Selected)
		assert.Nil(t, tuiColors.Added)
		assert.Nil(t, tuiColors.Removed)
		assert.Nil(t, tuiColors.Modified)
		assert.Nil(t, tuiColors.Unmodified)
		assert.Nil(t, tuiColors.Layer)
		assert.Nil(t, tuiColors.Header)
	})

	t.Run("fallback handling for invalid color values", func(t *testing.T) {
		colors := UIColors{
			Selected: "invalid-color-value",
			Added:    "green",
		}

		tuiColors := colors.ToTUIColors()
		assert.Nil(t, tuiColors.Selected, "invalid color string should fallback to nil")
		assert.NotNil(t, tuiColors.Added)
	})
}

func TestApplication_TUIPreferences_Colors(t *testing.T) {
	app := DefaultApplication()
	app.UI.Colors = UIColors{
		Selected: "bright-white",
		Layer:    "yellow",
	}

	prefs := app.TUIPreferences()
	assert.NotNil(t, prefs.Colors.Selected)
	assert.NotNil(t, prefs.Colors.Layer)
	assert.Nil(t, prefs.Colors.Added)
}
