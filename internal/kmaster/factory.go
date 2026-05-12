package kmaster

import (
	"context"
	"fmt"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
)

func Open(ctx context.Context, cfg *config.Config) (Wrapper, error) {
	switch cfg.KMasterBackend {
	case "aead-local":
		return NewAEADLocal(ctx, cfg)
	case "vault-transit":
		return NewVaultTransit(ctx, cfg)
	case "aws-kms":
		return NewAWSKMS(ctx, cfg)
	case "pkcs11":
		return NewPKCS11(ctx, cfg)
	default:
		return nil, fmt.Errorf("unknown KMasterBackend: %s", cfg.KMasterBackend)
	}
}
