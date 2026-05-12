package kmaster

import "context"

type Wrapper interface {
	Wrap(ctx context.Context, plaintext []byte, aad []byte) (WrappedKey, error)
	Unwrap(ctx context.Context, wrapped WrappedKey, aad []byte) ([]byte, error)
	CurrentVersion(ctx context.Context) (int, error)
	LoadedVersions(ctx context.Context) ([]int, error)
	Close() error
}

type WrappedKey struct {
	Version    int    `json:"v"`
	Ciphertext []byte `json:"ct"`
}
