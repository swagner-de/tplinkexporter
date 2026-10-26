package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return path
}

func TestLoadValid(t *testing.T) {
	path := writeTempConfig(t, `
listen_port: 9100
timeout_seconds: 5
switches:
  - host: 192.0.2.10
    username: admin
    password: secret
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.ListenPort != 9100 {
		t.Errorf("ListenPort = %d, want 9100", cfg.ListenPort)
	}
	if cfg.TimeoutSeconds != 5 {
		t.Errorf("TimeoutSeconds = %d, want 5", cfg.TimeoutSeconds)
	}
	if len(cfg.Switches) != 1 {
		t.Fatalf("len(Switches) = %d, want 1", len(cfg.Switches))
	}
	if cfg.Switches[0].Host != "192.0.2.10" {
		t.Errorf("Host = %q, want 192.0.2.10", cfg.Switches[0].Host)
	}
}

func TestLoadDefaults(t *testing.T) {
	path := writeTempConfig(t, `
switches:
  - host: 10.0.0.3
    password: secret
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.ListenPort != DefaultListenPort {
		t.Errorf("ListenPort = %d, want default %d", cfg.ListenPort, DefaultListenPort)
	}
	if cfg.TimeoutSeconds != DefaultTimeoutSeconds {
		t.Errorf("TimeoutSeconds = %d, want default %d", cfg.TimeoutSeconds, DefaultTimeoutSeconds)
	}
	if cfg.Switches[0].Username != "admin" {
		t.Errorf("Username = %q, want default admin", cfg.Switches[0].Username)
	}
}

func TestLoadMultipleSwitches(t *testing.T) {
	path := writeTempConfig(t, `
switches:
  - host: 10.0.0.3
    password: a
  - host: 10.0.0.4
    username: operator
    password: b
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Switches) != 2 {
		t.Fatalf("len(Switches) = %d, want 2", len(cfg.Switches))
	}
	if cfg.Switches[1].Username != "operator" {
		t.Errorf("Switches[1].Username = %q, want operator", cfg.Switches[1].Username)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadMalformed(t *testing.T) {
	path := writeTempConfig(t, "switches: [this is : not : valid")
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestLoadNoSwitches(t *testing.T) {
	path := writeTempConfig(t, "listen_port: 9717\n")
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for empty switches, got nil")
	}
}

func TestLoadMissingHost(t *testing.T) {
	path := writeTempConfig(t, `
switches:
  - password: secret
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing host, got nil")
	}
}

func TestLoadMissingPassword(t *testing.T) {
	path := writeTempConfig(t, `
switches:
  - host: 10.0.0.3
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing password, got nil")
	}
}
