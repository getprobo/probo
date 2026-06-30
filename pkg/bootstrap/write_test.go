// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package bootstrap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/probodconfig"
	"sigs.k8s.io/yaml"
)

func TestWriteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "probod.yml")

	cfg := &probodconfig.FullConfig{
		Unit: probodconfig.UnitConfig{
			Metrics: probodconfig.MetricsConfig{Addr: "localhost:9090"},
		},
		Probod: probodconfig.Config{
			BaseURL:       "http://localhost:8080",
			EncryptionKey: "test-key",
		},
	}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var loaded probodconfig.FullConfig

	err = yaml.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, cfg.Unit.Metrics.Addr, loaded.Unit.Metrics.Addr)
	assert.Equal(t, cfg.Probod.BaseURL, loaded.Probod.BaseURL)
	assert.Equal(t, cfg.Probod.EncryptionKey, loaded.Probod.EncryptionKey)
}

func TestWriteConfig_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "dir", "probod.yml")

	cfg := &probodconfig.FullConfig{
		Probod: probodconfig.Config{BaseURL: "http://localhost:8080"},
	}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	_, err = os.Stat(configPath)
	require.NoError(t, err)
}

func TestWriteConfig_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "probod.yml")

	cfg := &probodconfig.FullConfig{}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	info, err := os.Stat(configPath)
	require.NoError(t, err)

	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestWriteConfig_DropsEmptyStrings(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "probod.yml")

	cfg := &probodconfig.FullConfig{
		Unit: probodconfig.UnitConfig{
			Metrics: probodconfig.MetricsConfig{Addr: "localhost:9090"},
			Tracing: probodconfig.TracingConfig{Addr: ""},
		},
		Probod: probodconfig.Config{
			BaseURL:       "http://localhost:8080",
			EncryptionKey: "",
			ChromeDPAddr:  "",
			Pg: probodconfig.PgConfig{
				Addr:     "localhost:5432",
				Username: "postgres",
				Password: "",
				Database: "",
				PoolSize: 100,
			},
		},
	}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var tree map[string]any

	err = yaml.Unmarshal(data, &tree)
	require.NoError(t, err)

	probod, ok := tree["probod"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "http://localhost:8080", probod["base-url"])
	assert.NotContains(t, probod, "encryption-key")
	assert.NotContains(t, probod, "chrome-dp-addr")

	pg, ok := probod["pg"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "localhost:5432", pg["addr"])
	assert.Equal(t, "postgres", pg["username"])
	assert.NotContains(t, pg, "password")
	assert.NotContains(t, pg, "database")
	assert.Contains(t, pg, "pool-size")

	unit, ok := tree["unit"].(map[string]any)
	require.True(t, ok)

	metrics, ok := unit["metrics"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "localhost:9090", metrics["addr"])

	tracing, ok := unit["tracing"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, tracing, "addr")

	loaded := probodconfig.FullConfig{}
	err = yaml.Unmarshal(data, &loaded)
	require.NoError(t, err)
	assert.Equal(t, cfg.Probod.BaseURL, loaded.Probod.BaseURL)
	assert.Empty(t, loaded.Probod.EncryptionKey)
}

func TestWriteConfig_CompleteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "probod.yml")

	cfg := &probodconfig.FullConfig{
		Unit: probodconfig.UnitConfig{
			Metrics: probodconfig.MetricsConfig{Addr: "localhost:8081"},
			Tracing: probodconfig.TracingConfig{
				Addr:          "localhost:4317",
				MaxBatchSize:  512,
				BatchTimeout:  5,
				ExportTimeout: 30,
				MaxQueueSize:  2048,
			},
		},
		Probod: probodconfig.Config{
			BaseURL:       "http://localhost:8080",
			EncryptionKey: "test-key",
			ChromeDPAddr:  "localhost:9222",
			Api: probodconfig.APIConfig{
				Addr: ":8080",
				Cors: probodconfig.CorsConfig{
					AllowedOrigins: []string{"http://localhost:8080"},
				},
				ExtraHeaderFields: map[string]string{},
			},
			Pg: probodconfig.PgConfig{
				Addr:                   "localhost:5432",
				Username:               "postgres",
				Password:               "postgres",
				Database:               "probod",
				PoolSize:               100,
				MinPoolSize:            10,
				MaxConnIdleTimeSeconds: 1800,
				MaxConnLifetimeSeconds: 3600,
			},
			Connectors: []probodconfig.ConnectorConfig{
				{
					Provider: "slack",
					Protocol: "oauth2",
					RawConfig: probodconfig.ConnectorConfigOAuth2{
						ClientID:     "client-id",
						ClientSecret: "client-secret",
					},
					RawSettings: map[string]any{
						"signing-secret": "secret",
					},
				},
			},
		},
	}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var loaded probodconfig.FullConfig

	err = yaml.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, cfg.Unit.Metrics.Addr, loaded.Unit.Metrics.Addr)
	assert.Equal(t, cfg.Unit.Tracing.MaxBatchSize, loaded.Unit.Tracing.MaxBatchSize)
	assert.Equal(t, cfg.Probod.Api.Cors.AllowedOrigins, loaded.Probod.Api.Cors.AllowedOrigins)
	assert.Equal(t, cfg.Probod.Pg.PoolSize, loaded.Probod.Pg.PoolSize)
	assert.Equal(t, cfg.Probod.Pg.MinPoolSize, loaded.Probod.Pg.MinPoolSize)
	assert.Equal(t, cfg.Probod.Pg.MaxConnIdleTimeSeconds, loaded.Probod.Pg.MaxConnIdleTimeSeconds)
	assert.Equal(t, cfg.Probod.Pg.MaxConnLifetimeSeconds, loaded.Probod.Pg.MaxConnLifetimeSeconds)
	require.Len(t, loaded.Probod.Connectors, 1)
	assert.Equal(t, "SLACK", loaded.Probod.Connectors[0].Provider)
}

func TestWriteConfig_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "probod.json")

	cfg := &probodconfig.FullConfig{
		Unit: probodconfig.UnitConfig{
			Metrics: probodconfig.MetricsConfig{Addr: "localhost:9090"},
		},
		Probod: probodconfig.Config{
			BaseURL:       "http://localhost:8080",
			EncryptionKey: "",
		},
	}

	err := WriteConfig(cfg, configPath, FormatJSON)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var tree map[string]any

	err = json.Unmarshal(data, &tree)
	require.NoError(t, err)

	probod, ok := tree["probod"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "http://localhost:8080", probod["base-url"])
	assert.NotContains(t, probod, "encryption-key")

	var loaded probodconfig.FullConfig

	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, cfg.Unit.Metrics.Addr, loaded.Unit.Metrics.Addr)
	assert.Equal(t, cfg.Probod.BaseURL, loaded.Probod.BaseURL)
}

func TestWriteConfig_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "probod.txt")

	cfg := &probodconfig.FullConfig{}

	err := WriteConfig(cfg, configPath, Format("toml"))
	require.Error(t, err)
}
