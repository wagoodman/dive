package event

import (
	"github.com/wagoodman/go-partybus"
)

const (
	typePrefix = "dive-cli"

	// TaskStarted encompasses all events that are related to the analysis of a docker image (build, fetch, analyze)
	TaskStarted partybus.EventType = typePrefix + "-task-started"

	// ExploreAnalysis is a partybus event that occurs when an analysis result is ready for presentation to stdout
	ExploreAnalysis partybus.EventType = typePrefix + "-analysis"

	// Report is a partybus event that occurs when an analysis result is ready for final presentation to stdout
	Report partybus.EventType = typePrefix + "-report"

	// Notification is a partybus event that occurs when auxiliary information is ready for presentation to stderr
	Notification partybus.EventType = typePrefix + "-notification"
)
