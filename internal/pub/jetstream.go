// Package pub provides event publishing functionality to NATS JetStream.
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

const (
	// streamName is the name of the NATS JetStream stream
	streamName = "TRACKER"
	// streamSubjectPattern is the subject pattern for tracker events
	streamSubjectPattern = "TRACKER.*"
	// streamCreateTimeout is the timeout for stream creation
	streamCreateTimeout = 10 * time.Second
	// duplicateWindow is the time window for duplicate message detection
	duplicateWindow = 20 * time.Minute
)

var (
	// streamSubjects contains the subject patterns for the JetStream stream
	streamSubjects = []string{
		streamSubjectPattern,
	}
)

type (
	// JetStreamOpts contains configuration options for creating a new JetStream publisher.
	JetStreamOpts struct {
		Endpoint        string        // NATS server endpoint
		PersistDuration time.Duration // Message persistence duration
		Logg            *slog.Logger  // Structured logger
	}

	// jetStreamPub implements the Pub interface using NATS JetStream.
	jetStreamPub struct {
		js       jetstream.JetStream
		natsConn *nats.Conn
		logg     *slog.Logger
	}
)

// NewJetStreamPub creates a new JetStream publisher and initializes the stream.
func NewJetStreamPub(o JetStreamOpts) (Pub, error) {
	natsConn, err := nats.Connect(o.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := jetstream.New(natsConn)
	if err != nil {
		natsConn.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), streamCreateTimeout)
	defer cancel()

	// Create or update the stream
	_, err = js.CreateStream(ctx, jetstream.StreamConfig{
		Name:       streamName,
		Subjects:   streamSubjects,
		MaxAge:     o.PersistDuration,
		Storage:    jetstream.FileStorage,
		Duplicates: duplicateWindow,
	})
	if err != nil {
		natsConn.Close()
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	o.Logg.Info("JetStream publisher initialized",
		"stream", streamName,
		"subjects", streamSubjects,
		"persist_duration", o.PersistDuration,
	)

	return &jetStreamPub{
		natsConn: natsConn,
		js:       js,
		logg:     o.Logg,
	}, nil
}

// Close closes the NATS connection.
func (p *jetStreamPub) Close() {
	if p.natsConn != nil {
		p.natsConn.Close()
		p.logg.Debug("NATS connection closed")
	}
}

// Send publishes an event to NATS JetStream.
// The event is published to a subject based on the transaction type.
func (p *jetStreamPub) Send(ctx context.Context, payload event.Event) error {
	data, err := payload.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	subject := fmt.Sprintf("%s.%s", streamName, payload.TxType)
	msgID := fmt.Sprintf("%s:%d", payload.TxHash, payload.Index)

	_, err = p.js.Publish(ctx, subject, data, jetstream.WithMsgID(msgID))
	if err != nil {
		return fmt.Errorf("failed to publish event to %s: %w", subject, err)
	}

	return nil
}
