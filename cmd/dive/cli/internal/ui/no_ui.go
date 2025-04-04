package ui

import (
	"github.com/wagoodman/go-partybus"

	"github.com/anchore/clio"
)

var _ clio.UI = (*NoUI)(nil)

type NoUI struct {
	subscription partybus.Unsubscribable
}

func None() *NoUI {
	return &NoUI{}
}

func (n *NoUI) Setup(subscription partybus.Unsubscribable) error {
	n.subscription = subscription
	return nil
}

func (n *NoUI) Handle(_ partybus.Event) error {
	return nil
}

func (n NoUI) Teardown(_ bool) error {
	return nil
}
