package natsconn

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/config"
)

// Connection wraps NATS connection and JetStream
type Connection struct {
	NC  *nats.Conn
	JS  jetstream.JetStream
	cfg *config.Config
	log *slog.Logger
}

// Provide creates a NATS connection for dependency injection
func Provide(i do.Injector) (*Connection, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i).With("component", "nats")

	log.Debug("connecting to NATS", "url", cfg.NATS.URL)

	var opts []nats.Option
	if cfg.NATS.Username != "" {
		opts = append(opts, nats.UserInfo(cfg.NATS.Username, cfg.NATS.Password))
	}

	nc, err := nats.Connect(cfg.NATS.URL, opts...)
	if err != nil {
		log.Error("failed to connect to NATS", "error", err)
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		log.Error("failed to create JetStream context", "error", err)
		return nil, err
	}

	log.Info("connected to NATS", "url", cfg.NATS.URL)

	return &Connection{
		NC:  nc,
		JS:  js,
		cfg: cfg,
		log: log,
	}, nil
}

// Close closes the NATS connection
func (c *Connection) Close() {
	if c.NC != nil {
		c.NC.Close()
	}
}

// EnsureStream creates or updates a JetStream stream
func (c *Connection) EnsureStream(ctx context.Context, name string, subjects []string) (jetstream.Stream, error) {
	c.log.Debug("ensuring stream exists", "name", name, "subjects", subjects)

	stream, err := c.JS.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     name,
		Subjects: subjects,
	})
	if err != nil {
		c.log.Error("failed to ensure stream", "name", name, "error", err)
		return nil, err
	}

	return stream, nil
}

// Publish publishes a message to NATS
func (c *Connection) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := c.JS.Publish(ctx, subject, data)
	return err
}

// PublishWithHeaders publishes a message with headers to NATS
func (c *Connection) PublishWithHeaders(ctx context.Context, subject string, headers map[string]string, data []byte) error {
	msg := &nats.Msg{
		Subject: subject,
		Header:  make(nats.Header),
		Data:    data,
	}
	for k, v := range headers {
		msg.Header.Set(k, v)
	}

	_, err := c.JS.PublishMsg(ctx, msg)
	return err
}

// Subscribe creates a consumer and subscribes to subjects
func (c *Connection) Subscribe(
	ctx context.Context,
	stream string,
	subjects []string,
	durable string,
	handler func(msg jetstream.Msg),
) (jetstream.ConsumeContext, error) {
	consumerName := c.cfg.NATS.Prefix + "-" + durable

	c.log.Debug("creating consumer", "stream", stream, "consumer", consumerName, "subjects", subjects)

	consumer, err := c.JS.CreateOrUpdateConsumer(ctx, stream, jetstream.ConsumerConfig{
		Durable:        consumerName,
		AckPolicy:      jetstream.AckExplicitPolicy,
		FilterSubjects: subjects,
	})
	if err != nil {
		c.log.Error("failed to create consumer", "error", err)
		return nil, err
	}

	cc, err := consumer.Consume(handler)
	if err != nil {
		c.log.Error("failed to start consuming", "error", err)
		return nil, err
	}

	c.log.Info("subscribed to subjects", "stream", stream, "consumer", consumerName, "subjects", subjects)
	return cc, nil
}
