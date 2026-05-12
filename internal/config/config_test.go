package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ListenAddr != ":8443" || c.RedisMode != "single" || c.KMasterBackend != "aead-local" {
		t.Errorf("unexpected defaults: %+v", c)
	}
	if c.RedisTimeout != 250*time.Millisecond {
		t.Errorf("wrong redis timeout: %v", c.RedisTimeout)
	}
}

func TestLoad_InvalidBackend(t *testing.T) {
	clearEnv(t)
	t.Setenv("TOKENIZER_KMASTER_BACKEND", "bogus")
	if _, err := Load(); err == nil {
		t.Fatal("expected validation error for invalid backend")
	}
}

func TestLoad_InvalidRedisMode(t *testing.T) {
	clearEnv(t)
	t.Setenv("TOKENIZER_REDIS_MODE", "bogus")
	if _, err := Load(); err == nil {
		t.Fatal("expected validation error for invalid redis mode")
	}
}

func TestLoad_OverrideListen(t *testing.T) {
	clearEnv(t)
	t.Setenv("TOKENIZER_LISTEN_ADDR", ":9999")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ListenAddr != ":9999" {
		t.Errorf("override failed: %s", c.ListenAddr)
	}
}

func TestLoad_DurationOverride(t *testing.T) {
	clearEnv(t)
	t.Setenv("TOKENIZER_READ_TIMEOUT", "5s")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ReadTimeout != 5*time.Second {
		t.Errorf("duration override failed: %v", c.ReadTimeout)
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"TOKENIZER_LISTEN_ADDR",
		"TOKENIZER_REDIS_MODE",
		"TOKENIZER_KMASTER_BACKEND",
		"TOKENIZER_READ_TIMEOUT",
		"TOKENIZER_WRITE_TIMEOUT",
	} {
		t.Setenv(k, "")
	}
}
