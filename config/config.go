package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultListenPort     = 9717
	DefaultTimeoutSeconds = 10
)

// Switch holds the connection details for a single TP-Link Easy Smart switch.
// Exactly one of Password or PasswordFile must be set, which lets the
// non-secret fields (host, username) live in a ConfigMap while the password is
// mounted separately from a Secret.
type Switch struct {
	Host         string `yaml:"host"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	PasswordFile string `yaml:"password_file"`
}

// Config is the top-level exporter configuration.
type Config struct {
	ListenPort     int      `yaml:"listen_port"`
	TimeoutSeconds int      `yaml:"timeout_seconds"`
	Switches       []Switch `yaml:"switches"`
}

// Load reads and validates the YAML configuration at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config file %q: %w", path, err)
	}
	if err := cfg.resolvePasswordFiles(); err != nil {
		return nil, fmt.Errorf("invalid config file %q: %w", path, err)
	}
	return cfg, nil
}

// resolvePasswordFiles reads each switch's PasswordFile into Password so the
// rest of the program only needs to read Password. Called after validate, which
// has already enforced the Password/PasswordFile XOR.
func (c *Config) resolvePasswordFiles() error {
	for i := range c.Switches {
		if c.Switches[i].PasswordFile == "" {
			continue
		}
		data, err := os.ReadFile(c.Switches[i].PasswordFile)
		if err != nil {
			return fmt.Errorf("switches[%d] (host %q): reading password_file: %w", i, c.Switches[i].Host, err)
		}
		password := strings.TrimSpace(string(data))
		if password == "" {
			return fmt.Errorf("switches[%d] (host %q): password_file %q is empty",
				i, c.Switches[i].Host, c.Switches[i].PasswordFile)
		}
		c.Switches[i].Password = password
	}
	return nil
}

func (c *Config) applyDefaults() {
	if c.ListenPort == 0 {
		c.ListenPort = DefaultListenPort
	}
	if c.TimeoutSeconds == 0 {
		c.TimeoutSeconds = DefaultTimeoutSeconds
	}
	for i := range c.Switches {
		if c.Switches[i].Username == "" {
			c.Switches[i].Username = "admin"
		}
	}
}

func (c *Config) validate() error {
	if len(c.Switches) == 0 {
		return errors.New("at least one switch must be configured under 'switches'")
	}
	for i, sw := range c.Switches {
		if sw.Host == "" {
			return fmt.Errorf("switches[%d]: 'host' is required", i)
		}
		switch {
		case sw.Password == "" && sw.PasswordFile == "":
			return fmt.Errorf("switches[%d] (host %q): one of 'password' or 'password_file' is required", i, sw.Host)
		case sw.Password != "" && sw.PasswordFile != "":
			return fmt.Errorf("switches[%d] (host %q): 'password' and 'password_file' are mutually exclusive", i, sw.Host)
		}
	}
	if c.TimeoutSeconds < 0 {
		return fmt.Errorf("'timeout_seconds' must not be negative, got %d", c.TimeoutSeconds)
	}
	return nil
}
