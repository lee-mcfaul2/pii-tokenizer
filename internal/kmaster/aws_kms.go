package kmaster

import (
	"context"
	"fmt"

	wrapping "github.com/hashicorp/go-kms-wrapping/v2"
	awskms "github.com/hashicorp/go-kms-wrapping/wrappers/awskms/v2"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
	"google.golang.org/protobuf/proto"
)

type awsKMS struct {
	w *awskms.Wrapper
}

func NewAWSKMS(ctx context.Context, cfg *config.Config) (Wrapper, error) {
	if cfg.AWSRegion == "" || cfg.AWSKeyID == "" {
		return nil, fmt.Errorf("AWS_REGION and TOKENIZER_AWS_KMS_KEY_ID required")
	}
	w := awskms.NewWrapper()
	_, err := w.SetConfig(ctx, wrapping.WithConfigMap(map[string]string{
		"region":     cfg.AWSRegion,
		"kms_key_id": cfg.AWSKeyID,
	}))
	if err != nil {
		return nil, fmt.Errorf("aws-kms setconfig: %w", err)
	}
	return &awsKMS{w: w}, nil
}

func (a *awsKMS) Wrap(ctx context.Context, pt, aad []byte) (WrappedKey, error) {
	blob, err := a.w.Encrypt(ctx, pt, wrapping.WithAad(aad))
	if err != nil {
		return WrappedKey{}, err
	}
	b, err := proto.Marshal(blob)
	if err != nil {
		return WrappedKey{}, err
	}
	return WrappedKey{Version: -1, Ciphertext: b}, nil
}

func (a *awsKMS) Unwrap(ctx context.Context, w WrappedKey, aad []byte) ([]byte, error) {
	blob := &wrapping.BlobInfo{}
	if err := proto.Unmarshal(w.Ciphertext, blob); err != nil {
		return nil, err
	}
	return a.w.Decrypt(ctx, blob, wrapping.WithAad(aad))
}

func (a *awsKMS) CurrentVersion(ctx context.Context) (int, error)   { return -1, nil }
func (a *awsKMS) LoadedVersions(ctx context.Context) ([]int, error) { return nil, nil }
func (a *awsKMS) Close() error                                      { return nil }
