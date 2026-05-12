package kmaster

import (
	"context"
	"fmt"

	wrapping "github.com/hashicorp/go-kms-wrapping/v2"
	transit "github.com/hashicorp/go-kms-wrapping/wrappers/transit/v2"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
	"google.golang.org/protobuf/proto"
)

type vaultTransit struct {
	w *transit.Wrapper
}

func NewVaultTransit(ctx context.Context, cfg *config.Config) (Wrapper, error) {
	if cfg.VaultAddr == "" {
		return nil, fmt.Errorf("VAULT_ADDR not set")
	}
	if cfg.VaultKeyName == "" {
		return nil, fmt.Errorf("vault transit key name not set")
	}
	w := transit.NewWrapper()
	_, err := w.SetConfig(ctx,
		wrapping.WithConfigMap(map[string]string{
			"address":    cfg.VaultAddr,
			"key_name":   cfg.VaultKeyName,
			"mount_path": "transit",
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("vault transit setconfig: %w", err)
	}
	return &vaultTransit{w: w}, nil
}

func (v *vaultTransit) Wrap(ctx context.Context, pt, aad []byte) (WrappedKey, error) {
	blob, err := v.w.Encrypt(ctx, pt, wrapping.WithAad(aad))
	if err != nil {
		return WrappedKey{}, err
	}
	b, err := proto.Marshal(blob)
	if err != nil {
		return WrappedKey{}, err
	}
	version := 0
	if blob.KeyInfo != nil {
		fmt.Sscanf(blob.KeyInfo.KeyId, "v%d", &version)
	}
	return WrappedKey{Version: version, Ciphertext: b}, nil
}

func (v *vaultTransit) Unwrap(ctx context.Context, w WrappedKey, aad []byte) ([]byte, error) {
	blob := &wrapping.BlobInfo{}
	if err := proto.Unmarshal(w.Ciphertext, blob); err != nil {
		return nil, err
	}
	return v.w.Decrypt(ctx, blob, wrapping.WithAad(aad))
}

func (v *vaultTransit) CurrentVersion(ctx context.Context) (int, error)  { return -1, nil }
func (v *vaultTransit) LoadedVersions(ctx context.Context) ([]int, error) { return nil, nil }
func (v *vaultTransit) Close() error                                       { return nil }
