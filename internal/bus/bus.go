package bus

import "github.com/wagoodman/go-partybus"

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
