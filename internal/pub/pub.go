// Package pub provides the interface for event publishing.
package pub

import (
	"context"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
)

// Pub defines the interface for publishing events to external systems.
type Pub interface {
	// Send publishes an event to the configured destination.
	Send(context.Context, event.Event) error

	// Close closes the publisher and releases any resources.
	Close()
}
