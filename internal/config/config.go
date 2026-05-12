package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// HTTP server
	ListenAddr      string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	// Redis
	RedisMode     string
	RedisAddrs    string
	RedisDB       int
	RedisPassword string
	RedisPoolSize int
	RedisTimeout  time.Duration

	// K_master backend
	KMasterBackend string

	// aead-local backend specifics
	AEADKeyPath string

	// vault-transit backend specifics
	VaultAddr    string
	VaultRole    string
	VaultKeyName string

	// aws-kms backend specifics
	AWSRegion string
	AWSKeyID  string

	// pkcs11 backend specifics
	PKCS11Lib   string
	PKCS11Slot  string
	PKCS11Pin   string
	PKCS11Label string

	// Telemetry
	OTLPEndpoint string
	LogLevel     string

	// Server identity
	ServiceName string
}

func Load() (*Config, error) {
	c := &Config{
		ListenAddr:      getenv("TOKENIZER_LISTEN_ADDR", ":8443"),
		ReadTimeout:     getdur("TOKENIZER_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getdur("TOKENIZER_WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimeout: getdur("TOKENIZER_SHUTDOWN_TIMEOUT", 30*time.Second),
		RedisMode:       getenv("TOKENIZER_REDIS_MODE", "single"),
		RedisAddrs:      getenv("TOKENIZER_REDIS_ADDRS", "localhost:6379"),
		RedisDB:         getint("TOKENIZER_REDIS_DB", 0),
		RedisPassword:   getenv("TOKENIZER_REDIS_PASSWORD", ""),
		RedisPoolSize:   getint("TOKENIZER_REDIS_POOLSIZE", 50),
		RedisTimeout:    getdur("TOKENIZER_REDIS_TIMEOUT", 250*time.Millisecond),
		KMasterBackend:  getenv("TOKENIZER_KMASTER_BACKEND", "aead-local"),
		AEADKeyPath:     getenv("TOKENIZER_AEAD_KEY_PATH", "/etc/tokenizer/kmaster"),
		VaultAddr:       getenv("VAULT_ADDR", ""),
		VaultRole:       getenv("VAULT_ROLE", ""),
		VaultKeyName:    getenv("VAULT_TRANSIT_KEY", "pii-tokenizer"),
		AWSRegion:       getenv("AWS_REGION", ""),
		AWSKeyID:        getenv("TOKENIZER_AWS_KMS_KEY_ID", ""),
		PKCS11Lib:       getenv("TOKENIZER_PKCS11_LIB", ""),
		PKCS11Slot:      getenv("TOKENIZER_PKCS11_SLOT", ""),
		PKCS11Pin:       getenv("TOKENIZER_PKCS11_PIN", ""),
		PKCS11Label:     getenv("TOKENIZER_PKCS11_LABEL", "kmaster"),
		OTLPEndpoint:    getenv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		LogLevel:        getenv("TOKENIZER_LOG_LEVEL", "info"),
		ServiceName:     getenv("OTEL_SERVICE_NAME", "pii-tokenizer"),
	}

	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) validate() error {
	switch c.RedisMode {
	case "single", "sentinel", "cluster":
	default:
		return fmt.Errorf("invalid TOKENIZER_REDIS_MODE: %s", c.RedisMode)
	}
	switch c.KMasterBackend {
	case "aead-local", "vault-transit", "aws-kms", "pkcs11":
	default:
		return fmt.Errorf("invalid TOKENIZER_KMASTER_BACKEND: %s", c.KMasterBackend)
	}
	return nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getint(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getdur(k string, def time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
