package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHighestVersion(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "k_master_v1"), []byte("x"), 0o600)
	os.WriteFile(filepath.Join(dir, "k_master_v2"), []byte("x"), 0o600)
	os.WriteFile(filepath.Join(dir, "k_master_v10"), []byte("x"), 0o600)
	os.WriteFile(filepath.Join(dir, "current"), []byte("v2"), 0o600)

	h, err := highestVersion(dir)
	if err != nil {
		t.Fatal(err)
	}
	if h != 10 {
		t.Errorf("highest: got %d want 10", h)
	}
}

func TestHighestVersion_EmptyDir(t *testing.T) {
	if _, err := highestVersion(t.TempDir()); err == nil {
		t.Fatal("expected error on empty dir")
	}
}
