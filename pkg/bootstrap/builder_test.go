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
)

func mockEnv(env map[string]string) EnvGetter {
	return func(key string) string {
		return env[key]
	}
}

func requiredEnv() map[string]string {
	return map[string]string{
		"PROBOD_ENCRYPTION_KEY":     "test-encryption-key-32-bytes-long",
		"AUTH_COOKIE_SECRET":        "test-cookie-secret-32-bytes-long!",
		"AUTH_PASSWORD_PEPPER":      "test-password-pepper-32-bytes-lo",
		"OAUTH2_SERVER_SIGNING_KEY": "test-oauth2-signing-key",
	}
}

func TestBuilder_Build_MissingRequiredEnvVars(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantMissing []string
	}{
		{
			name:        "all missing",
			env:         map[string]string{},
			wantMissing: []string{"PROBOD_ENCRYPTION_KEY", "AUTH_COOKIE_SECRET", "AUTH_PASSWORD_PEPPER", "OAUTH2_SERVER_SIGNING_KEY"},
		},
		{
			name: "missing oauth2 signing key",
			env: map[string]string{
				"PROBOD_ENCRYPTION_KEY": "key",
				"AUTH_COOKIE_SECRET":    "secret",
				"AUTH_PASSWORD_PEPPER":  "pepper",
			},
			wantMissing: []string{"OAUTH2_SERVER_SIGNING_KEY"},
		},
		{
			name: "missing encryption key",
			env: map[string]string{
				"AUTH_COOKIE_SECRET":   "secret",
				"AUTH_PASSWORD_PEPPER": "pepper",
			},
			wantMissing: []string{"PROBOD_ENCRYPTION_KEY"},
		},
		{
			name: "missing cookie secret",
			env: map[string]string{
				"PROBOD_ENCRYPTION_KEY": "key",
				"AUTH_PASSWORD_PEPPER":  "pepper",
			},
			wantMissing: []string{"AUTH_COOKIE_SECRET"},
		},
		{
			name: "slack connector missing required fields",
			env: map[string]string{
				"PROBOD_ENCRYPTION_KEY":     "key",
				"AUTH_COOKIE_SECRET":        "secret",
				"AUTH_PASSWORD_PEPPER":      "pepper",
				"CONNECTOR_SLACK_CLIENT_ID": "client-id",
			},
			wantMissing: []string{"CONNECTOR_SLACK_CLIENT_SECRET", "CONNECTOR_SLACK_SIGNING_SECRET"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(mockEnv(tt.env))
			_, err := b.Build()

			require.Error(t, err)
			for _, missing := range tt.wantMissing {
				assert.Contains(t, err.Error(), missing)
			}
		})
	}
}

func TestBuilder_Build_Defaults(t *testing.T) {
	b := NewBuilder(mockEnv(requiredEnv()))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	// Unit config
	assert.Equal(t, "localhost:8081", cfg.Unit.Metrics.Addr)
	assert.Equal(t, "localhost:4317", cfg.Unit.Tracing.Addr)
	assert.Equal(t, 512, cfg.Unit.Tracing.MaxBatchSize)
	assert.Equal(t, 5, cfg.Unit.Tracing.BatchTimeout)
	assert.Equal(t, 30, cfg.Unit.Tracing.ExportTimeout)
	assert.Equal(t, 2048, cfg.Unit.Tracing.MaxQueueSize)

	// Probod base config
	assert.Equal(t, "http://localhost:8080", cfg.Probod.BaseURL)
	assert.Equal(t, "localhost:9222", cfg.Probod.ChromeDPAddr)

	// API config
	assert.Equal(t, ":8080", cfg.Probod.Api.Addr)
	assert.Nil(t, cfg.Probod.Api.ProxyProtocol.TrustedProxies)
	assert.Equal(t, []string{"http://localhost:8080"}, cfg.Probod.Api.Cors.AllowedOrigins)

	// PG config
	assert.Equal(t, "localhost:5432", cfg.Probod.Pg.Addr)
	assert.Equal(t, "postgres", cfg.Probod.Pg.Username)
	assert.Equal(t, "postgres", cfg.Probod.Pg.Password)
	assert.Equal(t, "probod", cfg.Probod.Pg.Database)
	assert.Equal(t, int32(100), cfg.Probod.Pg.PoolSize)
	assert.Equal(t, int32(10), cfg.Probod.Pg.MinPoolSize)
	assert.Equal(t, 1800, cfg.Probod.Pg.MaxConnIdleTimeSeconds)
	assert.Equal(t, 3600, cfg.Probod.Pg.MaxConnLifetimeSeconds)
	assert.False(t, cfg.Probod.Pg.Debug)

	// Auth config
	assert.False(t, cfg.Probod.Auth.DisableSignup)
	assert.Equal(t, 3600, cfg.Probod.Auth.InvitationConfirmationTokenValidity)
	assert.Equal(t, 3600, cfg.Probod.Auth.PasswordResetTokenValidity)
	assert.Equal(t, 900, cfg.Probod.Auth.MagicLinkTokenValidity)
	assert.Equal(t, "SSID", cfg.Probod.Auth.Cookie.Name)
	assert.Equal(t, "localhost", cfg.Probod.Auth.Cookie.Domain)
	assert.Equal(t, 24, cfg.Probod.Auth.Cookie.Duration)
	assert.True(t, cfg.Probod.Auth.Cookie.Secure)
	assert.Equal(t, 1000000, cfg.Probod.Auth.Password.Iterations)

	// SAML config
	assert.Equal(t, 604800, cfg.Probod.Auth.SAML.SessionDuration)
	assert.Equal(t, 0, cfg.Probod.Auth.SAML.CleanupIntervalSeconds)
	assert.Equal(t, 60, cfg.Probod.Auth.SAML.DomainVerificationIntervalSeconds)
	assert.Equal(t, "8.8.8.8:53", cfg.Probod.Auth.SAML.DomainVerificationResolverAddr)

	// Trust center config
	assert.Equal(t, ":80", cfg.Probod.TrustCenter.HTTPAddr)
	assert.Equal(t, ":443", cfg.Probod.TrustCenter.HTTPSAddr)
	assert.Nil(t, cfg.Probod.TrustCenter.ProxyProtocol.TrustedProxies)

	// AWS config
	assert.Equal(t, "us-east-1", cfg.Probod.AWS.Region)
	assert.Equal(t, "probod", cfg.Probod.AWS.Bucket)
	assert.False(t, cfg.Probod.AWS.UsePathStyle)

	// Notifications config
	assert.Equal(t, "Probo", cfg.Probod.Notifications.Mailer.SenderName)
	assert.Equal(t, "no-reply@notification.getprobo.com", cfg.Probod.Notifications.Mailer.SenderEmail)
	assert.Equal(t, "localhost:1025", cfg.Probod.Notifications.Mailer.SMTP.Addr)
	assert.False(t, cfg.Probod.Notifications.Mailer.SMTP.TLSRequired)
	assert.Equal(t, 60, cfg.Probod.Notifications.Mailer.MailerInterval)
	assert.Equal(t, 60, cfg.Probod.Notifications.Slack.SenderInterval)
	assert.Empty(t, cfg.Probod.Notifications.Slack.SigningSecret)
	assert.Equal(t, 5, cfg.Probod.Notifications.Webhook.SenderInterval)
	assert.Equal(t, 86400, cfg.Probod.Notifications.Webhook.CacheTTL)

	// Agents config — default
	assert.Equal(t, "openai", cfg.Probod.Agents.Default.Provider)
	assert.Equal(t, "gpt-4o", cfg.Probod.Agents.Default.ModelName)
	assert.Equal(t, new(0.1), cfg.Probod.Agents.Default.Temperature)
	assert.Equal(t, new(4096), cfg.Probod.Agents.Default.MaxTokens)
	// Agents config — per-agent overrides are empty (inherit from default)
	assert.Empty(t, cfg.Probod.Agents.Probo.Provider)
	assert.Empty(t, cfg.Probod.Agents.Probo.ModelName)
	assert.Nil(t, cfg.Probod.Agents.Probo.Temperature)
	assert.Nil(t, cfg.Probod.Agents.Probo.MaxTokens)
	assert.Empty(t, cfg.Probod.Agents.EvidenceDescriber.Provider)
	assert.Empty(t, cfg.Probod.Agents.EvidenceDescriber.ModelName)
	assert.Nil(t, cfg.Probod.Agents.EvidenceDescriber.Temperature)
	assert.Nil(t, cfg.Probod.Agents.EvidenceDescriber.MaxTokens)

	// Custom domains config
	assert.Equal(t, 3600, cfg.Probod.CustomDomains.RenewalInterval)
	assert.Equal(t, 30, cfg.Probod.CustomDomains.ProvisionInterval)
	assert.Equal(t, "custom.getprobo.com", cfg.Probod.CustomDomains.CnameTarget)
	assert.Equal(t, "8.8.8.8:53", cfg.Probod.CustomDomains.ResolverAddr)
	assert.Equal(t, "https://acme-v02.api.letsencrypt.org/directory", cfg.Probod.CustomDomains.ACME.Directory)
	assert.Equal(t, "admin@getprobo.com", cfg.Probod.CustomDomains.ACME.Email)
	assert.Equal(t, "EC256", cfg.Probod.CustomDomains.ACME.KeyType)

	// SCIM bridge config
	assert.Equal(t, 900, cfg.Probod.SCIMBridge.SyncInterval)
	assert.Equal(t, 30, cfg.Probod.SCIMBridge.PollInterval)

	// ESign config
	assert.Equal(t, "http://timestamp.digicert.com", cfg.Probod.ESign.TSAURL)

	// Branding
	assert.True(t, cfg.Probod.Branding)

	// No connectors by default
	assert.Empty(t, cfg.Probod.Connectors)
}

func TestBuilder_Build_CustomValues(t *testing.T) {
	env := requiredEnv()
	// Unit
	env["METRICS_ADDR"] = "0.0.0.0:9090"
	env["TRACING_ADDR"] = "jaeger:4317"
	env["TRACING_MAX_BATCH_SIZE"] = "1024"
	// Probod
	env["PROBOD_BASE_URL"] = "https://app.example.com"
	env["CHROME_DP_ADDR"] = "chrome:9222"
	// API
	env["API_ADDR"] = "0.0.0.0:8080"
	env["API_CORS_ALLOWED_ORIGINS"] = "https://app.example.com,https://admin.example.com"
	env["API_PROXY_PROTOCOL_TRUSTED_PROXIES"] = "10.0.0.1,10.0.0.2"
	// PG
	env["PG_ADDR"] = "postgres.example.com:5432"
	env["PG_USERNAME"] = "probo"
	env["PG_PASSWORD"] = "secret123"
	env["PG_DATABASE"] = "probo_prod"
	env["PG_POOL_SIZE"] = "200"
	env["PG_MIN_POOL_SIZE"] = "25"
	env["PG_MAX_CONN_IDLE_TIME_SECONDS"] = "900"
	env["PG_MAX_CONN_LIFETIME_SECONDS"] = "7200"
	env["PG_DEBUG"] = "true"
	// Auth
	env["AUTH_DISABLE_SIGNUP"] = "true"
	env["AUTH_INVITATION_TOKEN_VALIDITY"] = "7200"
	env["AUTH_PASSWORD_RESET_TOKEN_VALIDITY"] = "1800"
	env["AUTH_MAGIC_LINK_TOKEN_VALIDITY"] = "600"
	env["AUTH_COOKIE_DOMAIN"] = ".example.com"
	env["AUTH_COOKIE_DURATION"] = "48"
	// SAML
	env["SAML_DOMAIN_VERIFICATION_INTERVAL_SECONDS"] = "120"
	env["SAML_DOMAIN_VERIFICATION_RESOLVER_ADDR"] = "1.1.1.1:53"
	// Trust center
	env["TRUST_CENTER_HTTP_ADDR"] = ":8080"
	env["TRUST_CENTER_HTTPS_ADDR"] = ":8443"
	env["TRUST_CENTER_PROXY_PROTOCOL_TRUSTED_PROXIES"] = "10.0.1.1,10.0.1.2"
	// AWS
	env["AWS_REGION"] = "eu-west-1"
	env["AWS_BUCKET"] = "probo-files"
	env["AWS_ACCESS_KEY_ID"] = "AKIAIOSFODNN7EXAMPLE"
	env["AWS_SECRET_ACCESS_KEY"] = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	env["AWS_ENDPOINT"] = "https://s3.example.com"
	env["AWS_USE_PATH_STYLE"] = "true"
	// Notifications
	env["WEBHOOK_SENDER_INTERVAL"] = "10"
	env["WEBHOOK_CACHE_TTL"] = "3600"
	env["CONNECTOR_SLACK_SIGNING_SECRET"] = "slack-signing-secret"
	// Agents — providers
	env["OPENAI_API_KEY"] = "sk-test-key"
	env["ANTHROPIC_API_KEY"] = "sk-ant-test-key"
	// Agents — default
	env["AGENT_DEFAULT_PROVIDER"] = "openai"
	env["AGENT_DEFAULT_MODEL_NAME"] = "gpt-4-turbo"
	env["AGENT_DEFAULT_TEMPERATURE"] = "0.5"
	env["AGENT_DEFAULT_MAX_TOKENS"] = "8192"
	// Agents — evidence-describer override
	env["AGENT_EVIDENCE_DESCRIBER_PROVIDER"] = "anthropic"
	env["AGENT_EVIDENCE_DESCRIBER_MODEL_NAME"] = "claude-sonnet-4-20250514"
	env["AGENT_EVIDENCE_DESCRIBER_TEMPERATURE"] = "0.2"
	env["AGENT_EVIDENCE_DESCRIBER_MAX_TOKENS"] = "4096"
	// Custom domains
	env["CUSTOM_DOMAINS_RESOLVER_ADDR"] = "1.1.1.1:53"
	env["ACME_ACCOUNT_KEY"] = "-----BEGIN EC PRIVATE KEY-----\ntest\n-----END EC PRIVATE KEY-----"
	// SCIM bridge
	env["SCIM_BRIDGE_SYNC_INTERVAL"] = "1800"
	env["SCIM_BRIDGE_POLL_INTERVAL"] = "60"
	// ESign
	env["ESIGN_TSA_URL"] = "http://custom.tsa.example.com"
	// Branding
	env["BRANDING"] = "false"

	b := NewBuilder(mockEnv(env))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	// Unit
	assert.Equal(t, "0.0.0.0:9090", cfg.Unit.Metrics.Addr)
	assert.Equal(t, "jaeger:4317", cfg.Unit.Tracing.Addr)
	assert.Equal(t, 1024, cfg.Unit.Tracing.MaxBatchSize)
	// Probod
	assert.Equal(t, "https://app.example.com", cfg.Probod.BaseURL)
	assert.Equal(t, "chrome:9222", cfg.Probod.ChromeDPAddr)
	// API
	assert.Equal(t, "0.0.0.0:8080", cfg.Probod.Api.Addr)
	assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, cfg.Probod.Api.ProxyProtocol.TrustedProxies)
	assert.Equal(t, []string{"https://app.example.com", "https://admin.example.com"}, cfg.Probod.Api.Cors.AllowedOrigins)
	// PG
	assert.Equal(t, "postgres.example.com:5432", cfg.Probod.Pg.Addr)
	assert.Equal(t, "probo", cfg.Probod.Pg.Username)
	assert.Equal(t, "secret123", cfg.Probod.Pg.Password)
	assert.Equal(t, "probo_prod", cfg.Probod.Pg.Database)
	assert.Equal(t, int32(200), cfg.Probod.Pg.PoolSize)
	assert.Equal(t, int32(25), cfg.Probod.Pg.MinPoolSize)
	assert.Equal(t, 900, cfg.Probod.Pg.MaxConnIdleTimeSeconds)
	assert.Equal(t, 7200, cfg.Probod.Pg.MaxConnLifetimeSeconds)
	assert.True(t, cfg.Probod.Pg.Debug)
	// Auth
	assert.True(t, cfg.Probod.Auth.DisableSignup)
	assert.Equal(t, 7200, cfg.Probod.Auth.InvitationConfirmationTokenValidity)
	assert.Equal(t, 1800, cfg.Probod.Auth.PasswordResetTokenValidity)
	assert.Equal(t, 600, cfg.Probod.Auth.MagicLinkTokenValidity)
	assert.Equal(t, ".example.com", cfg.Probod.Auth.Cookie.Domain)
	assert.Equal(t, 48, cfg.Probod.Auth.Cookie.Duration)
	// SAML
	assert.Equal(t, 120, cfg.Probod.Auth.SAML.DomainVerificationIntervalSeconds)
	assert.Equal(t, "1.1.1.1:53", cfg.Probod.Auth.SAML.DomainVerificationResolverAddr)
	// Trust center
	assert.Equal(t, ":8080", cfg.Probod.TrustCenter.HTTPAddr)
	assert.Equal(t, ":8443", cfg.Probod.TrustCenter.HTTPSAddr)
	assert.Equal(t, []string{"10.0.1.1", "10.0.1.2"}, cfg.Probod.TrustCenter.ProxyProtocol.TrustedProxies)
	// AWS
	assert.Equal(t, "eu-west-1", cfg.Probod.AWS.Region)
	assert.Equal(t, "probo-files", cfg.Probod.AWS.Bucket)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", cfg.Probod.AWS.AccessKeyID)
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", cfg.Probod.AWS.SecretAccessKey)
	assert.Equal(t, "https://s3.example.com", cfg.Probod.AWS.Endpoint)
	assert.True(t, cfg.Probod.AWS.UsePathStyle)
	// Notifications
	assert.Equal(t, "slack-signing-secret", cfg.Probod.Notifications.Slack.SigningSecret)
	assert.Equal(t, 10, cfg.Probod.Notifications.Webhook.SenderInterval)
	assert.Equal(t, 3600, cfg.Probod.Notifications.Webhook.CacheTTL)
	// Agents — providers
	assert.Equal(t, "openai", cfg.Probod.Agents.Providers["openai"].Type)
	assert.Equal(t, "sk-test-key", cfg.Probod.Agents.Providers["openai"].APIKey)
	assert.Equal(t, "anthropic", cfg.Probod.Agents.Providers["anthropic"].Type)
	assert.Equal(t, "sk-ant-test-key", cfg.Probod.Agents.Providers["anthropic"].APIKey)
	// Agents — default
	assert.Equal(t, "openai", cfg.Probod.Agents.Default.Provider)
	assert.Equal(t, "gpt-4-turbo", cfg.Probod.Agents.Default.ModelName)
	assert.Equal(t, new(0.5), cfg.Probod.Agents.Default.Temperature)
	assert.Equal(t, new(8192), cfg.Probod.Agents.Default.MaxTokens)
	// Agents — probo inherits default (no overrides set)
	assert.Empty(t, cfg.Probod.Agents.Probo.Provider)
	assert.Empty(t, cfg.Probod.Agents.Probo.ModelName)
	// Agents — evidence-describer overrides
	assert.Equal(t, "anthropic", cfg.Probod.Agents.EvidenceDescriber.Provider)
	assert.Equal(t, "claude-sonnet-4-20250514", cfg.Probod.Agents.EvidenceDescriber.ModelName)
	assert.Equal(t, new(0.2), cfg.Probod.Agents.EvidenceDescriber.Temperature)
	assert.Equal(t, new(4096), cfg.Probod.Agents.EvidenceDescriber.MaxTokens)
	// Custom domains
	assert.Equal(t, "1.1.1.1:53", cfg.Probod.CustomDomains.ResolverAddr)
	assert.Equal(t, "-----BEGIN EC PRIVATE KEY-----\ntest\n-----END EC PRIVATE KEY-----", cfg.Probod.CustomDomains.ACME.AccountKey)
	// SCIM bridge
	assert.Equal(t, 1800, cfg.Probod.SCIMBridge.SyncInterval)
	assert.Equal(t, 60, cfg.Probod.SCIMBridge.PollInterval)
	// ESign
	assert.Equal(t, "http://custom.tsa.example.com", cfg.Probod.ESign.TSAURL)
	// Branding
	assert.False(t, cfg.Probod.Branding)
}

func TestBuilder_Build_SlackConnector(t *testing.T) {
	env := requiredEnv()
	env["CONNECTOR_SLACK_CLIENT_ID"] = "slack-client-id"
	env["CONNECTOR_SLACK_CLIENT_SECRET"] = "slack-client-secret"
	env["CONNECTOR_SLACK_SIGNING_SECRET"] = "slack-signing-secret"

	b := NewBuilder(mockEnv(env))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.Probod.Connectors, 1)
	connector := cfg.Probod.Connectors[0]
	assert.Equal(t, "SLACK", connector.Provider)
	assert.Equal(t, "oauth2", string(connector.Protocol))
	rawConfig := connector.RawConfig.(probod.ConnectorConfigOAuth2)
	assert.Equal(t, "slack-client-id", rawConfig.ClientID)
	assert.Equal(t, "slack-client-secret", rawConfig.ClientSecret)
	rawSettings := connector.RawSettings.(map[string]any)
	assert.Equal(t, "slack-signing-secret", rawSettings["signing-secret"])
}

func TestBuilder_Build_SAMLAutoGeneration(t *testing.T) {
	b := NewBuilder(mockEnv(requiredEnv()))

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Contains(t, cfg.Probod.Auth.SAML.Certificate, "-----BEGIN CERTIFICATE-----")
	assert.Contains(t, cfg.Probod.Auth.SAML.Certificate, "-----END CERTIFICATE-----")
	assert.Contains(t, cfg.Probod.Auth.SAML.PrivateKey, "-----BEGIN RSA PRIVATE KEY-----")
	assert.Contains(t, cfg.Probod.Auth.SAML.PrivateKey, "-----END RSA PRIVATE KEY-----")
}

func TestBuilder_Build_SAMLFromEnv(t *testing.T) {
	env := requiredEnv()
	env["SAML_CERTIFICATE"] = "env-cert"
	env["SAML_PRIVATE_KEY"] = "env-key"

	b := NewBuilder(mockEnv(env))

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "env-cert", cfg.Probod.Auth.SAML.Certificate)
	assert.Equal(t, "env-key", cfg.Probod.Auth.SAML.PrivateKey)
}

func TestBuilder_Build_SAMLPreset(t *testing.T) {
	b := NewBuilder(mockEnv(requiredEnv()))
	b.samlCertificate = "preset-cert"
	b.samlPrivateKey = "preset-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "preset-cert", cfg.Probod.Auth.SAML.Certificate)
	assert.Equal(t, "preset-key", cfg.Probod.Auth.SAML.PrivateKey)
}

func TestBuilder_Build_OAuth2Defaults(t *testing.T) {
	b := NewBuilder(mockEnv(requiredEnv()))

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.Probod.Auth.OAuth2Server.SigningKeys, 1)
	sk := cfg.Probod.Auth.OAuth2Server.SigningKeys[0]
	assert.Equal(t, "test-oauth2-signing-key", sk.PrivateKey)
	assert.Equal(t, "default", sk.KID)
	assert.True(t, sk.Active)

	assert.Equal(t, 3600, cfg.Probod.Auth.OAuth2Server.AccessTokenDuration)
	assert.Equal(t, 2592000, cfg.Probod.Auth.OAuth2Server.RefreshTokenDuration)
	assert.Equal(t, 600, cfg.Probod.Auth.OAuth2Server.AuthorizationCodeDuration)
	assert.Equal(t, 600, cfg.Probod.Auth.OAuth2Server.DeviceCodeDuration)
}

func TestBuilder_Build_OAuth2FromEnv(t *testing.T) {
	env := requiredEnv()
	env["OAUTH2_SERVER_SIGNING_KEY"] = "env-signing-key"
	env["OAUTH2_SERVER_SIGNING_KEY_KID"] = "env-kid"
	env["OAUTH2_SERVER_ACCESS_TOKEN_DURATION"] = "10"
	env["OAUTH2_SERVER_REFRESH_TOKEN_DURATION"] = "20"
	env["OAUTH2_SERVER_AUTHORIZATION_CODE_DURATION"] = "30"
	env["OAUTH2_SERVER_DEVICE_CODE_DURATION"] = "40"

	b := NewBuilder(mockEnv(env))

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.Probod.Auth.OAuth2Server.SigningKeys, 1)
	sk := cfg.Probod.Auth.OAuth2Server.SigningKeys[0]
	assert.Equal(t, "env-signing-key", sk.PrivateKey)
	assert.Equal(t, "env-kid", sk.KID)
	assert.True(t, sk.Active)

	assert.Equal(t, 10, cfg.Probod.Auth.OAuth2Server.AccessTokenDuration)
	assert.Equal(t, 20, cfg.Probod.Auth.OAuth2Server.RefreshTokenDuration)
	assert.Equal(t, 30, cfg.Probod.Auth.OAuth2Server.AuthorizationCodeDuration)
	assert.Equal(t, 40, cfg.Probod.Auth.OAuth2Server.DeviceCodeDuration)
}

func TestBuilder_Build_OAuth2Preset(t *testing.T) {
	env := requiredEnv()
	delete(env, "OAUTH2_SERVER_SIGNING_KEY")

	b := NewBuilder(mockEnv(env))
	b.oauth2SigningKey = "preset-signing-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.Probod.Auth.OAuth2Server.SigningKeys, 1)
	assert.Equal(t, "preset-signing-key", cfg.Probod.Auth.OAuth2Server.SigningKeys[0].PrivateKey)
}

func TestBuilder_Build_PgCABundleFromEnv(t *testing.T) {
	env := requiredEnv()
	env["PG_CA_BUNDLE"] = "test-ca-bundle-content"

	b := NewBuilder(mockEnv(env))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "test-ca-bundle-content", cfg.Probod.Pg.CACertBundle)
}

func TestBuilder_Build_PgCABundleFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "ca-bundle.pem")
	err := os.WriteFile(caFile, []byte("ca-bundle-from-file"), 0644)
	require.NoError(t, err)

	env := requiredEnv()
	env["PG_CA_BUNDLE_PATH"] = caFile

	b := NewBuilder(mockEnv(env))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "ca-bundle-from-file", cfg.Probod.Pg.CACertBundle)
}

func TestBuilder_parseOriginsList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single origin",
			input: "http://localhost:8080",
			want:  []string{"http://localhost:8080"},
		},
		{
			name:  "multiple origins",
			input: "http://localhost:8080,https://example.com",
			want:  []string{"http://localhost:8080", "https://example.com"},
		},
		{
			name:  "quoted origins",
			input: `"http://localhost:8080","https://example.com"`,
			want:  []string{"http://localhost:8080", "https://example.com"},
		},
		{
			name:  "with spaces",
			input: "http://localhost:8080 , https://example.com",
			want:  []string{"http://localhost:8080", "https://example.com"},
		},
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(nil)
			got := b.parseOriginsList(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
