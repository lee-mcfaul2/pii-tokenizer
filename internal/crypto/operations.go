package crypto

func Tokenize(kreq []byte, uuid, piiType string, plaintext []byte) (string, error) {
	aad := []byte(piiType + ":" + uuid)
	sealed, err := Seal(kreq, plaintext, aad)
	if err != nil {
		return "", err
	}
	return EncodeToken(piiType, sealed)
}

func Detokenize(kreq []byte, uuid, token string) (piiType string, plaintext []byte, err error) {
	piiType, sealed, err := DecodeToken(token)
	if err != nil {
		return "", nil, err
	}
	aad := []byte(piiType + ":" + uuid)
	pt, err := Open(kreq, sealed, aad)
	if err != nil {
		return "", nil, err
	}
	return piiType, pt, nil
}
