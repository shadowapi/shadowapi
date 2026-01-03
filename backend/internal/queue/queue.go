package queue

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do/v2"
	"log/slog"

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

type Headers map[string]string

// Provide queue service instance provider for dependency injection
func Provide(injector do.Injector) (*Queue, error) {
	ctx := do.MustInvoke[context.Context](injector)
	cfg := do.MustInvoke[*config.Config](injector)
	log := do.MustInvoke[*slog.Logger](injector).With("service", "queue")

	log.Debug("connecting to NATS server", "url", cfg.Queue.URL)
	
	var opts []nats.Option
	if cfg.Queue.Username != "" {
		opts = append(opts, nats.UserInfo(cfg.Queue.Username, cfg.Queue.Password))
	}
	nc, err := nats.Connect(cfg.Queue.URL, opts...)

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
func (q *Queue) PublishWithHeaders(ctx context.Context, subject string, headers Headers, data []byte) error {
	msg := &nats.Msg{
		Subject: subject,
		Header:  make(nats.Header),
		Data:    data,
	}
	for k, v := range headers {
		msg.Header.Set(k, v)
	}

	// Correct argument order: msg first, then options
	_, err := q.js.PublishMsg(ctx, msg)
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

// CreateOrUpdateKeyValue creates or updates a key-value bucket
func (q *Queue) CreateOrUpdateKeyValue(ctx context.Context, cfg jetstream.KeyValueConfig) (jetstream.KeyValue, error) {
	return q.js.CreateOrUpdateKeyValue(ctx, cfg)
}

// Prefix returns the configured queue prefix
func (q *Queue) Prefix() string {
	return q.cfg.Queue.Prefix
}

// Consume messages from the stream with ability to filter by subjects
func (q *Queue) Consume(
	ctx context.Context,
	stream string,
	subjects []string,
	durable string,
	handler func(msg Msg),
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
		log.Error("failed to consume messages", "error", err)
		return nil, err
	}
	return cc.Stop, nil
}

// StreamMessage represents a message retrieved from a stream
type StreamMessage struct {
	Subject   string
	Data      []byte
	Headers   map[string]string
	Sequence  uint64
	Timestamp int64
}

// GetLastMessages retrieves the last N messages from a stream for a given subject filter
func (q *Queue) GetLastMessages(ctx context.Context, stream string, subjectFilter string, limit int) ([]StreamMessage, error) {
	log := q.log.With("method", "GetLastMessages", "stream", stream, "subject", subjectFilter, "limit", limit)
	log.Debug("fetching last messages")

	// Get the stream
	s, err := q.js.Stream(ctx, stream)
	if err != nil {
		log.Error("failed to get stream", "error", err)
		return nil, err
	}

	// Get stream info to find the last sequence
	info, err := s.Info(ctx)
	if err != nil {
		log.Error("failed to get stream info", "error", err)
		return nil, err
	}

	if info.State.Msgs == 0 {
		return []StreamMessage{}, nil
	}

	// Create an ephemeral consumer that delivers from the last N messages
	consumerCfg := jetstream.ConsumerConfig{
		AckPolicy:      jetstream.AckNonePolicy,
		DeliverPolicy:  jetstream.DeliverLastPerSubjectPolicy,
		FilterSubject:  subjectFilter,
		MaxDeliver:     1,
		InactiveThreshold: 30_000_000_000, // 30 seconds
	}

	// If we want last N messages, we need to use DeliverByStartSequence
	// Calculate start sequence
	startSeq := uint64(1)
	if info.State.LastSeq > uint64(limit) {
		startSeq = info.State.LastSeq - uint64(limit) + 1
	}

	consumerCfg.DeliverPolicy = jetstream.DeliverByStartSequencePolicy
	consumerCfg.OptStartSeq = startSeq

	consumer, err := s.CreateConsumer(ctx, consumerCfg)
	if err != nil {
		log.Error("failed to create consumer", "error", err)
		return nil, err
	}

	// Fetch messages
	messages := make([]StreamMessage, 0, limit)
	batch, err := consumer.Fetch(limit, jetstream.FetchMaxWait(2_000_000_000)) // 2 seconds
	if err != nil {
		log.Error("failed to fetch messages", "error", err)
		return nil, err
	}

	for msg := range batch.Messages() {
		meta, _ := msg.Metadata()
		headers := make(map[string]string)
		for k, v := range msg.Headers() {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		messages = append(messages, StreamMessage{
			Subject:   msg.Subject(),
			Data:      msg.Data(),
			Headers:   headers,
			Sequence:  meta.Sequence.Stream,
			Timestamp: meta.Timestamp.UnixMilli(),
		})
	}

	if batch.Error() != nil {
		log.Warn("batch fetch error", "error", batch.Error())
	}

	log.Debug("fetched messages", "count", len(messages))
	return messages, nil
}
