package queue

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
)

// Queue service
type Queue struct {
	ctx context.Context
	cfg *config.Config
	log *slog.Logger
	nc  *nats.Conn
	js  jetstream.JetStream
}

// Provide queue service instance provider for dependency injection
func Provide(injector do.Injector) (*Queue, error) {
	ctx := do.MustInvoke[context.Context](injector)
	cfg := do.MustInvoke[*config.Config](injector)
	log := do.MustInvoke[*slog.Logger](injector).With("service", "queue")

	log.Debug("connecting to NATS server", "url", cfg.Queue.URL)
	nc, err := nats.Connect(cfg.Queue.URL)
	if err != nil {
		log.Error("failed to connect to NATS server", "error", err)
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		log.Error("failed to connect to JetStream", "error", err)
		return nil, err
	}

	return &Queue{
		ctx: ctx,
		cfg: cfg,
		log: log,
		nc:  nc,
		js:  js,
	}, nil
}

// Publish message to the queue
func (q *Queue) Publish(ctx context.Context, subject string, data []byte) error {
	log := q.log.With("method", "publish", "subject", subject)
	log.Debug("publish message")
	_, err := q.js.PublishMsg(ctx, &nats.Msg{
		Subject: subject,
		Data:    data,
	})
	if err != nil {
		log.Error("failed to publish message", "error", err)
		return err
	}
	return err
}

// Ensure that stream created
func (q *Queue) Ensure(ctx context.Context, stream string, subjects []string) error {
	log := q.log.With("method", "ensure", "stream", stream, slog.Any("subjects", subjects))
	log.Debug("ensure stream exists")
	_, err := q.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     stream,
		Subjects: subjects,
	})
	if err != nil {
		log.Error("failed to create stream", "error", err)
		return err
	}
	return err
}

// Consume messages from the stream with ability to filter by subjects
func (q *Queue) Consume(
	ctx context.Context, stream string, subjects []string, durable string, handler func(msg Msg),
) (cancel func(), err error) {
	log := q.log.With("method", "consume", "stream", stream, slog.Any("subjects", subjects))
	log.Debug("start consuming messages")
	csCfg := jetstream.ConsumerConfig{
		AckPolicy:      jetstream.AckExplicitPolicy,
		FilterSubjects: subjects,
	}
	if durable != "" {
		csCfg.Durable = fmt.Sprintf("%s-%s", q.cfg.Queue.Prefix, durable)
	}
	consumer, err := q.js.CreateOrUpdateConsumer(ctx, stream, csCfg)
	if err != nil {
		log.Error("failed to create consumer", "error", err)
		return nil, err
	}
	cc, err := consumer.Consume(func(msg jetstream.Msg) {
		handler(msg)
	})
	if err != nil {
		log.Error("failed to consume message", "error", err)
		return nil, err
	}
	return cc.Stop, nil
}
