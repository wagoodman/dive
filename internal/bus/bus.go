package bus

import (
	"github.com/anchore/clio"
	"github.com/wagoodman/go-partybus"
)

var publisher partybus.Publisher

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
	Publish(clio.ExitEvent(true))
}
