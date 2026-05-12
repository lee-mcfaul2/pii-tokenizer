package kmaster

import (
	"context"
	"fmt"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
)

// PKCS#11 backend is deferred: hashicorp/go-kms-wrapping does not publish a
// pkcs11 wrapper module on the Go proxy, and our threat model lets operators
// reach for vault-transit or aws-kms (CloudHSM) for HSM-grade key custody
// without taking on a direct PKCS#11 integration here.
//
// To enable this backend, replace this constructor with one that wires
// a real PKCS#11 library (e.g., github.com/miekg/pkcs11) into a Wrapper.
func NewPKCS11(_ context.Context, _ *config.Config) (Wrapper, error) {
	return nil, fmt.Errorf("pkcs11 backend not built into this binary; use vault-transit or aws-kms")
}
