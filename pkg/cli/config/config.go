// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"

	"gopkg.in/yaml.v3"
)

const DefaultHTTPTimeout = 30 * time.Second

type (
	Config struct {
		Editor      string                 `yaml:"editor,omitempty"`
		Browser     string                 `yaml:"browser,omitempty"`
		Pager       string                 `yaml:"pager,omitempty"`
		Prompt      string                 `yaml:"prompt,omitempty"`
		HTTPTimeout string                 `yaml:"http_timeout,omitempty"`
		ActiveHost  string                 `yaml:"active_host,omitempty"`
		Hosts       map[string]*HostConfig `yaml:"hosts"`
	}

	HostConfig struct {
		Token        string `yaml:"token"`
		Organization string `yaml:"organization"`
	}
)

var ValidKeys = []string{
	"editor",
	"browser",
	"pager",
	"prompt",
	"http_timeout",
}

func (c *Config) Get(key string) (string, error) {
	switch key {
	case "editor":
		return c.Editor, nil
	case "browser":
		return c.Browser, nil
	case "pager":
		return c.Pager, nil
	case "prompt":
		return c.Prompt, nil
	case "http_timeout":
		return c.HTTPTimeout, nil
	default:
		return "", fmt.Errorf("unknown configuration key: %s", key)
	}
}

func (c *Config) Set(key, value string) error {
	switch key {
	case "editor":
		c.Editor = value
	case "browser":
		c.Browser = value
	case "pager":
		c.Pager = value
	case "prompt":
		if value != "enabled" && value != "disabled" {
			return fmt.Errorf("valid values for prompt are 'enabled' or 'disabled'")
		}
		c.Prompt = value
	case "http_timeout":
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("invalid duration for http_timeout: %w", err)
		}
		c.HTTPTimeout = value
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}
	return nil
}

func (c *Config) HTTPTimeoutDuration() time.Duration {
	if c.HTTPTimeout == "" {
		return DefaultHTTPTimeout
	}

	d, err := time.ParseDuration(c.HTTPTimeout)
	if err != nil {
		return DefaultHTTPTimeout
	}

	return d
}

func configDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config directory: %w", err)
	}
	return filepath.Join(dir, "proboctl"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Hosts: make(map[string]*HostConfig)}, nil
		}
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config file: %w", err)
	}

	if cfg.Hosts == nil {
		cfg.Hosts = make(map[string]*HostConfig)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}

	return nil
}

func (c *Config) DefaultHost() (string, *HostConfig, error) {
	if host := os.Getenv("PROBO_HOST"); host != "" {
		hc := &HostConfig{}
		if saved, ok := c.Hosts[host]; ok {
			*hc = *saved
		}
		if token := os.Getenv("PROBO_TOKEN"); token != "" {
			hc.Token = token
		}
		return host, hc, nil
	}

	hosts := slices.Sorted(maps.Keys(c.Hosts))

	if token := os.Getenv("PROBO_TOKEN"); token != "" {
		if len(hosts) == 0 {
			return "", nil, fmt.Errorf("PROBO_TOKEN is set but no host configured; run 'proboctl auth login' first")
		}

		host := hosts[0]
		if c.ActiveHost != "" {
			if _, ok := c.Hosts[c.ActiveHost]; ok {
				host = c.ActiveHost
			}
		}

		return host, &HostConfig{
			Token:        token,
			Organization: c.Hosts[host].Organization,
		}, nil
	}

	if c.ActiveHost != "" {
		if hc, ok := c.Hosts[c.ActiveHost]; ok {
			return c.ActiveHost, hc, nil
		}
	}

	if len(hosts) > 0 {
		host := hosts[0]
		return host, c.Hosts[host], nil
	}

	return "", nil, fmt.Errorf("not logged in; run 'proboctl auth login' first")
}
