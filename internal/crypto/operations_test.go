package crypto

import (
	"bytes"
	"testing"
)

func TestTokenizeDetokenize_RoundTrip(t *testing.T) {
	kreq := mustRand(KeySize)
	uuid := "f47ac10b-58cc-4372-a567-0e02b2c3d479"

	tok, err := Tokenize(kreq, uuid, "EMAIL", []byte("alice@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	gotType, pt, err := Detokenize(kreq, uuid, tok)
	if err != nil {
		t.Fatal(err)
	}
	if gotType != "EMAIL" {
		t.Errorf("type: got %s", gotType)
	}
	if !bytes.Equal(pt, []byte("alice@example.com")) {
		t.Errorf("plaintext mismatch: %q", pt)
	}
}

func TestTokenize_DeterministicWithinRequest(t *testing.T) {
	kreq := mustRand(KeySize)
	uuid := "u"
	a, _ := Tokenize(kreq, uuid, "EMAIL", []byte("x@y"))
	b, _ := Tokenize(kreq, uuid, "EMAIL", []byte("x@y"))
	if a != b {
		t.Error("same plaintext within request must produce identical token")
	}
}

func TestTokenize_DiffersAcrossRequests(t *testing.T) {
	k1 := mustRand(KeySize)
	k2 := mustRand(KeySize)
	a, _ := Tokenize(k1, "uuid-a", "EMAIL", []byte("x@y"))
	b, _ := Tokenize(k2, "uuid-b", "EMAIL", []byte("x@y"))
	if a == b {
		t.Error("same plaintext across different (kreq,uuid) must NOT produce same token")
	}
}

func TestDetokenize_RejectsWrongUUID(t *testing.T) {
	kreq := mustRand(KeySize)
	tok, _ := Tokenize(kreq, "uuid-1", "EMAIL", []byte("x"))
	if _, _, err := Detokenize(kreq, "uuid-2", tok); err == nil {
		t.Fatal("expected AAD mismatch")
	}
}

func TestDetokenize_RejectsTamperedToken(t *testing.T) {
	kreq := mustRand(KeySize)
	tok, _ := Tokenize(kreq, "uuid-1", "EMAIL", []byte("x"))
	tampered := []byte(tok)
	tampered[len(tok)-1] ^= 0x01
	if _, _, err := Detokenize(kreq, "uuid-1", string(tampered)); err == nil {
		t.Fatal("expected error on tampered token")
	}
}

func TestDetokenize_RejectsWrongKey(t *testing.T) {
	k1 := mustRand(KeySize)
	k2 := mustRand(KeySize)
	tok, _ := Tokenize(k1, "u", "EMAIL", []byte("x"))
	if _, _, err := Detokenize(k2, "u", tok); err == nil {
		t.Fatal("expected error with wrong K_request")
	}
}

func TestTokenize_EmptyPlaintext(t *testing.T) {
	kreq := mustRand(KeySize)
	tok, err := Tokenize(kreq, "u", "EMAIL", []byte(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(tok) < len("TOKEN_EMAIL_") {
		t.Fatal("empty plaintext should still produce a token")
	}
	_, pt, err := Detokenize(kreq, "u", tok)
	if err != nil {
		t.Fatal(err)
	}
	if len(pt) != 0 {
		t.Errorf("expected empty plaintext on round-trip, got %q", pt)
	}
}
