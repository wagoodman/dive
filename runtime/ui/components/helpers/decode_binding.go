package helpers

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/wagoodman/dive/internal/log"
	"gitlab.com/tslocum/cbind"
)

func DecodeBinding(key string) *tcell.EventKey {
	mod, tKey, ch, err := cbind.Decode(key)
	if err != nil {
		panic(err)
	}
	log.WithFields(
		"configuredKey", key,
		"mod", mod,
		"decodedKey", tKey,
		"char", fmt.Sprintf("%+v", ch),
	).Tracef("creating key event")

	return tcell.NewEventKey(tKey, ch, mod)
}
