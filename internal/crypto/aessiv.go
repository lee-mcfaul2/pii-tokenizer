package crypto

import (
	"fmt"

	siv "github.com/secure-io/siv-go"
)

const (
	KeySize = 64
	SIVSize = 16
)

func Seal(key, plaintext, aad []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("key must be %d bytes, got %d", KeySize, len(key))
	}
	a, err := siv.NewCMAC(key)
	if err != nil {
		return nil, fmt.Errorf("init aes-siv-cmac: %w", err)
	}
	return a.Seal(nil, nil, plaintext, aad), nil
}

func Open(key, sealed, aad []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("key must be %d bytes, got %d", KeySize, len(key))
	}
	if len(sealed) < SIVSize {
		return nil, fmt.Errorf("sealed payload too short: %d < %d", len(sealed), SIVSize)
	}
	a, err := siv.NewCMAC(key)
	if err != nil {
		return nil, fmt.Errorf("init aes-siv-cmac: %w", err)
	}
	return a.Open(nil, nil, sealed, aad)
}
