package pub

import (
	"context"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
)

type Pub interface {
	Send(context.Context, event.Event) error
	Close()
}
