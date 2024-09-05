package pub

import (
	"context"

	"github.com/grassrootseconomics/celo-tracker/pkg/event"
)

type Pub interface {
	Send(context.Context, event.Event) error
	Close()
}
