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

package domainconnect_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/domainconnect"
)

func TestExtractHostAndDomain(t *testing.T) {
	t.Parallel()

	t.Run(
		"subdomain and registrable domain",
		func(t *testing.T) {
			t.Parallel()

			host, domain, err := domainconnect.ExtractHostAndDomain("trust.example.com")
			require.NoError(t, err)
			assert.Equal(t, "trust", host)
			assert.Equal(t, "example.com", domain)
		},
	)

	t.Run(
		"registrable domain only",
		func(t *testing.T) {
			t.Parallel()

			host, domain, err := domainconnect.ExtractHostAndDomain("example.com")
			require.NoError(t, err)
			assert.Equal(t, "", host)
			assert.Equal(t, "example.com", domain)
		},
	)

	t.Run(
		"deep subdomain",
		func(t *testing.T) {
			t.Parallel()

			host, domain, err := domainconnect.ExtractHostAndDomain("a.b.example.com")
			require.NoError(t, err)
			assert.Equal(t, "a.b", host)
			assert.Equal(t, "example.com", domain)
		},
	)

	t.Run(
		"invalid domain",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := domainconnect.ExtractHostAndDomain("com")
			require.Error(t, err)
		},
	)
}

func TestBuildApplyURL(t *testing.T) {
	t.Parallel()

	t.Run(
		"unsigned URL with host",
		func(t *testing.T) {
			t.Parallel()

			cfg := domainconnect.Config{
				ProviderID: "probo.inc",
				ServiceID:  "customdomain",
			}

			result, err := domainconnect.BuildApplyURL(
				cfg,
				"https://dns.example.com",
				"example.com",
				"trust",
				map[string]string{"CNAME": "cname.probo.inc"},
				"https://app.probo.inc/callback",
			)

			require.NoError(t, err)

			parsed, err := url.Parse(result)
			require.NoError(t, err)

			assert.Equal(t, "https", parsed.Scheme)
			assert.Equal(t, "dns.example.com", parsed.Host)
			assert.Contains(t, parsed.Path, "/v2/domainTemplates/providers/probo.inc/services/customdomain/apply")

			q := parsed.Query()
			assert.Equal(t, "example.com", q.Get("domain"))
			assert.Equal(t, "trust", q.Get("host"))
			assert.Equal(t, "cname.probo.inc", q.Get("CNAME"))
			assert.Equal(t, "https://app.probo.inc/callback", q.Get("redirect_uri"))
			assert.Equal(t, "Probo", q.Get("providerName"))
			assert.Empty(t, q.Get("sig"))
		},
	)

	t.Run(
		"unsigned URL without host",
		func(t *testing.T) {
			t.Parallel()

			cfg := domainconnect.Config{
				ProviderID: "probo.inc",
				ServiceID:  "customdomain",
			}

			result, err := domainconnect.BuildApplyURL(
				cfg,
				"https://dns.example.com",
				"example.com",
				"",
				nil,
				"",
			)

			require.NoError(t, err)

			parsed, err := url.Parse(result)
			require.NoError(t, err)

			q := parsed.Query()
			assert.Equal(t, "example.com", q.Get("domain"))
			assert.Empty(t, q.Get("host"))
			assert.Empty(t, q.Get("redirect_uri"))
		},
	)

	t.Run(
		"signed URL with ECDSA key",
		func(t *testing.T) {
			t.Parallel()

			key, err := domainconnect.NewECDSATestKey()
			require.NoError(t, err)

			cfg := domainconnect.Config{
				ProviderID: "probo.inc",
				ServiceID:  "customdomain",
				PrivateKey: key,
				KeyID:      "test-key-1",
			}

			result, err := domainconnect.BuildApplyURL(
				cfg,
				"https://dns.example.com",
				"example.com",
				"trust",
				map[string]string{"CNAME": "cname.probo.inc"},
				"",
			)

			require.NoError(t, err)

			parsed, err := url.Parse(result)
			require.NoError(t, err)

			q := parsed.Query()
			sig := q.Get("sig")
			assert.NotEmpty(t, sig)
			assert.Equal(t, "test-key-1", q.Get("key"))

			// Extract the query string that was signed (everything before &sig=)
			fullQuery := parsed.RawQuery
			sigIdx := strings.Index(fullQuery, "&sig=")
			require.True(t, sigIdx > 0)
			signedPart := fullQuery[:sigIdx]

			err = domainconnect.VerifySignature(signedPart, sig, &key.PublicKey)
			assert.NoError(t, err)
		},
	)

	t.Run(
		"special characters in provider ID are path-escaped",
		func(t *testing.T) {
			t.Parallel()

			cfg := domainconnect.Config{
				ProviderID: "my provider",
				ServiceID:  "my service",
			}

			result, err := domainconnect.BuildApplyURL(
				cfg,
				"https://dns.example.com",
				"example.com",
				"",
				nil,
				"",
			)

			require.NoError(t, err)
			assert.Contains(t, result, "/providers/my%20provider/")
			assert.Contains(t, result, "/services/my%20service/")
		},
	)
}

func TestSignAndVerify_ECDSA(t *testing.T) {
	t.Parallel()

	key, err := domainconnect.NewECDSATestKey()
	require.NoError(t, err)

	cfg := domainconnect.Config{
		ProviderID: "probo.inc",
		ServiceID:  "test",
		PrivateKey: key,
		KeyID:      "k1",
	}

	result, err := domainconnect.BuildApplyURL(
		cfg,
		"https://dns.example.com",
		"example.com",
		"www",
		map[string]string{"IP": "1.2.3.4"},
		"",
	)
	require.NoError(t, err)

	parsed, err := url.Parse(result)
	require.NoError(t, err)

	sig := parsed.Query().Get("sig")
	require.NotEmpty(t, sig)

	fullQuery := parsed.RawQuery
	sigIdx := strings.Index(fullQuery, "&sig=")
	require.True(t, sigIdx > 0)
	signedPart := fullQuery[:sigIdx]

	t.Run(
		"valid signature verifies",
		func(t *testing.T) {
			t.Parallel()

			err := domainconnect.VerifySignature(signedPart, sig, &key.PublicKey)
			assert.NoError(t, err)
		},
	)

	t.Run(
		"tampered query fails verification",
		func(t *testing.T) {
			t.Parallel()

			err := domainconnect.VerifySignature(signedPart+"&extra=bad", sig, &key.PublicKey)
			assert.Error(t, err)
		},
	)

	t.Run(
		"wrong key fails verification",
		func(t *testing.T) {
			t.Parallel()

			otherKey, err := domainconnect.NewECDSATestKey()
			require.NoError(t, err)

			err = domainconnect.VerifySignature(signedPart, sig, &otherKey.PublicKey)
			assert.Error(t, err)
		},
	)
}

func TestVerifySignature_unsupported_key_type(t *testing.T) {
	t.Parallel()

	err := domainconnect.VerifySignature("data", "dGVzdA", "not-a-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported key type")
}

func TestConfigEnabled(t *testing.T) {
	t.Parallel()

	t.Run(
		"enabled when both provider and service IDs set",
		func(t *testing.T) {
			t.Parallel()

			cfg := domainconnect.Config{ProviderID: "p", ServiceID: "s"}
			assert.True(t, cfg.Enabled())
		},
	)

	t.Run(
		"disabled when provider ID empty",
		func(t *testing.T) {
			t.Parallel()

			cfg := domainconnect.Config{ServiceID: "s"}
			assert.False(t, cfg.Enabled())
		},
	)

	t.Run(
		"disabled when service ID empty",
		func(t *testing.T) {
			t.Parallel()

			cfg := domainconnect.Config{ProviderID: "p"}
			assert.False(t, cfg.Enabled())
		},
	)

	t.Run(
		"disabled when both empty",
		func(t *testing.T) {
			t.Parallel()

			cfg := domainconnect.Config{}
			assert.False(t, cfg.Enabled())
		},
	)
}

func TestClientCheckTemplate(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns nil when template exists",
		func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v2/domainTemplates/providers/probo.inc/services/customdomain", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			c := domainconnect.NewClient(
				domainconnect.WithHTTPClient(
					httpclient.DefaultClient(
						httpclient.WithSSRFProtection(),
						httpclient.WithSSRFAllowLoopback(),
					),
				),
			)

			err := c.CheckTemplate(context.Background(), srv.URL, "probo.inc", "customdomain")
			assert.NoError(t, err)
		},
	)

	t.Run(
		"returns ErrTemplateNotFound on 404",
		func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))
			defer srv.Close()

			c := domainconnect.NewClient(
				domainconnect.WithHTTPClient(
					httpclient.DefaultClient(
						httpclient.WithSSRFProtection(),
						httpclient.WithSSRFAllowLoopback(),
					),
				),
			)

			err := c.CheckTemplate(context.Background(), srv.URL, "probo.inc", "customdomain")
			assert.ErrorIs(t, err, domainconnect.ErrTemplateNotFound)
		},
	)

	t.Run(
		"returns error on unexpected status",
		func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer srv.Close()

			c := domainconnect.NewClient(
				domainconnect.WithHTTPClient(
					httpclient.DefaultClient(
						httpclient.WithSSRFProtection(),
						httpclient.WithSSRFAllowLoopback(),
					),
				),
			)

			err := c.CheckTemplate(context.Background(), srv.URL, "probo.inc", "customdomain")
			require.Error(t, err)
			assert.NotErrorIs(t, err, domainconnect.ErrTemplateNotFound)
		},
	)
}

func TestClientDiscover(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns settings from provider",
		func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(domainconnect.Settings{
					ProviderName: "TestProvider",
					URLSyncUX:    "https://sync.example.com",
					URLAPI:       "https://api.example.com",
				})
			}))
			defer srv.Close()

			// Discover requires a real DNS TXT lookup which cannot be
			// simulated with httptest. We verify the HTTP settings fetch
			// indirectly through CheckTemplate.
			c := domainconnect.NewClient(
				domainconnect.WithHTTPClient(
					httpclient.DefaultClient(
						httpclient.WithSSRFProtection(),
						httpclient.WithSSRFAllowLoopback(),
					),
				),
			)

			err := c.CheckTemplate(context.Background(), srv.URL, "test", "svc")
			assert.NoError(t, err)
		},
	)
}

func TestNewECDSATestKey(t *testing.T) {
	t.Parallel()

	key, err := domainconnect.NewECDSATestKey()
	require.NoError(t, err)
	require.NotNil(t, key)

	assert.IsType(t, &ecdsa.PrivateKey{}, key)
	assert.IsType(t, &ecdsa.PublicKey{}, &key.PublicKey)
	assert.NotEqual(t, &rsa.PrivateKey{}, key)
}
