package format

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseColor(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectNil bool
		wantErr   bool
	}{
		{
			name:      "empty string",
			input:     "",
			expectNil: true,
			wantErr:   false,
		},
		{
			name:      "whitespace only",
			input:     "   \t\n  ",
			expectNil: true,
			wantErr:   false,
		},
		{
			name:      "basic color green",
			input:     "green",
			expectNil: false,
			wantErr:   false,
		},
		{
			name:      "basic color red",
			input:     "red",
			expectNil: false,
			wantErr:   false,
		},
		{
			name:      "prefixed foreground",
			input:     "fg-blue",
			expectNil: false,
			wantErr:   false,
		},
		{
			name:      "bright color",
			input:     "bright-white",
			expectNil: false,
			wantErr:   false,
		},
		{
			name:      "hi alias",
			input:     "hi-yellow",
			expectNil: false,
			wantErr:   false,
		},
		{
			name:      "background color",
			input:     "bg-black",
			expectNil: false,
			wantErr:   false,
		},
		{
			name:      "attribute bold",
			input:     "bold",
			expectNil: false,
			wantErr:   false,
		},
		{
			name:      "attribute reverse",
			input:     "reverse",
			expectNil: false,
			wantErr:   false,
		},
		{
			name:      "multiple comma separated",
			input:     "bright-white, bold",
			expectNil: false,
			wantErr:   false,
		},
		{
			name:      "multiple space separated",
			input:     "reverse bold yellow",
			expectNil: false,
			wantErr:   false,
		},
		{
			name:      "case insensitive and underscore",
			input:     "BRIGHT_WHITE",
			expectNil: false,
			wantErr:   false,
		},
		{
			name:      "invalid color name",
			input:     "neon-pink",
			expectNil: true,
			wantErr:   true,
		},
		{
			name:      "partially invalid list",
			input:     "bold, invalid-color",
			expectNil: true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ParseColor(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, c)
			} else {
				require.NoError(t, err)
				if tt.expectNil {
					assert.Nil(t, c)
				} else {
					require.NotNil(t, c)
				}
			}
		})
	}
}

func TestSettersAndReset(t *testing.T) {
	defer ResetDefaults()

	customColor := color.New(color.FgCyan)

	// SetSelected
	SetSelected(customColor)
	assert.NotEmpty(t, Selected("test"))
	SetSelected(nil) // should not panic or reset
	assert.NotEmpty(t, Selected("test"))

	// SetLayer
	SetLayer(customColor)
	assert.NotEmpty(t, Layer("test"))
	SetLayer(nil)
	assert.NotEmpty(t, Layer("test"))

	// SetHeader
	SetHeader(customColor)
	assert.NotEmpty(t, Header("test"))
	SetHeader(nil)
	assert.NotEmpty(t, Header("test"))

	// ResetDefaults
	ResetDefaults()
	assert.NotEmpty(t, Selected("test"))
	assert.NotEmpty(t, Layer("test"))
	assert.NotEmpty(t, Header("test"))
}
