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

package bootstrap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/probod"
	"sigs.k8s.io/yaml"
)

func TestWriteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "probod.yml")

	cfg := &probod.FullConfig{
		Unit: probod.UnitConfig{
			Metrics: probod.MetricsConfig{Addr: "localhost:9090"},
		},
		Probod: probod.Config{
			BaseURL:       "http://localhost:8080",
			EncryptionKey: "test-key",
		},
	}

	err := WriteConfig(cfg, configPath)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var loaded probod.FullConfig
	err = yaml.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, cfg.Unit.Metrics.Addr, loaded.Unit.Metrics.Addr)
	assert.Equal(t, cfg.Probod.BaseURL, loaded.Probod.BaseURL)
	assert.Equal(t, cfg.Probod.EncryptionKey, loaded.Probod.EncryptionKey)
}

func TestWriteConfig_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "dir", "probod.yml")

	cfg := &probod.FullConfig{
		Probod: probod.Config{BaseURL: "http://localhost:8080"},
	}

	err := WriteConfig(cfg, configPath)
	require.NoError(t, err)

	_, err = os.Stat(configPath)
	require.NoError(t, err)
}

func TestWriteConfig_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "probod.yml")

	cfg := &probod.FullConfig{}

	err := WriteConfig(cfg, configPath)
	require.NoError(t, err)

	info, err := os.Stat(configPath)
	require.NoError(t, err)

	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestWriteConfig_CompleteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "probod.yml")

	cfg := &probod.FullConfig{
		Unit: probod.UnitConfig{
			Metrics: probod.MetricsConfig{Addr: "localhost:8081"},
			Tracing: probod.TracingConfig{
				Addr:          "localhost:4317",
				MaxBatchSize:  512,
				BatchTimeout:  5,
				ExportTimeout: 30,
				MaxQueueSize:  2048,
			},
		},
		Probod: probod.Config{
			BaseURL:       "http://localhost:8080",
			EncryptionKey: "test-key",
			ChromeDPAddr:  "localhost:9222",
			Api: probod.APIConfig{
				Addr: ":8080",
				Cors: probod.CorsConfig{
					AllowedOrigins: []string{"http://localhost:8080"},
				},
				ExtraHeaderFields: map[string]string{},
			},
			Pg: probod.PgConfig{
				Addr:     "localhost:5432",
				Username: "postgres",
				Password: "postgres",
				Database: "probod",
				PoolSize: 100,
			},
			Connectors: []probod.ConnectorConfig{
				{
					Provider: "slack",
					Protocol: "oauth2",
					RawConfig: probod.ConnectorConfigOAuth2{
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

	err := WriteConfig(cfg, configPath)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var loaded probod.FullConfig
	err = yaml.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, cfg.Unit.Metrics.Addr, loaded.Unit.Metrics.Addr)
	assert.Equal(t, cfg.Unit.Tracing.MaxBatchSize, loaded.Unit.Tracing.MaxBatchSize)
	assert.Equal(t, cfg.Probod.Api.Cors.AllowedOrigins, loaded.Probod.Api.Cors.AllowedOrigins)
	assert.Equal(t, cfg.Probod.Pg.PoolSize, loaded.Probod.Pg.PoolSize)
	require.Len(t, loaded.Probod.Connectors, 1)
	assert.Equal(t, "SLACK", loaded.Probod.Connectors[0].Provider)
}
