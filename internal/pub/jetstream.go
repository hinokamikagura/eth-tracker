package pub

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/nats-io/nats.go"
)

type (
	JetStreamOpts struct {
		Logg            *slog.Logger
		Endpoint        string
		PersistDuration time.Duration
	}

	jetStreamPub struct {
		natsConn *nats.Conn
		jsCtx    nats.JetStreamContext
	}
)

const streamName string = "TRACKER"

var streamSubjects = []string{
	"TRACKER.*",
}

func NewJetStreamPub(o JetStreamOpts) (Pub, error) {
	natsConn, err := nats.Connect(o.Endpoint)
	if err != nil {
		return nil, err
	}

	js, err := natsConn.JetStream()
	if err != nil {
		return nil, err
	}
	o.Logg.Info("successfully connected to NATS server")

	stream, err := js.StreamInfo(streamName)
	if err != nil && !errors.Is(err, nats.ErrStreamNotFound) {
		return nil, err
	}
	if stream == nil {
		_, err := js.AddStream(&nats.StreamConfig{
			Name:       streamName,
			MaxAge:     o.PersistDuration,
			Storage:    nats.FileStorage,
			Subjects:   streamSubjects,
			Duplicates: time.Minute,
		})
		if err != nil {
			return nil, err
		}
		o.Logg.Info("successfully created NATS JetStream stream", "stream_name", streamName)
	}

	return &jetStreamPub{
		natsConn: natsConn,
		jsCtx:    js,
	}, nil
}

func (p *jetStreamPub) Close() {
	if p.natsConn != nil {
		p.natsConn.Close()
	}
}

func (p *jetStreamPub) Send(_ context.Context, payload event.Event) error {
	data, err := payload.Serialize()
	if err != nil {
		return err
	}

	_, err = p.jsCtx.Publish(
		fmt.Sprintf("%s.%s", streamName, payload.TxType),
		data,
		nats.MsgId(fmt.Sprintf("%s:%d", payload.TxHash, payload.Index)),
	)
	if err != nil {
		return err
	}

	return nil
}
