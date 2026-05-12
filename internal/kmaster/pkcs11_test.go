package kmaster

import (
	"context"
	"strings"
	"testing"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
)

func TestPKCS11_Deferred(t *testing.T) {
	_, err := NewPKCS11(context.Background(), &config.Config{})
	if err == nil {
		t.Fatal("expected deferred-backend error")
	}
	if !strings.Contains(err.Error(), "pkcs11") {
		t.Errorf("expected pkcs11 in error, got: %v", err)
	}
}
