package runtime

import (
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/dive/runtime/logger"
)

// SetLogger sets the logger object used for all logging calls.
func SetLogger(logger logger.Logger) {
	log.Log = logger
}
