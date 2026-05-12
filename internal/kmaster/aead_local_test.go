package kmaster

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
)

func makeKeyDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	k1 := make([]byte, 32)
	if _, err := rand.Read(k1); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "k_master_v1"), k1, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "current"), []byte("v1"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAEADLocal_WrapUnwrap(t *testing.T) {
	dir := makeKeyDir(t)
	w, err := NewAEADLocal(context.Background(), &config.Config{AEADKeyPath: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	plaintext := []byte("a-K_request-here-but-this-is-fake")
	aad := []byte("uuid-a")

	wrapped, err := w.Wrap(context.Background(), plaintext, aad)
	if err != nil {
		t.Fatal(err)
	}
	if wrapped.Version != 1 {
		t.Errorf("version: got %d want 1", wrapped.Version)
	}

	got, err := w.Unwrap(context.Background(), wrapped, aad)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plaintext) {
		t.Errorf("plaintext mismatch")
	}
}

func TestAEADLocal_UnwrapWrongAAD(t *testing.T) {
	dir := makeKeyDir(t)
	w, err := NewAEADLocal(context.Background(), &config.Config{AEADKeyPath: dir})
	if err != nil {
		t.Fatal(err)
	}
	wrapped, _ := w.Wrap(context.Background(), []byte("x"), []byte("aad-a"))
	if _, err := w.Unwrap(context.Background(), wrapped, []byte("aad-b")); err == nil {
		t.Fatal("expected AAD mismatch")
	}
}

func TestAEADLocal_MultipleVersionsLoaded(t *testing.T) {
	dir := t.TempDir()
	for v := 1; v <= 3; v++ {
		k := make([]byte, 32)
		if _, err := rand.Read(k); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("k_master_v%d", v)), k, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "current"), []byte("v2"), 0o600); err != nil {
		t.Fatal(err)
	}
	w, err := NewAEADLocal(context.Background(), &config.Config{AEADKeyPath: dir})
	if err != nil {
		t.Fatal(err)
	}
	if cv, _ := w.CurrentVersion(context.Background()); cv != 2 {
		t.Errorf("current: got %d want 2", cv)
	}
	versions, _ := w.LoadedVersions(context.Background())
	if len(versions) != 3 {
		t.Errorf("versions: got %d want 3", len(versions))
	}
}

func TestAEADLocal_NoVersionsFound(t *testing.T) {
	dir := t.TempDir()
	if _, err := NewAEADLocal(context.Background(), &config.Config{AEADKeyPath: dir}); err == nil {
		t.Fatal("expected error for empty dir")
	}
}

func TestAEADLocal_RejectsWrongKeyLength(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "k_master_v1"), []byte("too-short"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewAEADLocal(context.Background(), &config.Config{AEADKeyPath: dir}); err == nil {
		t.Fatal("expected key-length error")
	}
}

func TestGenerateLocalKey_32Bytes(t *testing.T) {
	k, err := GenerateLocalKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(k) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(k))
	}
}
