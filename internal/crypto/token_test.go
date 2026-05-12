package crypto

import (
	"bytes"
	"testing"
)

func TestEncodeDecode_RoundTrip(t *testing.T) {
	body := make([]byte, SIVSize+5)
	for i := range body {
		body[i] = byte(i)
	}
	tok, err := EncodeToken("EMAIL", body)
	if err != nil {
		t.Fatal(err)
	}
	if tok[:6] != "TOKEN_" {
		t.Fatalf("missing prefix: %s", tok)
	}
	if tok[6:11] != "EMAIL" {
		t.Fatalf("wrong type in token: %s", tok[6:11])
	}
	gotType, gotBody, err := DecodeToken(tok)
	if err != nil {
		t.Fatal(err)
	}
	if gotType != "EMAIL" {
		t.Errorf("type: got %q want EMAIL", gotType)
	}
	if !bytes.Equal(gotBody, body) {
		t.Error("body mismatch")
	}
}

func TestEncodeToken_RejectsBadType(t *testing.T) {
	body := make([]byte, SIVSize+8)
	for _, bad := range []string{"", "email", "EMAIL!", "EMAIL ADDR", "123"} {
		if _, err := EncodeToken(bad, body); err == nil {
			t.Errorf("expected error for type %q", bad)
		}
	}
}

func TestDecodeToken_BadPrefix(t *testing.T) {
	if _, _, err := DecodeToken("foo_EMAIL_MFRGG"); err == nil {
		t.Fatal("expected prefix error")
	}
}

func TestDecodeToken_BadBase32(t *testing.T) {
	if _, _, err := DecodeToken("TOKEN_EMAIL_???"); err == nil {
		t.Fatal("expected base32 decode error")
	}
}

func TestDecodeToken_BodyTooShort(t *testing.T) {
	tok, _ := EncodeToken("EMAIL", []byte("short"))
	if _, _, err := DecodeToken(tok); err == nil {
		t.Fatal("expected body-too-short error")
	}
}

func TestDecodeToken_MissingTypeSeparator(t *testing.T) {
	if _, _, err := DecodeToken("TOKEN_EMAILMFRGG"); err == nil {
		t.Fatal("expected separator error")
	}
}
