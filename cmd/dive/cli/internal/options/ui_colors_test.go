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

func TestUIColors_ToV1Colors(t *testing.T) {
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

		v1Colors := colors.ToV1Colors()
		assert.NotNil(t, v1Colors.Selected)
		assert.NotNil(t, v1Colors.Added)
		assert.NotNil(t, v1Colors.Removed)
		assert.NotNil(t, v1Colors.Modified)
		assert.NotNil(t, v1Colors.Unmodified)
		assert.NotNil(t, v1Colors.Layer)
		assert.NotNil(t, v1Colors.Header)
	})

	t.Run("fallback handling for empty fields", func(t *testing.T) {
		colors := UIColors{
			Selected: "bright-white",
			// other fields empty
		}

		v1Colors := colors.ToV1Colors()
		assert.NotNil(t, v1Colors.Selected)
		assert.Nil(t, v1Colors.Added)
		assert.Nil(t, v1Colors.Removed)
		assert.Nil(t, v1Colors.Modified)
		assert.Nil(t, v1Colors.Unmodified)
		assert.Nil(t, v1Colors.Layer)
		assert.Nil(t, v1Colors.Header)
	})

	t.Run("fallback handling for invalid color values", func(t *testing.T) {
		colors := UIColors{
			Selected: "invalid-color-value",
			Added:    "green",
		}

		v1Colors := colors.ToV1Colors()
		assert.Nil(t, v1Colors.Selected, "invalid color string should fallback to nil")
		assert.NotNil(t, v1Colors.Added)
	})
}

func TestApplication_V1Preferences_Colors(t *testing.T) {
	app := DefaultApplication()
	app.UI.Colors = UIColors{
		Selected: "bright-white",
		Layer:    "yellow",
	}

	prefs := app.V1Preferences()
	assert.NotNil(t, prefs.Colors.Selected)
	assert.NotNil(t, prefs.Colors.Layer)
	assert.Nil(t, prefs.Colors.Added)
}
