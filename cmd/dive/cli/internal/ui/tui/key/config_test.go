package key

import (
	"reflect"
	"testing"

	"github.com/awesome-gocui/gocui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// terminals send the same byte for enter and ctrl+m, so a default bound to either fires when the user hits enter
func TestDefaultBindings_notEnter(t *testing.T) {
	groups := reflect.ValueOf(DefaultBindings())
	for i := range groups.NumField() {
		group := groups.Field(i)
		for j := range group.NumField() {
			name := groups.Type().Field(i).Name + "." + group.Type().Field(j).Name
			cfg := group.Field(j).Interface().(Config)
			require.NoError(t, cfg.Setup(), name)
			for _, k := range cfg.Keys {
				assert.False(t, k.Value == gocui.KeyEnter && k.Modifier == gocui.ModNone, "%s is bound to enter", name)
			}
		}
	}
}
