package kmaster

import (
	"context"
	"fmt"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
)

// Stubs so the package compiles before each backend is fleshed out.
// PT9–PT12 each move one constructor out of this file into its own backend.go.

func NewPKCS11(ctx context.Context, cfg *config.Config) (Wrapper, error) {
	return nil, fmt.Errorf("pkcs11 backend not yet implemented")
}
