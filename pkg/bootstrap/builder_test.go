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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/probodconfig"
)

func mockEnv(env map[string]string) EnvGetter {
	return func(key string) string {
		return env[key]
	}
}

func requiredEnv() map[string]string {
	return map[string]string{
		"PROBOD_ENCRYPTION_KEY":            "test-encryption-key-32-bytes-long",
		"PROBOD_AUTH_COOKIE_SECRET":        "test-cookie-secret-32-bytes-long!",
		"PROBOD_AUTH_PASSWORD_PEPPER":      "test-password-pepper-32-bytes-lo",
		"PROBOD_OAUTH2_SERVER_SIGNING_KEY": "test-oauth2-signing-key",
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
			wantMissing: []string{"PROBOD_ENCRYPTION_KEY", "PROBOD_AUTH_COOKIE_SECRET", "PROBOD_AUTH_PASSWORD_PEPPER", "PROBOD_OAUTH2_SERVER_SIGNING_KEY"},
		},
		{
			name: "missing oauth2 signing key",
			env: map[string]string{
				"PROBOD_ENCRYPTION_KEY":       "key",
				"PROBOD_AUTH_COOKIE_SECRET":   "secret",
				"PROBOD_AUTH_PASSWORD_PEPPER": "pepper",
			},
			wantMissing: []string{"PROBOD_OAUTH2_SERVER_SIGNING_KEY"},
		},
		{
			name: "missing encryption key",
			env: map[string]string{
				"PROBOD_AUTH_COOKIE_SECRET":   "secret",
				"PROBOD_AUTH_PASSWORD_PEPPER": "pepper",
			},
			wantMissing: []string{"PROBOD_ENCRYPTION_KEY"},
		},
		{
			name: "missing cookie secret",
			env: map[string]string{
				"PROBOD_ENCRYPTION_KEY":       "key",
				"PROBOD_AUTH_PASSWORD_PEPPER": "pepper",
			},
			wantMissing: []string{"PROBOD_AUTH_COOKIE_SECRET"},
		},
		{
			name: "slack connector missing required fields",
			env: map[string]string{
				"PROBOD_ENCRYPTION_KEY":            "key",
				"PROBOD_AUTH_COOKIE_SECRET":        "secret",
				"PROBOD_AUTH_PASSWORD_PEPPER":      "pepper",
				"PROBOD_CONNECTOR_SLACK_CLIENT_ID": "client-id",
			},
			wantMissing: []string{"PROBOD_CONNECTOR_SLACK_CLIENT_SECRET", "PROBOD_CONNECTOR_SLACK_SIGNING_SECRET"},
		},
		{
			name: "google workspace connector missing required fields",
			env: map[string]string{
				"PROBOD_ENCRYPTION_KEY":                       "key",
				"PROBOD_AUTH_COOKIE_SECRET":                   "secret",
				"PROBOD_AUTH_PASSWORD_PEPPER":                 "pepper",
				"PROBOD_CONNECTOR_GOOGLE_WORKSPACE_CLIENT_ID": "client-id",
			},
			wantMissing: []string{"PROBOD_CONNECTOR_GOOGLE_WORKSPACE_CLIENT_SECRET"},
		},
		{
			name: "microsoft 365 connector missing required fields",
			env: map[string]string{
				"PROBOD_ENCRYPTION_KEY":                    "key",
				"PROBOD_AUTH_COOKIE_SECRET":                "secret",
				"PROBOD_AUTH_PASSWORD_PEPPER":              "pepper",
				"PROBOD_CONNECTOR_MICROSOFT_365_CLIENT_ID": "client-id",
			},
			wantMissing: []string{"PROBOD_CONNECTOR_MICROSOFT_365_CLIENT_SECRET"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(NewResolver(mockEnv(tt.env)))
			_, err := b.Build()

			require.Error(t, err)

			for _, missing := range tt.wantMissing {
				assert.Contains(t, err.Error(), missing)
			}
		})
	}
}

func TestBuilder_Build_Defaults(t *testing.T) {
	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	// Unit config
	assert.Equal(t, "localhost:8081", cfg.Unit.Metrics.Addr)
	assert.Equal(t, "localhost:4318", cfg.Unit.Tracing.Addr)
	assert.Equal(t, 512, cfg.Unit.Tracing.MaxBatchSize)
	assert.Equal(t, 5, cfg.Unit.Tracing.BatchTimeout)
	assert.Equal(t, 30, cfg.Unit.Tracing.ExportTimeout)
	assert.Equal(t, 2048, cfg.Unit.Tracing.MaxQueueSize)

	// Probod base config
	assert.Empty(t, cfg.Probod.BaseURL)
	assert.Empty(t, cfg.Probod.ChromeDPAddr)

	// API config
	assert.Empty(t, cfg.Probod.Api.Addr)
	assert.Nil(t, cfg.Probod.Api.ProxyProtocol.TrustedProxies)
	assert.Nil(t, cfg.Probod.Api.Cors.AllowedOrigins)
	assert.Equal(t, 15000, cfg.Probod.Api.GraphQL.ParserTokenLimit)
	assert.Equal(t, 2000, cfg.Probod.Api.GraphQL.ComplexityLimit)
	assert.Equal(t, 1000, cfg.Probod.Api.GraphQL.QueryCacheSize)
	assert.True(t, cfg.Probod.Api.GraphQL.DisableSuggestion)

	// PG config
	assert.Empty(t, cfg.Probod.Pg.Addr)
	assert.Empty(t, cfg.Probod.Pg.Username)
	assert.Empty(t, cfg.Probod.Pg.Password)
	assert.Empty(t, cfg.Probod.Pg.Database)
	assert.Equal(t, int32(100), cfg.Probod.Pg.PoolSize)
	assert.Equal(t, int32(10), cfg.Probod.Pg.MinPoolSize)
	assert.Equal(t, 1800, cfg.Probod.Pg.MaxConnIdleTimeSeconds)
	assert.Equal(t, 3600, cfg.Probod.Pg.MaxConnLifetimeSeconds)
	assert.Equal(t, 300, cfg.Probod.Pg.MaxConnLifetimeJitterSeconds)
	assert.Equal(t, 60, cfg.Probod.Pg.HealthCheckPeriodSeconds)
	assert.False(t, cfg.Probod.Pg.Debug)

	// Auth config
	assert.False(t, cfg.Probod.Auth.DisableSignup)
	assert.Equal(t, 3600, cfg.Probod.Auth.InvitationConfirmationTokenValidity)
	assert.Equal(t, 3600, cfg.Probod.Auth.PasswordResetTokenValidity)
	assert.Equal(t, 900, cfg.Probod.Auth.MagicLinkTokenValidity)
	assert.Empty(t, cfg.Probod.Auth.Cookie.Name)
	assert.Empty(t, cfg.Probod.Auth.Cookie.Domain)
	assert.Equal(t, 24, cfg.Probod.Auth.Cookie.Duration)
	assert.True(t, cfg.Probod.Auth.Cookie.Secure)
	assert.Equal(t, 1000000, cfg.Probod.Auth.Password.Iterations)

	// SAML config
	assert.Equal(t, 604800, cfg.Probod.Auth.SAML.SessionDuration)
	assert.Equal(t, 0, cfg.Probod.Auth.SAML.CleanupIntervalSeconds)
	assert.Equal(t, 60, cfg.Probod.Auth.SAML.DomainVerificationIntervalSeconds)
	assert.Empty(t, cfg.Probod.Auth.SAML.DomainVerificationResolverAddr)

	// Trust center config
	assert.Empty(t, cfg.Probod.TrustCenter.HTTPAddr)
	assert.Empty(t, cfg.Probod.TrustCenter.HTTPSAddr)
	assert.Nil(t, cfg.Probod.TrustCenter.ProxyProtocol.TrustedProxies)

	// AWS config
	assert.Empty(t, cfg.Probod.AWS.Region)
	assert.Empty(t, cfg.Probod.AWS.Bucket)
	assert.False(t, cfg.Probod.AWS.UsePathStyle)

	// Notifications config
	assert.Empty(t, cfg.Probod.Notifications.Mailer.SenderName)
	assert.Empty(t, cfg.Probod.Notifications.Mailer.SenderEmail)
	assert.Empty(t, cfg.Probod.Notifications.Mailer.SMTP.Addr)
	assert.False(t, cfg.Probod.Notifications.Mailer.SMTP.TLSRequired)
	assert.Empty(t, cfg.Probod.Notifications.Mailer.SMTP.HelloName)
	assert.Equal(t, 60, cfg.Probod.Notifications.Mailer.MailerInterval)
	assert.Equal(t, 60, cfg.Probod.Notifications.Slack.SenderInterval)
	assert.Empty(t, cfg.Probod.Notifications.Slack.SigningSecret)
	assert.Equal(t, 5, cfg.Probod.Notifications.Webhook.SenderInterval)
	assert.Equal(t, 86400, cfg.Probod.Notifications.Webhook.CacheTTL)
	assert.Equal(t, 300, cfg.Probod.Notifications.Document.Interval)
	assert.Equal(t, 900, cfg.Probod.Notifications.Document.DebounceDelay)
	assert.Equal(t, 86400, cfg.Probod.Notifications.Document.ReminderInterval)

	// Agents tools — Firecrawl empty by default
	assert.Empty(t, cfg.Probod.Agents.Tools.FirecrawlAPIKey)

	// Agents config — default
	assert.Equal(t, "openai", cfg.Probod.Agents.Default.Provider)
	assert.Equal(t, "gpt-4o", cfg.Probod.Agents.Default.ModelName)
	assert.Equal(t, new(0.1), cfg.Probod.Agents.Default.Temperature)
	assert.Equal(t, new(4096), cfg.Probod.Agents.Default.MaxTokens)
	assert.Nil(t, cfg.Probod.Agents.Providers)
	// Agents config — per-agent overrides are empty (inherit from default)
	assert.Empty(t, cfg.Probod.Agents.Probo.Provider)
	assert.Empty(t, cfg.Probod.Agents.Probo.ModelName)
	assert.Nil(t, cfg.Probod.Agents.Probo.Temperature)
	assert.Nil(t, cfg.Probod.Agents.Probo.MaxTokens)
	assert.Empty(t, cfg.Probod.Agents.EvidenceDescriber.Provider)
	assert.Empty(t, cfg.Probod.Agents.EvidenceDescriber.ModelName)
	assert.Nil(t, cfg.Probod.Agents.EvidenceDescriber.Temperature)
	assert.Nil(t, cfg.Probod.Agents.EvidenceDescriber.MaxTokens)
	assert.Empty(t, cfg.Probod.Agents.ThirdPartyVetter.Provider)
	assert.Empty(t, cfg.Probod.Agents.ThirdPartyVetter.ModelName)
	assert.Nil(t, cfg.Probod.Agents.ThirdPartyVetter.Temperature)
	assert.Nil(t, cfg.Probod.Agents.ThirdPartyVetter.MaxTokens)
	assert.Empty(t, cfg.Probod.Agents.ThirdPartyDisambiguation.Provider)
	assert.Empty(t, cfg.Probod.Agents.ThirdPartyDisambiguation.ModelName)
	assert.Nil(t, cfg.Probod.Agents.ThirdPartyDisambiguation.Temperature)
	assert.Equal(t, new(4096), cfg.Probod.Agents.ThirdPartyDisambiguation.MaxTokens)
	assert.Empty(t, cfg.Probod.Agents.TrackerMapping.Provider)
	assert.Empty(t, cfg.Probod.Agents.TrackerMapping.ModelName)
	assert.Nil(t, cfg.Probod.Agents.TrackerMapping.Temperature)
	assert.Equal(t, new(4096), cfg.Probod.Agents.TrackerMapping.MaxTokens)
	assert.Empty(t, cfg.Probod.Agents.TrackerEnrichment.Provider)
	assert.Empty(t, cfg.Probod.Agents.TrackerEnrichment.ModelName)
	assert.Nil(t, cfg.Probod.Agents.TrackerEnrichment.Temperature)
	assert.Equal(t, new(4096), cfg.Probod.Agents.TrackerEnrichment.MaxTokens)

	// Tracker worker tuning — defaults
	assert.Equal(t, 10, cfg.Probod.TrackerMappingWorker.Interval)
	assert.Equal(t, 3, cfg.Probod.TrackerMappingWorker.MaxConcurrency)
	assert.Equal(t, 600, cfg.Probod.TrackerMappingWorker.StaleAfter)
	assert.Equal(t, 45, cfg.Probod.TrackerMappingWorker.AgentTimeout)
	assert.Equal(t, 10, cfg.Probod.TrackerMappingWorker.AgentMaxTurns)
	assert.Equal(t, 45, cfg.Probod.TrackerMappingWorker.DisambiguationAgentTimeout)
	assert.Equal(t, 10, cfg.Probod.CommonPatternEnrichmentWorker.Interval)
	assert.Equal(t, 2, cfg.Probod.CommonPatternEnrichmentWorker.MaxConcurrency)
	assert.Equal(t, 600, cfg.Probod.CommonPatternEnrichmentWorker.StaleAfter)
	assert.Equal(t, 45, cfg.Probod.CommonPatternEnrichmentWorker.AgentTimeout)
	assert.Equal(t, 10, cfg.Probod.CommonPatternEnrichmentWorker.AgentMaxTurns)
	assert.Equal(t, 10, cfg.Probod.CommonThirdPartyEnrichmentWorker.Interval)
	assert.Equal(t, 1, cfg.Probod.CommonThirdPartyEnrichmentWorker.MaxConcurrency)
	assert.Equal(t, 900, cfg.Probod.CommonThirdPartyEnrichmentWorker.StaleAfter)
	assert.Equal(t, 90, cfg.Probod.CommonThirdPartyEnrichmentWorker.AgentTimeout)
	assert.Equal(t, 12, cfg.Probod.CommonThirdPartyEnrichmentWorker.AgentMaxTurns)
	assert.Equal(t, 0.7, cfg.Probod.CommonThirdPartyEnrichmentWorker.ConfidenceThreshold)
	assert.Equal(t, 3, cfg.Probod.CommonThirdPartyEnrichmentWorker.MaxAttempts)
	assert.Equal(t, 10, cfg.Probod.ThirdPartyVetting.Interval)
	assert.Equal(t, 1500, cfg.Probod.ThirdPartyVetting.StaleAfter)
	assert.Equal(t, 1, cfg.Probod.ThirdPartyVetting.MaxConcurrency)

	// Custom domains config
	assert.Equal(t, 3600, cfg.Probod.CustomDomains.RenewalInterval)
	assert.Equal(t, 30, cfg.Probod.CustomDomains.ProvisionInterval)
	assert.Equal(t, "custom.getprobo.com", cfg.Probod.CustomDomains.CnameTarget)
	assert.Empty(t, cfg.Probod.CustomDomains.ResolverAddr)
	assert.Empty(t, cfg.Probod.CustomDomains.ACME.Directory)
	assert.Empty(t, cfg.Probod.CustomDomains.ACME.Email)
	assert.Empty(t, cfg.Probod.CustomDomains.ACME.KeyType)

	// SCIM bridge config
	assert.Equal(t, 900, cfg.Probod.SCIMBridge.SyncInterval)
	assert.Equal(t, 30, cfg.Probod.SCIMBridge.PollInterval)

	// ESign config
	assert.Empty(t, cfg.Probod.ESign.TSAURL)

	// Branding
	assert.True(t, cfg.Probod.Branding)

	// No connectors by default
	assert.Empty(t, cfg.Probod.Connectors)
}

func TestBuilder_Build_CustomValues(t *testing.T) {
	env := requiredEnv()
	// Unit
	env["PROBOD_METRICS_ADDR"] = "0.0.0.0:9090"
	env["PROBOD_TRACING_ADDR"] = "jaeger:4317"
	env["PROBOD_TRACING_MAX_BATCH_SIZE"] = "1024"
	// Probod
	env["PROBOD_BASE_URL"] = "https://app.example.com"
	env["PROBOD_CHROME_DP_ADDR"] = "chrome:9222"
	// API
	env["PROBOD_API_ADDR"] = "0.0.0.0:8080"
	env["PROBOD_API_CORS_ALLOWED_ORIGINS"] = "https://app.example.com,https://admin.example.com"
	env["PROBOD_API_PROXY_PROTOCOL_TRUSTED_PROXIES"] = "10.0.0.1,10.0.0.2"
	env["PROBOD_API_GRAPHQL_PARSER_TOKEN_LIMIT"] = "20000"
	env["PROBOD_API_GRAPHQL_COMPLEXITY_LIMIT"] = "5000"
	env["PROBOD_API_GRAPHQL_QUERY_CACHE_SIZE"] = "2000"
	env["PROBOD_API_GRAPHQL_DISABLE_SUGGESTION"] = "false"
	// PG
	env["PROBOD_PG_ADDR"] = "postgres.example.com:5432"
	env["PROBOD_PG_USERNAME"] = "probo"
	env["PROBOD_PG_PASSWORD"] = "secret123"
	env["PROBOD_PG_DATABASE"] = "probo_prod"
	env["PROBOD_PG_POOL_SIZE"] = "200"
	env["PROBOD_PG_MIN_POOL_SIZE"] = "25"
	env["PROBOD_PG_MAX_CONN_IDLE_TIME_SECONDS"] = "900"
	env["PROBOD_PG_MAX_CONN_LIFETIME_SECONDS"] = "7200"
	env["PROBOD_PG_MAX_CONN_LIFETIME_JITTER_SECONDS"] = "600"
	env["PROBOD_PG_HEALTH_CHECK_PERIOD_SECONDS"] = "30"
	env["PROBOD_PG_DEBUG"] = "true"
	// Auth
	env["PROBOD_AUTH_DISABLE_SIGNUP"] = "true"
	env["PROBOD_AUTH_INVITATION_TOKEN_VALIDITY"] = "7200"
	env["PROBOD_AUTH_PASSWORD_RESET_TOKEN_VALIDITY"] = "1800"
	env["PROBOD_AUTH_MAGIC_LINK_TOKEN_VALIDITY"] = "600"
	env["PROBOD_AUTH_COOKIE_DOMAIN"] = ".example.com"
	env["PROBOD_AUTH_COOKIE_DURATION"] = "48"
	// SAML
	env["PROBOD_SAML_DOMAIN_VERIFICATION_INTERVAL_SECONDS"] = "120"
	env["PROBOD_SAML_DOMAIN_VERIFICATION_RESOLVER_ADDR"] = "1.1.1.1:53"
	// Trust center
	env["PROBOD_TRUST_CENTER_HTTP_ADDR"] = ":8080"
	env["PROBOD_TRUST_CENTER_HTTPS_ADDR"] = ":8443"
	env["PROBOD_TRUST_CENTER_PROXY_PROTOCOL_TRUSTED_PROXIES"] = "10.0.1.1,10.0.1.2"
	// AWS
	env["PROBOD_AWS_REGION"] = "eu-west-1"
	env["PROBOD_AWS_BUCKET"] = "probo-files"
	env["PROBOD_AWS_ACCESS_KEY_ID"] = "AKIAIOSFODNN7EXAMPLE"
	env["PROBOD_AWS_SECRET_ACCESS_KEY"] = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	env["PROBOD_AWS_ENDPOINT"] = "https://s3.example.com"
	env["PROBOD_AWS_USE_PATH_STYLE"] = "true"
	// Notifications
	env["PROBOD_WEBHOOK_SENDER_INTERVAL"] = "10"
	env["PROBOD_WEBHOOK_CACHE_TTL"] = "3600"
	env["PROBOD_CONNECTOR_SLACK_SIGNING_SECRET"] = "slack-signing-secret"
	env["PROBOD_DOCUMENT_NOTIFICATION_INTERVAL"] = "120"
	env["PROBOD_DOCUMENT_NOTIFICATION_DEBOUNCE_DELAY"] = "60"
	env["PROBOD_DOCUMENT_NOTIFICATION_REMINDER_INTERVAL"] = "43200"
	// Firecrawl
	env["PROBOD_FIRECRAWL_API_KEY"] = "fc-test-key"
	// Agents — providers
	env["PROBOD_OPENAI_API_KEY"] = "sk-test-key"
	env["PROBOD_ANTHROPIC_API_KEY"] = "sk-ant-test-key"
	// Agents — default
	env["PROBOD_AGENT_DEFAULT_PROVIDER"] = "openai"
	env["PROBOD_AGENT_DEFAULT_MODEL_NAME"] = "gpt-4-turbo"
	env["PROBOD_AGENT_DEFAULT_TEMPERATURE"] = "0.5"
	env["PROBOD_AGENT_DEFAULT_MAX_TOKENS"] = "8192"
	// Agents — evidence-describer override
	env["PROBOD_AGENT_EVIDENCE_DESCRIBER_PROVIDER"] = "anthropic"
	env["PROBOD_AGENT_EVIDENCE_DESCRIBER_MODEL_NAME"] = "claude-sonnet-4-20250514"
	env["PROBOD_AGENT_EVIDENCE_DESCRIBER_TEMPERATURE"] = "0.2"
	env["PROBOD_AGENT_EVIDENCE_DESCRIBER_MAX_TOKENS"] = "4096"
	// Agents — third-party-vetter override
	env["PROBOD_AGENT_THIRD_PARTY_VETTER_PROVIDER"] = "openai"
	env["PROBOD_AGENT_THIRD_PARTY_VETTER_MODEL_NAME"] = "gpt-4o"
	env["PROBOD_AGENT_THIRD_PARTY_VETTER_TEMPERATURE"] = "0.3"
	env["PROBOD_AGENT_THIRD_PARTY_VETTER_MAX_TOKENS"] = "8192"
	// Agents — third-party-disambiguation override
	env["PROBOD_AGENT_THIRD_PARTY_DISAMBIGUATION_PROVIDER"] = "anthropic"
	env["PROBOD_AGENT_THIRD_PARTY_DISAMBIGUATION_MODEL_NAME"] = "claude-sonnet-4-20250514"
	env["PROBOD_AGENT_THIRD_PARTY_DISAMBIGUATION_TEMPERATURE"] = "0.4"
	env["PROBOD_AGENT_THIRD_PARTY_DISAMBIGUATION_MAX_TOKENS"] = "2048"
	// Agents — tracker-mapping override
	env["PROBOD_AGENT_TRACKER_MAPPING_PROVIDER"] = "openai"
	env["PROBOD_AGENT_TRACKER_MAPPING_MODEL_NAME"] = "gpt-4o-mini"
	env["PROBOD_AGENT_TRACKER_MAPPING_TEMPERATURE"] = "0.1"
	env["PROBOD_AGENT_TRACKER_MAPPING_MAX_TOKENS"] = "1024"
	// Agents — tracker-enrichment override
	env["PROBOD_AGENT_TRACKER_ENRICHMENT_PROVIDER"] = "openai"
	env["PROBOD_AGENT_TRACKER_ENRICHMENT_MODEL_NAME"] = "gpt-4o"
	env["PROBOD_AGENT_TRACKER_ENRICHMENT_TEMPERATURE"] = "0.2"
	env["PROBOD_AGENT_TRACKER_ENRICHMENT_MAX_TOKENS"] = "2048"
	// Tracker worker tuning override
	env["PROBOD_TRACKER_MAPPING_INTERVAL"] = "20"
	env["PROBOD_TRACKER_MAPPING_MAX_CONCURRENCY"] = "5"
	env["PROBOD_TRACKER_MAPPING_STALE_AFTER"] = "1200"
	env["PROBOD_TRACKER_MAPPING_AGENT_TIMEOUT"] = "30"
	env["PROBOD_TRACKER_MAPPING_AGENT_MAX_TURNS"] = "6"
	env["PROBOD_TRACKER_MAPPING_DISAMBIGUATION_AGENT_TIMEOUT"] = "35"
	env["PROBOD_COMMON_PATTERN_ENRICHMENT_INTERVAL"] = "15"
	env["PROBOD_COMMON_PATTERN_ENRICHMENT_MAX_CONCURRENCY"] = "4"
	env["PROBOD_COMMON_PATTERN_ENRICHMENT_STALE_AFTER"] = "900"
	env["PROBOD_COMMON_PATTERN_ENRICHMENT_AGENT_TIMEOUT"] = "50"
	env["PROBOD_COMMON_PATTERN_ENRICHMENT_AGENT_MAX_TURNS"] = "5"
	// Common third party enrichment agent + worker tuning override
	env["PROBOD_AGENT_COMMON_THIRD_PARTY_ENRICHMENT_PROVIDER"] = "openai"
	env["PROBOD_AGENT_COMMON_THIRD_PARTY_ENRICHMENT_MODEL_NAME"] = "gpt-4o"
	env["PROBOD_AGENT_COMMON_THIRD_PARTY_ENRICHMENT_MAX_TOKENS"] = "16384"
	env["PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_INTERVAL"] = "25"
	env["PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_MAX_CONCURRENCY"] = "2"
	env["PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_STALE_AFTER"] = "1200"
	env["PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_AGENT_TIMEOUT"] = "120"
	env["PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_AGENT_MAX_TURNS"] = "8"
	env["PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_CONFIDENCE_THRESHOLD"] = "0.85"
	env["PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_MAX_ATTEMPTS"] = "5"
	env["PROBOD_THIRD_PARTY_VETTING_INTERVAL"] = "15"
	env["PROBOD_THIRD_PARTY_VETTING_STALE_AFTER"] = "1800"
	env["PROBOD_THIRD_PARTY_VETTING_MAX_CONCURRENCY"] = "2"
	// Custom domains
	env["PROBOD_CUSTOM_DOMAINS_RESOLVER_ADDR"] = "1.1.1.1:53"
	env["PROBOD_ACME_ACCOUNT_KEY"] = "-----BEGIN EC PRIVATE KEY-----\ntest\n-----END EC PRIVATE KEY-----"
	// SCIM bridge
	env["PROBOD_SCIM_BRIDGE_SYNC_INTERVAL"] = "1800"
	env["PROBOD_SCIM_BRIDGE_POLL_INTERVAL"] = "60"
	// ESign
	env["PROBOD_ESIGN_TSA_URL"] = "http://custom.tsa.example.com"
	// Branding
	env["PROBOD_BRANDING"] = "false"

	b := NewBuilder(NewResolver(mockEnv(env)))
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
	assert.Equal(t, 20000, cfg.Probod.Api.GraphQL.ParserTokenLimit)
	assert.Equal(t, 5000, cfg.Probod.Api.GraphQL.ComplexityLimit)
	assert.Equal(t, 2000, cfg.Probod.Api.GraphQL.QueryCacheSize)
	assert.False(t, cfg.Probod.Api.GraphQL.DisableSuggestion)
	// PG
	assert.Equal(t, "postgres.example.com:5432", cfg.Probod.Pg.Addr)
	assert.Equal(t, "probo", cfg.Probod.Pg.Username)
	assert.Equal(t, "secret123", cfg.Probod.Pg.Password)
	assert.Equal(t, "probo_prod", cfg.Probod.Pg.Database)
	assert.Equal(t, int32(200), cfg.Probod.Pg.PoolSize)
	assert.Equal(t, int32(25), cfg.Probod.Pg.MinPoolSize)
	assert.Equal(t, 900, cfg.Probod.Pg.MaxConnIdleTimeSeconds)
	assert.Equal(t, 7200, cfg.Probod.Pg.MaxConnLifetimeSeconds)
	assert.Equal(t, 600, cfg.Probod.Pg.MaxConnLifetimeJitterSeconds)
	assert.Equal(t, 30, cfg.Probod.Pg.HealthCheckPeriodSeconds)
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
	assert.Equal(t, 120, cfg.Probod.Notifications.Document.Interval)
	assert.Equal(t, 60, cfg.Probod.Notifications.Document.DebounceDelay)
	assert.Equal(t, 43200, cfg.Probod.Notifications.Document.ReminderInterval)
	// Agents tools — Firecrawl
	assert.Equal(t, "fc-test-key", cfg.Probod.Agents.Tools.FirecrawlAPIKey)
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
	// Agents — third-party-vetter overrides
	assert.Equal(t, "openai", cfg.Probod.Agents.ThirdPartyVetter.Provider)
	assert.Equal(t, "gpt-4o", cfg.Probod.Agents.ThirdPartyVetter.ModelName)
	assert.Equal(t, new(0.3), cfg.Probod.Agents.ThirdPartyVetter.Temperature)
	assert.Equal(t, new(8192), cfg.Probod.Agents.ThirdPartyVetter.MaxTokens)
	// Agents — third-party-disambiguation overrides
	assert.Equal(t, "anthropic", cfg.Probod.Agents.ThirdPartyDisambiguation.Provider)
	assert.Equal(t, "claude-sonnet-4-20250514", cfg.Probod.Agents.ThirdPartyDisambiguation.ModelName)
	assert.Equal(t, new(0.4), cfg.Probod.Agents.ThirdPartyDisambiguation.Temperature)
	assert.Equal(t, new(2048), cfg.Probod.Agents.ThirdPartyDisambiguation.MaxTokens)
	// Agents — tracker-mapping overrides
	assert.Equal(t, "openai", cfg.Probod.Agents.TrackerMapping.Provider)
	assert.Equal(t, "gpt-4o-mini", cfg.Probod.Agents.TrackerMapping.ModelName)
	assert.Equal(t, new(0.1), cfg.Probod.Agents.TrackerMapping.Temperature)
	assert.Equal(t, new(1024), cfg.Probod.Agents.TrackerMapping.MaxTokens)
	// Agents — tracker-enrichment overrides
	assert.Equal(t, "openai", cfg.Probod.Agents.TrackerEnrichment.Provider)
	assert.Equal(t, "gpt-4o", cfg.Probod.Agents.TrackerEnrichment.ModelName)
	assert.Equal(t, new(0.2), cfg.Probod.Agents.TrackerEnrichment.Temperature)
	assert.Equal(t, new(2048), cfg.Probod.Agents.TrackerEnrichment.MaxTokens)
	// Tracker worker tuning — overrides
	assert.Equal(t, 20, cfg.Probod.TrackerMappingWorker.Interval)
	assert.Equal(t, 5, cfg.Probod.TrackerMappingWorker.MaxConcurrency)
	assert.Equal(t, 1200, cfg.Probod.TrackerMappingWorker.StaleAfter)
	assert.Equal(t, 30, cfg.Probod.TrackerMappingWorker.AgentTimeout)
	assert.Equal(t, 6, cfg.Probod.TrackerMappingWorker.AgentMaxTurns)
	assert.Equal(t, 35, cfg.Probod.TrackerMappingWorker.DisambiguationAgentTimeout)
	assert.Equal(t, 15, cfg.Probod.CommonPatternEnrichmentWorker.Interval)
	assert.Equal(t, 4, cfg.Probod.CommonPatternEnrichmentWorker.MaxConcurrency)
	assert.Equal(t, 900, cfg.Probod.CommonPatternEnrichmentWorker.StaleAfter)
	assert.Equal(t, 50, cfg.Probod.CommonPatternEnrichmentWorker.AgentTimeout)
	assert.Equal(t, 5, cfg.Probod.CommonPatternEnrichmentWorker.AgentMaxTurns)
	assert.Equal(t, "openai", cfg.Probod.Agents.CommonThirdPartyEnrichment.Provider)
	assert.Equal(t, "gpt-4o", cfg.Probod.Agents.CommonThirdPartyEnrichment.ModelName)
	require.NotNil(t, cfg.Probod.Agents.CommonThirdPartyEnrichment.MaxTokens)
	assert.Equal(t, 16384, *cfg.Probod.Agents.CommonThirdPartyEnrichment.MaxTokens)
	assert.Equal(t, 25, cfg.Probod.CommonThirdPartyEnrichmentWorker.Interval)
	assert.Equal(t, 2, cfg.Probod.CommonThirdPartyEnrichmentWorker.MaxConcurrency)
	assert.Equal(t, 1200, cfg.Probod.CommonThirdPartyEnrichmentWorker.StaleAfter)
	assert.Equal(t, 120, cfg.Probod.CommonThirdPartyEnrichmentWorker.AgentTimeout)
	assert.Equal(t, 8, cfg.Probod.CommonThirdPartyEnrichmentWorker.AgentMaxTurns)
	assert.Equal(t, 0.85, cfg.Probod.CommonThirdPartyEnrichmentWorker.ConfidenceThreshold)
	assert.Equal(t, 5, cfg.Probod.CommonThirdPartyEnrichmentWorker.MaxAttempts)
	assert.Equal(t, 15, cfg.Probod.ThirdPartyVetting.Interval)
	assert.Equal(t, 1800, cfg.Probod.ThirdPartyVetting.StaleAfter)
	assert.Equal(t, 2, cfg.Probod.ThirdPartyVetting.MaxConcurrency)
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

func TestBuilder_Build_GoogleWorkspaceConnector(t *testing.T) {
	env := requiredEnv()
	env["PROBOD_CONNECTOR_GOOGLE_WORKSPACE_CLIENT_ID"] = "gw-client-id"
	env["PROBOD_CONNECTOR_GOOGLE_WORKSPACE_CLIENT_SECRET"] = "gw-client-secret"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.Probod.Connectors, 1)
	connector := cfg.Probod.Connectors[0]
	assert.Equal(t, "GOOGLE_WORKSPACE", connector.Provider)
	assert.Equal(t, "oauth2", string(connector.Protocol))
	rawConfig := connector.RawConfig.(probodconfig.ConnectorConfigOAuth2)
	assert.Equal(t, "gw-client-id", rawConfig.ClientID)
	assert.Equal(t, "gw-client-secret", rawConfig.ClientSecret)
}

func TestBuilder_Build_Microsoft365Connector(t *testing.T) {
	env := requiredEnv()
	env["PROBOD_CONNECTOR_MICROSOFT_365_CLIENT_ID"] = "ms365-client-id"
	env["PROBOD_CONNECTOR_MICROSOFT_365_CLIENT_SECRET"] = "ms365-client-secret"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.Probod.Connectors, 1)
	connector := cfg.Probod.Connectors[0]
	assert.Equal(t, "MICROSOFT_365", connector.Provider)
	assert.Equal(t, "oauth2", string(connector.Protocol))
	rawConfig := connector.RawConfig.(probodconfig.ConnectorConfigOAuth2)
	assert.Equal(t, "ms365-client-id", rawConfig.ClientID)
	assert.Equal(t, "ms365-client-secret", rawConfig.ClientSecret)
}

func TestBuilder_Build_AccessReviewConnectors(t *testing.T) {
	// All non-Vercel access-review providers added by this PR. Vercel
	// has its own dedicated test because it carries an additional
	// CONNECTOR_VERCEL_INTEGRATION_SLUG env var.
	providers := []string{
		"GITLAB", "BITBUCKET", "HEROKU", "PAGERDUTY",
		"ASANA", "NETLIFY", "CLICKUP", "MONDAY", "DATADOG",
		"ZENDESK", "LINEAR", "GOOGLE_ANALYTICS", "SQUARE",
	}

	env := requiredEnv()
	for _, provider := range providers {
		env["PROBOD_CONNECTOR_"+provider+"_CLIENT_ID"] = strings.ToLower(provider) + "-id"
		env["PROBOD_CONNECTOR_"+provider+"_CLIENT_SECRET"] = strings.ToLower(provider) + "-secret"
	}

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.Probod.Connectors, len(providers))

	byProvider := make(map[string]probodconfig.ConnectorConfig, len(cfg.Probod.Connectors))
	for _, c := range cfg.Probod.Connectors {
		byProvider[c.Provider] = c
	}

	for _, provider := range providers {
		c, ok := byProvider[provider]
		require.True(t, ok, "missing %s connector", provider)
		assert.Equal(t, "oauth2", string(c.Protocol))
		raw := c.RawConfig.(probodconfig.ConnectorConfigOAuth2)
		assert.NotEmpty(t, raw.ClientID, "%s client-id", provider)
		assert.NotEmpty(t, raw.ClientSecret, "%s client-secret", provider)
		assert.Empty(t, raw.IntegrationSlug, "%s should not carry integration-slug", provider)
	}
}

func TestBuilder_Build_VercelConnector(t *testing.T) {
	env := requiredEnv()
	env["PROBOD_CONNECTOR_VERCEL_CLIENT_ID"] = "vercel-id"
	env["PROBOD_CONNECTOR_VERCEL_CLIENT_SECRET"] = "vercel-secret"
	env["PROBOD_CONNECTOR_VERCEL_INTEGRATION_SLUG"] = "probo-app"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.Probod.Connectors, 1)
	c := cfg.Probod.Connectors[0]
	assert.Equal(t, "VERCEL", c.Provider)
	assert.Equal(t, "oauth2", string(c.Protocol))
	raw := c.RawConfig.(probodconfig.ConnectorConfigOAuth2)
	assert.Equal(t, "vercel-id", raw.ClientID)
	assert.Equal(t, "vercel-secret", raw.ClientSecret)
	assert.Equal(t, "probo-app", raw.IntegrationSlug)
}

func TestBuilder_Build_SlackConnector(t *testing.T) {
	env := requiredEnv()
	env["PROBOD_CONNECTOR_SLACK_CLIENT_ID"] = "slack-client-id"
	env["PROBOD_CONNECTOR_SLACK_CLIENT_SECRET"] = "slack-client-secret"
	env["PROBOD_CONNECTOR_SLACK_SIGNING_SECRET"] = "slack-signing-secret"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.Probod.Connectors, 1)
	connector := cfg.Probod.Connectors[0]
	assert.Equal(t, "SLACK", connector.Provider)
	assert.Equal(t, "oauth2", string(connector.Protocol))
	rawConfig := connector.RawConfig.(probodconfig.ConnectorConfigOAuth2)
	assert.Equal(t, "slack-client-id", rawConfig.ClientID)
	assert.Equal(t, "slack-client-secret", rawConfig.ClientSecret)

	rawSettings := connector.RawSettings.(map[string]any)
	assert.Equal(t, "slack-signing-secret", rawSettings["signing-secret"])
}

func TestBuilder_Build_CrispConnector(t *testing.T) {
	env := requiredEnv()
	env["PROBOD_CONNECTOR_CRISP_PLUGIN_TOKEN"] = "plugin-identifier:plugin-key"
	env["PROBOD_CONNECTOR_CRISP_PLUGIN_ID"] = "e979a1c3-2c41-4e93-a8ed-410ace27318e"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.Probod.Connectors, 1)
	connector := cfg.Probod.Connectors[0]
	assert.Equal(t, "CRISP", connector.Provider)
	assert.Equal(t, "api_key", string(connector.Protocol))
	rawConfig := connector.RawConfig.(probodconfig.ConnectorConfigAPIKey)
	assert.Equal(t, "plugin-identifier:plugin-key", rawConfig.APIKey)
	assert.Equal(t, "e979a1c3-2c41-4e93-a8ed-410ace27318e", rawConfig.ResourceID)
}

func TestBuilder_Build_CrispConnectorAbsentWithoutToken(t *testing.T) {
	// Without the plugin token the connector must not be configured, which
	// is what keeps Crisp deactivated until Crisp validates the plugin.
	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	for _, c := range cfg.Probod.Connectors {
		assert.NotEqual(t, "CRISP", c.Provider)
	}
}

func TestBuilder_Build_CrispConnectorAbsentWithoutPluginID(t *testing.T) {
	// The plugin token alone is not enough: the per-website plugin API needs
	// the plugin ID to verify website ownership, so a half-configured Crisp
	// connector stays hidden rather than activating in a broken state.
	env := requiredEnv()
	env["PROBOD_CONNECTOR_CRISP_PLUGIN_TOKEN"] = "plugin-identifier:plugin-key"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = "test-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	for _, c := range cfg.Probod.Connectors {
		assert.NotEqual(t, "CRISP", c.Provider)
	}
}

func TestBuilder_Build_SAMLAutoGeneration(t *testing.T) {
	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Contains(t, cfg.Probod.Auth.SAML.Certificate, "-----BEGIN CERTIFICATE-----")
	assert.Contains(t, cfg.Probod.Auth.SAML.Certificate, "-----END CERTIFICATE-----")
	assert.Contains(t, cfg.Probod.Auth.SAML.PrivateKey, "-----BEGIN RSA PRIVATE KEY-----")
	assert.Contains(t, cfg.Probod.Auth.SAML.PrivateKey, "-----END RSA PRIVATE KEY-----")
}

func TestBuilder_Build_SAMLFromEnv(t *testing.T) {
	env := requiredEnv()
	env["PROBOD_SAML_CERTIFICATE"] = "env-cert"
	env["PROBOD_SAML_PRIVATE_KEY"] = "env-key"

	b := NewBuilder(NewResolver(mockEnv(env)))

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "env-cert", cfg.Probod.Auth.SAML.Certificate)
	assert.Equal(t, "env-key", cfg.Probod.Auth.SAML.PrivateKey)
}

func TestBuilder_Build_SAMLPreset(t *testing.T) {
	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))
	b.samlCertificate = "preset-cert"
	b.samlPrivateKey = "preset-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "preset-cert", cfg.Probod.Auth.SAML.Certificate)
	assert.Equal(t, "preset-key", cfg.Probod.Auth.SAML.PrivateKey)
}

func TestBuilder_Build_OAuth2Defaults(t *testing.T) {
	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))

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
	assert.Nil(t, cfg.Probod.Auth.OAuth2Server.CIMDAllowedClientIDs)
}

func TestBuilder_Build_OAuth2FromEnv(t *testing.T) {
	env := requiredEnv()
	env["PROBOD_OAUTH2_SERVER_SIGNING_KEY"] = "env-signing-key"
	env["PROBOD_OAUTH2_SERVER_SIGNING_KEY_KID"] = "env-kid"
	env["PROBOD_OAUTH2_SERVER_ACCESS_TOKEN_DURATION"] = "10"
	env["PROBOD_OAUTH2_SERVER_REFRESH_TOKEN_DURATION"] = "20"
	env["PROBOD_OAUTH2_SERVER_AUTHORIZATION_CODE_DURATION"] = "30"
	env["PROBOD_OAUTH2_SERVER_DEVICE_CODE_DURATION"] = "40"
	env["PROBOD_OAUTH2_SERVER_CIMD_ALLOWED_CLIENT_IDS"] = "https://chatgpt.com/oauth/client.json,https://claude.ai/oauth/client.json"

	b := NewBuilder(NewResolver(mockEnv(env)))

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
	assert.Equal(
		t,
		[]string{
			"https://chatgpt.com/oauth/client.json",
			"https://claude.ai/oauth/client.json",
		},
		cfg.Probod.Auth.OAuth2Server.CIMDAllowedClientIDs,
	)
}

func TestBuilder_Build_OAuth2Preset(t *testing.T) {
	env := requiredEnv()
	delete(env, "OAUTH2_SERVER_SIGNING_KEY")

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.oauth2SigningKey = "preset-signing-key"

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.Probod.Auth.OAuth2Server.SigningKeys, 1)
	assert.Equal(t, "preset-signing-key", cfg.Probod.Auth.OAuth2Server.SigningKeys[0].PrivateKey)
}

func TestBuilder_Build_PgCABundleFromEnv(t *testing.T) {
	env := requiredEnv()
	env["PROBOD_PG_CA_BUNDLE"] = "test-ca-bundle-content"

	b := NewBuilder(NewResolver(mockEnv(env)))
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
	env["PROBOD_PG_CA_BUNDLE_PATH"] = caFile

	b := NewBuilder(NewResolver(mockEnv(env)))
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
