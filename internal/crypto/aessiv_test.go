package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestSealOpen_RoundTrip(t *testing.T) {
	key := mustRand(KeySize)
	plaintext := []byte("alice@example.com")
	aad := []byte("EMAIL:f47ac10b-58cc-4372-a567-0e02b2c3d479")

	ct, err := Seal(key, plaintext, aad)
	if err != nil {
		t.Fatal(err)
	}
	if len(ct) < SIVSize {
		t.Fatalf("ciphertext shorter than SIV size: %d", len(ct))
	}
	pt, err := Open(key, ct, aad)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pt, plaintext) {
		t.Errorf("roundtrip mismatch: got %q want %q", pt, plaintext)
	}
}

func TestSealOpen_AADMismatch(t *testing.T) {
	key := mustRand(KeySize)
	plaintext := []byte("alice@example.com")

	ct, err := Seal(key, plaintext, []byte("EMAIL:uuid-1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(key, ct, []byte("EMAIL:uuid-2")); err == nil {
		t.Fatal("expected AAD mismatch error")
	}
}

func TestSeal_DeterministicSameInputs(t *testing.T) {
	key := mustRand(KeySize)
	plaintext := []byte("repeat")
	aad := []byte("EMAIL:uuid-a")

	ct1, err := Seal(key, plaintext, aad)
	if err != nil {
		t.Fatal(err)
	}
	ct2, err := Seal(key, plaintext, aad)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ct1, ct2) {
		t.Errorf("expected deterministic ciphertext for same (key, aad, pt)")
	}
}

func TestSeal_DiffersByAAD(t *testing.T) {
	key := mustRand(KeySize)
	pt := []byte("alice@example.com")
	c1, _ := Seal(key, pt, []byte("EMAIL:uuid-a"))
	c2, _ := Seal(key, pt, []byte("EMAIL:uuid-b"))
	if bytes.Equal(c1, c2) {
		t.Fatal("same plaintext under different AAD must yield different ciphertext")
	}
}

func TestSeal_RejectsBadKeyLen(t *testing.T) {
	if _, err := Seal(make([]byte, 32), []byte("x"), nil); err == nil {
		t.Fatal("expected key-length error (32-byte key should be rejected; want 64)")
	}
}

func TestOpen_RejectsShortSealed(t *testing.T) {
	key := mustRand(KeySize)
	if _, err := Open(key, []byte("short"), nil); err == nil {
		t.Fatal("expected too-short error")
	}
}

func mustRand(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}
