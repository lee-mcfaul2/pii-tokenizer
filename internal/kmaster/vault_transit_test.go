//go:build integration_vault

package kmaster

import (
	"context"
	"os"
	"testing"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
)

func TestVaultTransit_WrapUnwrap(t *testing.T) {
	if os.Getenv("VAULT_ADDR") == "" {
		t.Skip("VAULT_ADDR not set")
	}
	w, err := NewVaultTransit(context.Background(), &config.Config{
		VaultAddr:    os.Getenv("VAULT_ADDR"),
		VaultKeyName: "pii-tokenizer-test",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	pt := []byte("a-K_request-here-but-still-fake")
	wrapped, err := w.Wrap(context.Background(), pt, []byte("uuid-a"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := w.Unwrap(context.Background(), wrapped, []byte("uuid-a"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(pt) {
		t.Errorf("plaintext mismatch")
	}
}
