package config

import (
	"strings"
	"testing"
)

func validConfig() Config {
	return Config{
		UUID:        "47be59a9-103b-4065-9c3a-ebb9ecf8ffd3",
		PrivateKey:  FixedPrivateKey,
		ShortID:     "0123456789abcdef",
		ServerName:  FixedServerName,
		Target:      FixedTarget,
		Path:        "/reality",
		BridgeToken: strings.Repeat("a", 64),
		Port:        "8080",
		MaxSessions: 256,
	}
}

func TestConfigValidate(t *testing.T) {
	if err := validConfig().Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
}

func TestConfigRejectsInvalidShortID(t *testing.T) {
	cfg := validConfig()
	cfg.ShortID = "abc"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected odd-length short ID to fail")
	}
}

func TestConfigRejectsInvalidPrivateKey(t *testing.T) {
	cfg := validConfig()
	cfg.PrivateKey = "not-a-key"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid private key to fail")
	}
}
