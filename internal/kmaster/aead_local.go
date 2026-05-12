package kmaster

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	wrapping "github.com/hashicorp/go-kms-wrapping/v2"
	aeadw "github.com/hashicorp/go-kms-wrapping/v2/aead"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
	"google.golang.org/protobuf/proto"
)

type aeadLocal struct {
	dir string

	mu       sync.RWMutex
	current  int
	wrappers map[int]*aeadw.Wrapper
}

func NewAEADLocal(ctx context.Context, cfg *config.Config) (Wrapper, error) {
	a := &aeadLocal{
		dir:      cfg.AEADKeyPath,
		wrappers: map[int]*aeadw.Wrapper{},
	}
	if err := a.reload(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *aeadLocal) reload(ctx context.Context) error {
	entries, err := os.ReadDir(a.dir)
	if err != nil {
		return fmt.Errorf("read kmaster dir %s: %w", a.dir, err)
	}

	loaded := map[int]*aeadw.Wrapper{}
	current := 0

	for _, e := range entries {
		name := e.Name()
		switch {
		case strings.HasPrefix(name, "k_master_v"):
			v, err := strconv.Atoi(strings.TrimPrefix(name, "k_master_v"))
			if err != nil {
				continue
			}
			keyBytes, err := os.ReadFile(filepath.Join(a.dir, name))
			if err != nil {
				return fmt.Errorf("read %s: %w", name, err)
			}
			if len(keyBytes) != 32 {
				return fmt.Errorf("%s: expected 32 bytes, got %d", name, len(keyBytes))
			}
			w := aeadw.NewWrapper()
			if _, err := w.SetConfig(ctx, wrapping.WithKeyId(fmt.Sprintf("v%d", v))); err != nil {
				return err
			}
			if err := w.SetAesGcmKeyBytes(keyBytes); err != nil {
				return err
			}
			loaded[v] = w
		case name == "current":
			b, err := os.ReadFile(filepath.Join(a.dir, name))
			if err != nil {
				return fmt.Errorf("read current: %w", err)
			}
			current, err = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(string(b), "v")))
			if err != nil {
				return fmt.Errorf("parse current: %w", err)
			}
		}
	}

	if len(loaded) == 0 {
		return errors.New("no K_master versions found")
	}
	if current == 0 {
		for v := range loaded {
			if v > current {
				current = v
			}
		}
	}
	if _, ok := loaded[current]; !ok {
		return fmt.Errorf("current=%d but version not loaded", current)
	}

	a.mu.Lock()
	a.current = current
	a.wrappers = loaded
	a.mu.Unlock()
	return nil
}

func (a *aeadLocal) Wrap(ctx context.Context, plaintext, aad []byte) (WrappedKey, error) {
	a.mu.RLock()
	current := a.current
	w := a.wrappers[current]
	a.mu.RUnlock()

	if w == nil {
		return WrappedKey{}, fmt.Errorf("no current wrapper")
	}
	blob, err := w.Encrypt(ctx, plaintext, wrapping.WithAad(aad))
	if err != nil {
		return WrappedKey{}, err
	}
	ctBytes, err := proto.Marshal(blob)
	if err != nil {
		return WrappedKey{}, err
	}
	return WrappedKey{Version: current, Ciphertext: ctBytes}, nil
}

func (a *aeadLocal) Unwrap(ctx context.Context, wrapped WrappedKey, aad []byte) ([]byte, error) {
	a.mu.RLock()
	w, ok := a.wrappers[wrapped.Version]
	a.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("version %d not loaded", wrapped.Version)
	}
	blob := &wrapping.BlobInfo{}
	if err := proto.Unmarshal(wrapped.Ciphertext, blob); err != nil {
		return nil, err
	}
	return w.Decrypt(ctx, blob, wrapping.WithAad(aad))
}

func (a *aeadLocal) CurrentVersion(ctx context.Context) (int, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.current == 0 {
		return 0, errors.New("no current version")
	}
	return a.current, nil
}

func (a *aeadLocal) LoadedVersions(ctx context.Context) ([]int, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]int, 0, len(a.wrappers))
	for v := range a.wrappers {
		out = append(out, v)
	}
	return out, nil
}

func (a *aeadLocal) Close() error { return nil }

func GenerateLocalKey() ([]byte, error) {
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		return nil, err
	}
	return k, nil
}
