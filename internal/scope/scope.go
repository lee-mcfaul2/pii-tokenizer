package scope

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/crypto"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/kmaster"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/store"
)

type Service struct {
	store *store.Client
	km    kmaster.Wrapper
}

func New(s *store.Client, km kmaster.Wrapper) *Service {
	return &Service{store: s, km: km}
}

type InitResult struct {
	ExpiresAt time.Time
	Created   bool
}

func (s *Service) InitRequest(ctx context.Context, uuid string, ttl time.Duration) (InitResult, error) {
	kreq := make([]byte, crypto.KeySize)
	if _, err := rand.Read(kreq); err != nil {
		return InitResult{}, err
	}

	wrapped, err := s.km.Wrap(ctx, kreq, []byte(uuid))
	if err != nil {
		return InitResult{}, fmt.Errorf("wrap K_request: %w", err)
	}
	now := time.Now().UTC()
	entry := store.ScopeEntry{
		WrappedKReq:    wrapped.Ciphertext,
		KMasterVersion: wrapped.Version,
		CreatedAt:      now,
		ExpiresAt:      now.Add(ttl),
	}

	err = s.store.Init(ctx, uuid, entry, ttl)
	if err == nil {
		return InitResult{ExpiresAt: entry.ExpiresAt, Created: true}, nil
	}
	var ae *store.ErrScopeAlreadyExists
	if errors.As(err, &ae) {
		return InitResult{ExpiresAt: ae.ExpiresAt, Created: false}, nil
	}
	return InitResult{}, err
}

func (s *Service) Tokenize(ctx context.Context, uuid, piiType string, plaintext []byte) (string, error) {
	entry, err := s.store.Get(ctx, uuid)
	if err != nil {
		return "", err
	}
	kreq, err := s.km.Unwrap(ctx, kmaster.WrappedKey{
		Version:    entry.KMasterVersion,
		Ciphertext: entry.WrappedKReq,
	}, []byte(uuid))
	if err != nil {
		return "", fmt.Errorf("unwrap K_request: %w", err)
	}
	return crypto.Tokenize(kreq, uuid, piiType, plaintext)
}

func (s *Service) Detokenize(ctx context.Context, uuid, token string) (piiType string, plaintext []byte, err error) {
	entry, err := s.store.Get(ctx, uuid)
	if err != nil {
		return "", nil, err
	}
	kreq, err := s.km.Unwrap(ctx, kmaster.WrappedKey{
		Version:    entry.KMasterVersion,
		Ciphertext: entry.WrappedKReq,
	}, []byte(uuid))
	if err != nil {
		return "", nil, fmt.Errorf("unwrap K_request: %w", err)
	}
	return crypto.Detokenize(kreq, uuid, token)
}

func (s *Service) ReleaseRequest(ctx context.Context, uuid string) error {
	return s.store.Release(ctx, uuid)
}
