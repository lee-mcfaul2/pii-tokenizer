//go:build integration_aws

package kmaster

import (
	"context"
	"os"
	"testing"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
)

func TestAWSKMS_WrapUnwrap(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" || os.Getenv("TOKENIZER_AWS_KMS_KEY_ID") == "" {
		t.Skip("AWS env not configured")
	}
	w, err := NewAWSKMS(context.Background(), &config.Config{
		AWSRegion: os.Getenv("AWS_REGION"),
		AWSKeyID:  os.Getenv("TOKENIZER_AWS_KMS_KEY_ID"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	pt := []byte("a-K_request-fake")
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
