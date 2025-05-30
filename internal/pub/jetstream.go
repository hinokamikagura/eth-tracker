package pub

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type (
	JetStreamOpts struct {
		Endpoint        string
		PersistDuration time.Duration
		Logg            *slog.Logger
	}

	jetStreamPub struct {
		js       jetstream.JetStream
		natsConn *nats.Conn
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

	js, err := jetstream.New(natsConn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	js.CreateStream(ctx, jetstream.StreamConfig{
		Name:       streamName,
		Subjects:   streamSubjects,
		MaxAge:     o.PersistDuration,
		Storage:    jetstream.FileStorage,
		Duplicates: time.Minute * 20,
	})

	return &jetStreamPub{
		natsConn: natsConn,
		js:       js,
	}, nil
}

func (p *jetStreamPub) Close() {
	if p.natsConn != nil {
		p.natsConn.Close()
	}
}

func (p *jetStreamPub) Send(ctx context.Context, payload event.Event) error {
	data, err := payload.Serialize()
	if err != nil {
		return err
	}

	_, err = p.js.Publish(
		ctx,
		fmt.Sprintf("%s.%s", streamName, payload.TxType),
		data,
		jetstream.WithMsgID(fmt.Sprintf("%s:%d", payload.TxHash, payload.Index)),
	)
	if err != nil {
		return err
	}

	return nil
}
