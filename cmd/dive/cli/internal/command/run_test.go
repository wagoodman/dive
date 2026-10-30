package command

import (
	"errors"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestEscaped(t *testing.T) {
	// a registry body that would set the clipboard (OSC 52) and wipe the line, plus a bidi override
	inner := errors.New("404: \x1b]52;c;cHduZWQ=\a\x1b[2K\r/etc/\u202edwssap\nsecond")
	err := escaped(func(*cobra.Command, []string) error {
		return fmt.Errorf("cannot load image: %w", inner)
	})(nil, nil)

	assert.Equal(t, `cannot load image: 404: \x1b]52;c;cHduZWQ=\a\x1b[2K\r/etc/\u202edwssap`+"\nsecond", err.Error())
	assert.ErrorIs(t, err, inner)

	assert.NoError(t, escaped(func(*cobra.Command, []string) error { return nil })(nil, nil))
}
