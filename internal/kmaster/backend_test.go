package kmaster

import (
	"context"
	"testing"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
)

func TestWrappedKey_Zero(t *testing.T) {
	var w WrappedKey
	if w.Version != 0 || w.Ciphertext != nil {
		t.Errorf("zero value wrong: %+v", w)
	}
}

func TestOpen_UnknownBackend(t *testing.T) {
	_, err := Open(context.Background(), &config.Config{KMasterBackend: "bogus"})
	if err == nil {
		t.Fatal("expected error for unknown backend")
	}
}
