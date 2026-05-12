package crypto

import (
	"encoding/base32"
	"fmt"
	"strings"
)

const TokenPrefix = "TOKEN_"

var base32Codec = base32.StdEncoding.WithPadding(base32.NoPadding)

func EncodeToken(piiType string, body []byte) (string, error) {
	if err := validateType(piiType); err != nil {
		return "", err
	}
	return TokenPrefix + piiType + "_" + base32Codec.EncodeToString(body), nil
}

func DecodeToken(token string) (piiType string, body []byte, err error) {
	if !strings.HasPrefix(token, TokenPrefix) {
		return "", nil, fmt.Errorf("token missing prefix")
	}
	rest := token[len(TokenPrefix):]
	sep := strings.Index(rest, "_")
	if sep < 1 {
		return "", nil, fmt.Errorf("token missing type separator")
	}
	piiType = rest[:sep]
	if err := validateType(piiType); err != nil {
		return "", nil, err
	}
	encoded := rest[sep+1:]
	body, err = base32Codec.DecodeString(strings.ToUpper(encoded))
	if err != nil {
		return "", nil, fmt.Errorf("base32 decode: %w", err)
	}
	if len(body) < SIVSize {
		return "", nil, fmt.Errorf("token body too short: %d bytes", len(body))
	}
	return piiType, body, nil
}

func validateType(t string) error {
	if t == "" {
		return fmt.Errorf("type cannot be empty")
	}
	for _, r := range t {
		if !((r >= 'A' && r <= 'Z') || r == '_') {
			return fmt.Errorf("type must be uppercase ASCII letters or underscore: %q", t)
		}
	}
	return nil
}
