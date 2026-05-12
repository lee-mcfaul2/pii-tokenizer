package crypto

import (
	"fmt"

	tinksiv "github.com/google/tink/go/daead/subtle"
)

const (
	KeySize = 64
	SIVSize = 16
)

func Seal(key, plaintext, aad []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("key must be %d bytes, got %d", KeySize, len(key))
	}
	a, err := tinksiv.NewAESSIV(key)
	if err != nil {
		return nil, fmt.Errorf("init aes-siv-cmac: %w", err)
	}
	return a.EncryptDeterministically(plaintext, aad)
}

func Open(key, sealed, aad []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("key must be %d bytes, got %d", KeySize, len(key))
	}
	if len(sealed) < SIVSize {
		return nil, fmt.Errorf("sealed payload too short: %d < %d", len(sealed), SIVSize)
	}
	a, err := tinksiv.NewAESSIV(key)
	if err != nil {
		return nil, fmt.Errorf("init aes-siv-cmac: %w", err)
	}
	return a.DecryptDeterministically(sealed, aad)
}
