// Package authstate provides NATS JetStream KV storage for OAuth2 PKCE state
package authstate

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/queue"
)

const (
	// DefaultTTL is the time-to-live for auth states (10 minutes)
	DefaultTTL = 10 * time.Minute

	// BucketSuffix is appended to the queue prefix for the bucket name
	BucketSuffix = "auth-state"
)

// State holds the PKCE code verifier and redirect URI for an in-flight OAuth2 flow
type State struct {
	CodeVerifier string `json:"code_verifier"`
	RedirectURI  string `json:"redirect_uri"`
}

// Store manages OAuth2 auth states in NATS KV
type Store struct {
	kv  jetstream.KeyValue
	log *slog.Logger
}

// Provide creates a Store for dependency injection
func Provide(i do.Injector) (*Store, error) {
	q := do.MustInvoke[*queue.Queue](i)
	log := do.MustInvoke[*slog.Logger](i).With("component", "authstate")

	ctx := context.Background()
	bucketName := q.Prefix() + "-" + BucketSuffix

	log.Debug("creating auth state KV bucket", "bucket", bucketName)

	kv, err := q.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      bucketName,
		Description: "OAuth2 PKCE state storage",
		TTL:         DefaultTTL,
	})
	if err != nil {
		log.Error("failed to create KV bucket", "error", err)
		return nil, err
	}

	log.Info("auth state KV store initialized", "bucket", bucketName, "ttl", DefaultTTL)

	return &Store{
		kv:  kv,
		log: log,
	}, nil
}

// Put stores an auth state keyed by the state token
func (s *Store) Put(ctx context.Context, stateToken string, state *State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	_, err = s.kv.Put(ctx, stateToken, data)
	if err != nil {
		s.log.Error("failed to put auth state", "error", err)
		return err
	}

	s.log.Debug("auth state stored", "state_prefix", stateToken[:8]+"...")
	return nil
}

// GetAndDelete retrieves and atomically deletes an auth state for replay protection
func (s *Store) GetAndDelete(ctx context.Context, stateToken string) (*State, error) {
	entry, err := s.kv.Get(ctx, stateToken)
	if err != nil {
		return nil, err
	}

	// Delete immediately for replay protection
	if err := s.kv.Delete(ctx, stateToken); err != nil {
		s.log.Warn("failed to delete auth state after get", "error", err)
	}

	var state State
	if err := json.Unmarshal(entry.Value(), &state); err != nil {
		return nil, err
	}

	return &state, nil
}
