package bus

import (
	"sync/atomic"

	"github.com/anchore/clio"
	"github.com/wagoodman/go-partybus"
)

var publisher partybus.Publisher

var interrupted atomic.Bool

func Set(p partybus.Publisher) {
	publisher = p
}

func Get() partybus.Publisher {
	return publisher
}

func Publish(e partybus.Event) {
	if publisher != nil {
		publisher.Publish(e)
	}
}

// ExitWithInterrupt asks the clio event loop to stop now (the UI holds the terminal in raw mode, so ctrl-c arrives
// as a key press, not as a SIGINT)
func ExitWithInterrupt() {
	interrupted.Store(true)
	Publish(clio.ExitEvent(true))
}

// Interrupted reports whether ExitWithInterrupt was called. clio exits 0 after an interrupt, main uses this to exit
// like a ctrl-c would instead.
func Interrupted() bool {
	return interrupted.Load()
}
