// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package bootstrap

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/probodconfig"
)

// connectorEndpointFields are the field segments buildConnectorEndpoints reads
// off PROBOD_CONNECTOR_<PROVIDER>_ENDPOINT_<FIELD>. checkConnectorEndpointEnvTypos
// keys off the same list so the two never drift apart.
var connectorEndpointFields = []string{"AUTH", "TOKEN", "PROBE", "IDENTITY", "API_BASE"}

type EnvGetter func(key string) string

type Builder struct {
	resolver         *Resolver
	samlCertificate  string
	samlPrivateKey   string
	oauth2SigningKey string
}

func NewBuilder(resolver *Resolver) *Builder {
	if resolver == nil {
		resolver = NewResolver(nil)
	}

	return &Builder{resolver: resolver}
}

func (b *Builder) Build() (*probodconfig.FullConfig, error) {
	if err := b.validateRequired(); err != nil {
		return nil, err
	}

	samlCert, samlKey, err := b.getSAMLCredentials()
	if err != nil {
		return nil, fmt.Errorf("cannot get SAML credentials: %w", err)
	}

	oauth2SigningKey := b.getOAuth2SigningKey()

	pgCACertBundle := b.getPgCACertBundle()

	authCookieSameSite, err := probodconfig.ParseCookieSameSite(
		b.resolver.getEnvOrDefault("PROBOD_AUTH_COOKIE_SAMESITE", "lax"),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot parse PROBOD_AUTH_COOKIE_SAMESITE: %w", err)
	}

	compliancePortalTLSMode, err := probodconfig.ParseCompliancePortalTLSMode(
		b.resolver.getEnvOrDefault("PROBOD_TRUST_CENTER_TLS_MODE", "direct"),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot parse PROBOD_TRUST_CENTER_TLS_MODE: %w", err)
	}

	cfg := &probodconfig.FullConfig{
		Unit: probodconfig.UnitConfig{
			Metrics: probodconfig.MetricsConfig{
				Addr: b.resolver.getEnvOrDefault("PROBOD_METRICS_ADDR", "localhost:8081"),
			},
			Tracing: probodconfig.TracingConfig{
				Addr:          b.resolver.getEnvOrDefault("PROBOD_TRACING_ADDR", "localhost:4318"),
				MaxBatchSize:  b.resolver.getEnvIntOrDefault("PROBOD_TRACING_MAX_BATCH_SIZE", 512),
				BatchTimeout:  b.resolver.getEnvIntOrDefault("PROBOD_TRACING_BATCH_TIMEOUT", 5),
				ExportTimeout: b.resolver.getEnvIntOrDefault("PROBOD_TRACING_EXPORT_TIMEOUT", 30),
				MaxQueueSize:  b.resolver.getEnvIntOrDefault("PROBOD_TRACING_MAX_QUEUE_SIZE", 2048),
			},
		},
		Probod: probodconfig.Config{
			BaseURL:       b.resolver.getEnv("PROBOD_BASE_URL"),
			EncryptionKey: b.resolver.getEnv("PROBOD_ENCRYPTION_KEY"),
			ChromeDPAddr:  b.resolver.getEnv("PROBOD_CHROME_DP_ADDR"),
			Api: probodconfig.APIConfig{
				Addr: b.resolver.getEnv("PROBOD_API_ADDR"),
				ProxyProtocol: probodconfig.ProxyProtocolConfig{
					TrustedProxies: b.parseOriginsList(b.resolver.getEnv("PROBOD_API_PROXY_PROTOCOL_TRUSTED_PROXIES")),
				},
				Cors: probodconfig.CorsConfig{
					AllowedOrigins: b.parseOriginsList(b.resolver.getEnv("PROBOD_API_CORS_ALLOWED_ORIGINS")),
				},
				ExtraHeaderFields: nil,
				GraphQL: probodconfig.GraphQLConfig{
					ParserTokenLimit:  b.resolver.getEnvIntOrDefault("PROBOD_API_GRAPHQL_PARSER_TOKEN_LIMIT", 15000),
					ComplexityLimit:   b.resolver.getEnvIntOrDefault("PROBOD_API_GRAPHQL_COMPLEXITY_LIMIT", 2000),
					QueryCacheSize:    b.resolver.getEnvIntOrDefault("PROBOD_API_GRAPHQL_QUERY_CACHE_SIZE", 1000),
					DisableSuggestion: b.resolver.getEnvBoolOrDefault("PROBOD_API_GRAPHQL_DISABLE_SUGGESTION", true),
				},
			},
			Pg: probodconfig.PgConfig{
				Addr:                         b.resolver.getEnv("PROBOD_PG_ADDR"),
				Username:                     b.resolver.getEnv("PROBOD_PG_USERNAME"),
				Password:                     b.resolver.getEnv("PROBOD_PG_PASSWORD"),
				Database:                     b.resolver.getEnv("PROBOD_PG_DATABASE"),
				PoolSize:                     int32(b.resolver.getEnvIntOrDefault("PROBOD_PG_POOL_SIZE", 100)),
				MinPoolSize:                  int32(b.resolver.getEnvIntOrDefault("PROBOD_PG_MIN_POOL_SIZE", 10)),
				MaxConnIdleTimeSeconds:       b.resolver.getEnvIntOrDefault("PROBOD_PG_MAX_CONN_IDLE_TIME_SECONDS", 1800),
				MaxConnLifetimeSeconds:       b.resolver.getEnvIntOrDefault("PROBOD_PG_MAX_CONN_LIFETIME_SECONDS", 3600),
				MaxConnLifetimeJitterSeconds: b.resolver.getEnvIntOrDefault("PROBOD_PG_MAX_CONN_LIFETIME_JITTER_SECONDS", 300),
				HealthCheckPeriodSeconds:     b.resolver.getEnvIntOrDefault("PROBOD_PG_HEALTH_CHECK_PERIOD_SECONDS", 60),
				CACertBundle:                 pgCACertBundle,
				Debug:                        b.resolver.getEnvBoolOrDefault("PROBOD_PG_DEBUG", false),
			},
			Auth: probodconfig.AuthConfig{
				DisableSignup:                       b.resolver.getEnvBoolOrDefault("PROBOD_AUTH_DISABLE_SIGNUP", false),
				InvitationConfirmationTokenValidity: b.resolver.getEnvIntOrDefault("PROBOD_AUTH_INVITATION_TOKEN_VALIDITY", 3600),
				PasswordResetTokenValidity:          b.resolver.getEnvIntOrDefault("PROBOD_AUTH_PASSWORD_RESET_TOKEN_VALIDITY", 3600),
				MagicLinkTokenValidity:              b.resolver.getEnvIntOrDefault("PROBOD_AUTH_MAGIC_LINK_TOKEN_VALIDITY", 900),
				EmailConfirmationTokenValidity:      b.resolver.getEnvIntOrDefault("PROBOD_AUTH_EMAIL_CONFIRMATION_TOKEN_VALIDITY", 3600),
				Cookie: probodconfig.CookieConfig{
					Name:     b.resolver.getEnv("PROBOD_AUTH_COOKIE_NAME"),
					Domain:   b.resolver.getEnv("PROBOD_AUTH_COOKIE_DOMAIN"),
					Secret:   b.resolver.getEnv("PROBOD_AUTH_COOKIE_SECRET"),
					Duration: b.resolver.getEnvIntOrDefault("PROBOD_AUTH_COOKIE_DURATION", 24),
					Secure:   b.resolver.getEnvBoolOrDefault("PROBOD_AUTH_COOKIE_SECURE", true),
					SameSite: authCookieSameSite,
				},
				Password: probodconfig.PasswordConfig{
					Pepper:     b.resolver.getEnv("PROBOD_AUTH_PASSWORD_PEPPER"),
					Iterations: b.resolver.getEnvIntOrDefault("PROBOD_AUTH_PASSWORD_ITERATIONS", 1000000),
				},
				SAML: probodconfig.SAMLConfig{
					SessionDuration:                   b.resolver.getEnvIntOrDefault("PROBOD_SAML_SESSION_DURATION", 604800),
					CleanupIntervalSeconds:            b.resolver.getEnvIntOrDefault("PROBOD_SAML_CLEANUP_INTERVAL_SECONDS", 0),
					Certificate:                       samlCert,
					PrivateKey:                        samlKey,
					DomainVerificationIntervalSeconds: b.resolver.getEnvIntOrDefault("PROBOD_SAML_DOMAIN_VERIFICATION_INTERVAL_SECONDS", 60),
					DomainVerificationResolverAddr:    b.resolver.getEnv("PROBOD_SAML_DOMAIN_VERIFICATION_RESOLVER_ADDR"),
				},
				Google: probodconfig.OIDCProviderConfig{
					ClientID:     b.resolver.getEnv("PROBOD_AUTH_GOOGLE_CLIENT_ID"),
					ClientSecret: b.resolver.getEnv("PROBOD_AUTH_GOOGLE_CLIENT_SECRET"),
					Enabled:      b.resolver.getEnv("PROBOD_AUTH_GOOGLE_CLIENT_ID") != "" && b.resolver.getEnv("PROBOD_AUTH_GOOGLE_CLIENT_SECRET") != "",
				},
				Microsoft: probodconfig.OIDCProviderConfig{
					ClientID:     b.resolver.getEnv("PROBOD_AUTH_MICROSOFT_CLIENT_ID"),
					ClientSecret: b.resolver.getEnv("PROBOD_AUTH_MICROSOFT_CLIENT_SECRET"),
					Enabled:      b.resolver.getEnv("PROBOD_AUTH_MICROSOFT_CLIENT_ID") != "" && b.resolver.getEnv("PROBOD_AUTH_MICROSOFT_CLIENT_SECRET") != "",
				},
				OAuth2Server: probodconfig.OAuth2ServerConfig{
					SigningKeys: []probodconfig.OAuth2SigningKeyConfig{{
						PrivateKey: oauth2SigningKey,
						KID:        b.resolver.getEnvOrDefault("PROBOD_OAUTH2_SERVER_SIGNING_KEY_KID", "default"),
						Active:     true,
					}},
					AccessTokenDuration:       b.resolver.getEnvIntOrDefault("PROBOD_OAUTH2_SERVER_ACCESS_TOKEN_DURATION", 3600),
					RefreshTokenDuration:      b.resolver.getEnvIntOrDefault("PROBOD_OAUTH2_SERVER_REFRESH_TOKEN_DURATION", 2592000),
					AuthorizationCodeDuration: b.resolver.getEnvIntOrDefault("PROBOD_OAUTH2_SERVER_AUTHORIZATION_CODE_DURATION", 600),
					DeviceCodeDuration:        b.resolver.getEnvIntOrDefault("PROBOD_OAUTH2_SERVER_DEVICE_CODE_DURATION", 600),
					CIMDAllowedClientIDs: b.parseOriginsList(
						b.resolver.getEnv("PROBOD_OAUTH2_SERVER_CIMD_ALLOWED_CLIENT_IDS"),
					),
				},
			},
			ITAM: probodconfig.ITAMConfig{
				DeviceEnrollmentTokenValidity: b.resolver.getEnvIntOrDefault(
					"PROBOD_ITAM_DEVICE_ENROLLMENT_TOKEN_VALIDITY",
					604800,
				),
			},
			CompliancePortal: probodconfig.CompliancePortalConfig{
				HTTPAddr:   b.resolver.getEnv("PROBOD_TRUST_CENTER_HTTP_ADDR"),
				HTTPSAddr:  b.resolver.getEnv("PROBOD_TRUST_CENTER_HTTPS_ADDR"),
				BaseDomain: b.resolver.getEnv("PROBOD_TRUST_CENTER_BASE_DOMAIN"),
				TLSMode:    compliancePortalTLSMode,
				ProxyProtocol: probodconfig.ProxyProtocolConfig{
					TrustedProxies: b.parseOriginsList(b.resolver.getEnv("PROBOD_TRUST_CENTER_PROXY_PROTOCOL_TRUSTED_PROXIES")),
				},
			},
			AWS: probodconfig.AWSConfig{
				Region:          b.resolver.getEnv("PROBOD_AWS_REGION"),
				Bucket:          b.resolver.getEnv("PROBOD_AWS_BUCKET"),
				AccessKeyID:     b.resolver.getEnv("PROBOD_AWS_ACCESS_KEY_ID"),
				SecretAccessKey: b.resolver.getEnv("PROBOD_AWS_SECRET_ACCESS_KEY"),
				Endpoint:        b.resolver.getEnv("PROBOD_AWS_ENDPOINT"),
				UsePathStyle:    b.resolver.getEnvBoolOrDefault("PROBOD_AWS_USE_PATH_STYLE", false),
			},
			Notifications: probodconfig.NotificationsConfig{
				Mailer: probodconfig.MailerConfig{
					SenderName:     b.resolver.getEnv("PROBOD_MAILER_SENDER_NAME"),
					SenderEmail:    b.resolver.getEnv("PROBOD_MAILER_SENDER_EMAIL"),
					MailerInterval: b.resolver.getEnvIntOrDefault("PROBOD_MAILER_INTERVAL", 60),
					SMTP: probodconfig.SMTPConfig{
						Addr:        b.resolver.getEnv("PROBOD_SMTP_ADDR"),
						User:        b.resolver.getEnv("PROBOD_SMTP_USER"),
						Password:    b.resolver.getEnv("PROBOD_SMTP_PASSWORD"),
						TLSRequired: b.resolver.getEnvBoolOrDefault("PROBOD_SMTP_TLS_REQUIRED", false),
						HelloName:   b.resolver.getEnv("PROBOD_SMTP_HELLO_NAME"),
					},
				},
				Slack: probodconfig.SlackConfig{
					SenderInterval: b.resolver.getEnvIntOrDefault("PROBOD_SLACK_SENDER_INTERVAL", 60),
					SigningSecret:  b.resolver.getEnv("PROBOD_CONNECTOR_SLACK_SIGNING_SECRET"),
				},
				Webhook: probodconfig.WebhookConfig{
					SenderInterval: b.resolver.getEnvIntOrDefault("PROBOD_WEBHOOK_SENDER_INTERVAL", 5),
					CacheTTL:       b.resolver.getEnvIntOrDefault("PROBOD_WEBHOOK_CACHE_TTL", 86400),
				},
				Document: probodconfig.DocumentNotificationConfig{
					Interval:         b.resolver.getEnvIntOrDefault("PROBOD_DOCUMENT_NOTIFICATION_INTERVAL", 300),
					DebounceDelay:    b.resolver.getEnvIntOrDefault("PROBOD_DOCUMENT_NOTIFICATION_DEBOUNCE_DELAY", 900),
					ReminderInterval: b.resolver.getEnvIntOrDefault("PROBOD_DOCUMENT_NOTIFICATION_REMINDER_INTERVAL", 86400),
				},
			},
			Agents: func() probodconfig.AgentsConfig {
				defaultProvider := b.resolver.getEnvOrDefault("PROBOD_AGENT_DEFAULT_PROVIDER", "openai")

				return probodconfig.AgentsConfig{
					Providers: b.buildLLMProviders(),
					Default: probodconfig.LLMAgentConfig{
						Provider:    defaultProvider,
						ModelName:   b.resolver.getEnvOrDefault("PROBOD_AGENT_DEFAULT_MODEL_NAME", "gpt-4o"),
						Temperature: new(b.resolver.getEnvFloatOrDefault("PROBOD_AGENT_DEFAULT_TEMPERATURE", 0.1)),
						MaxTokens:   new(b.resolver.getEnvIntOrDefault("PROBOD_AGENT_DEFAULT_MAX_TOKENS", 4096)),
					},
					Probo: probodconfig.LLMAgentConfig{
						Provider:    b.resolver.getEnvOrDefault("PROBOD_AGENT_PROBO_PROVIDER", ""),
						ModelName:   b.resolver.getEnvOrDefault("PROBOD_AGENT_PROBO_MODEL_NAME", ""),
						Temperature: b.resolver.getEnvFloatPtr("PROBOD_AGENT_PROBO_TEMPERATURE"),
						MaxTokens:   b.resolver.getEnvIntPtr("PROBOD_AGENT_PROBO_MAX_TOKENS"),
					},
					EvidenceDescriber: probodconfig.LLMAgentConfig{
						Provider:    b.resolver.getEnvOrDefault("PROBOD_AGENT_EVIDENCE_DESCRIBER_PROVIDER", ""),
						ModelName:   b.resolver.getEnvOrDefault("PROBOD_AGENT_EVIDENCE_DESCRIBER_MODEL_NAME", ""),
						Temperature: b.resolver.getEnvFloatPtr("PROBOD_AGENT_EVIDENCE_DESCRIBER_TEMPERATURE"),
						MaxTokens:   b.resolver.getEnvIntPtr("PROBOD_AGENT_EVIDENCE_DESCRIBER_MAX_TOKENS"),
					},
					ThirdPartyVetter: probodconfig.LLMAgentConfig{
						Provider:    b.resolver.getEnvOrDefault("PROBOD_AGENT_THIRD_PARTY_VETTER_PROVIDER", ""),
						ModelName:   b.resolver.getEnvOrDefault("PROBOD_AGENT_THIRD_PARTY_VETTER_MODEL_NAME", ""),
						Temperature: b.resolver.getEnvFloatPtr("PROBOD_AGENT_THIRD_PARTY_VETTER_TEMPERATURE"),
						MaxTokens:   b.resolver.getEnvIntPtr("PROBOD_AGENT_THIRD_PARTY_VETTER_MAX_TOKENS"),
					},
					ThirdPartyDisambiguation: probodconfig.LLMAgentConfig{
						Provider:  b.resolver.getEnvOrDefault("PROBOD_AGENT_THIRD_PARTY_DISAMBIGUATION_PROVIDER", ""),
						ModelName: b.resolver.getEnvOrDefault("PROBOD_AGENT_THIRD_PARTY_DISAMBIGUATION_MODEL_NAME", ""),
						// The disambiguation agent emits a single id plus a
						// short rationale, but the budget must leave headroom
						// for reasoning models whose reasoning tokens count
						// against max_tokens; too small a budget truncates the
						// JSON.
						Temperature: b.resolver.getEnvFloatPtr("PROBOD_AGENT_THIRD_PARTY_DISAMBIGUATION_TEMPERATURE"),
						MaxTokens:   new(b.resolver.getEnvIntOrDefault("PROBOD_AGENT_THIRD_PARTY_DISAMBIGUATION_MAX_TOKENS", 4096)),
					},
					TrackerMapping: probodconfig.LLMAgentConfig{
						Provider:  b.resolver.getEnvOrDefault("PROBOD_AGENT_TRACKER_MAPPING_PROVIDER", ""),
						ModelName: b.resolver.getEnvOrDefault("PROBOD_AGENT_TRACKER_MAPPING_MODEL_NAME", ""),
						// The tracker agents emit tiny structured JSON, but
						// the budget must leave headroom for reasoning
						// models whose reasoning tokens count against
						// max_tokens; too small a budget truncates the JSON.
						Temperature: b.resolver.getEnvFloatPtr("PROBOD_AGENT_TRACKER_MAPPING_TEMPERATURE"),
						MaxTokens:   new(b.resolver.getEnvIntOrDefault("PROBOD_AGENT_TRACKER_MAPPING_MAX_TOKENS", 4096)),
					},
					TrackerEnrichment: probodconfig.LLMAgentConfig{
						Provider:  b.resolver.getEnvOrDefault("PROBOD_AGENT_TRACKER_ENRICHMENT_PROVIDER", ""),
						ModelName: b.resolver.getEnvOrDefault("PROBOD_AGENT_TRACKER_ENRICHMENT_MODEL_NAME", ""),
						// See the tracker-mapping note: keep ample headroom so
						// reasoning models do not truncate the structured JSON.
						Temperature: b.resolver.getEnvFloatPtr("PROBOD_AGENT_TRACKER_ENRICHMENT_TEMPERATURE"),
						MaxTokens:   new(b.resolver.getEnvIntOrDefault("PROBOD_AGENT_TRACKER_ENRICHMENT_MAX_TOKENS", 4096)),
					},
					CommonThirdPartyEnrichment: probodconfig.LLMAgentConfig{
						Provider:  b.resolver.getEnvOrDefault("PROBOD_AGENT_COMMON_THIRD_PARTY_ENRICHMENT_PROVIDER", ""),
						ModelName: b.resolver.getEnvOrDefault("PROBOD_AGENT_COMMON_THIRD_PARTY_ENRICHMENT_MODEL_NAME", ""),
						// Agent B browses pages and emits a moderate structured
						// output; the budget must leave headroom for reasoning
						// models whose reasoning tokens count against max_tokens.
						Temperature: b.resolver.getEnvFloatPtr("PROBOD_AGENT_COMMON_THIRD_PARTY_ENRICHMENT_TEMPERATURE"),
						MaxTokens:   new(b.resolver.getEnvIntOrDefault("PROBOD_AGENT_COMMON_THIRD_PARTY_ENRICHMENT_MAX_TOKENS", 8192)),
					},
					Tools: probodconfig.AgentToolsConfig{
						FirecrawlAPIKey: b.resolver.getEnv("PROBOD_FIRECRAWL_API_KEY"),
					},
				}
			}(),
			CustomDomains: probodconfig.CustomDomainsConfig{
				RenewalInterval:   b.resolver.getEnvIntOrDefault("PROBOD_CUSTOM_DOMAINS_RENEWAL_INTERVAL", 3600),
				ProvisionInterval: b.resolver.getEnvIntOrDefault("PROBOD_CUSTOM_DOMAINS_PROVISION_INTERVAL", 30),
				CnameTarget:       b.resolver.getEnvOrDefault("PROBOD_CUSTOM_DOMAINS_CNAME_TARGET", "custom.getprobo.com"),
				ResolverAddr:      b.resolver.getEnv("PROBOD_CUSTOM_DOMAINS_RESOLVER_ADDR"),
				CAAIssuerDomain:   b.resolver.getEnvOrDefault("PROBOD_CUSTOM_DOMAINS_CAA_ISSUER_DOMAIN", "letsencrypt.org"),
				ACME: probodconfig.ACMEConfig{
					Directory:  b.resolver.getEnv("PROBOD_ACME_DIRECTORY"),
					Email:      b.resolver.getEnv("PROBOD_ACME_EMAIL"),
					KeyType:    b.resolver.getEnv("PROBOD_ACME_KEY_TYPE"),
					RootCA:     b.resolver.getEnv("PROBOD_ACME_ROOT_CA"),
					AccountKey: b.resolver.getEnv("PROBOD_ACME_ACCOUNT_KEY"),
				},
			},
			SCIMBridge: probodconfig.SCIMBridgeConfig{
				SyncInterval: b.resolver.getEnvIntOrDefault("PROBOD_SCIM_BRIDGE_SYNC_INTERVAL", 900),
				PollInterval: b.resolver.getEnvIntOrDefault("PROBOD_SCIM_BRIDGE_POLL_INTERVAL", 30),
			},
			ESign: probodconfig.ESignConfig{
				TSAURL: b.resolver.getEnv("PROBOD_ESIGN_TSA_URL"),
			},
			EvidenceDescriber: probodconfig.EvidenceDescriberConfig{
				Interval:       b.resolver.getEnvIntOrDefault("PROBOD_EVIDENCE_DESCRIBER_INTERVAL", 10),
				StaleAfter:     b.resolver.getEnvIntOrDefault("PROBOD_EVIDENCE_DESCRIBER_STALE_AFTER", 300),
				MaxConcurrency: b.resolver.getEnvIntOrDefault("PROBOD_EVIDENCE_DESCRIBER_MAX_CONCURRENCY", 10),
			},
			ThirdPartyVetting: probodconfig.ThirdPartyVettingWorkerConfig{
				Interval:       b.resolver.getEnvIntOrDefault("PROBOD_THIRD_PARTY_VETTING_INTERVAL", 10),
				StaleAfter:     b.resolver.getEnvIntOrDefault("PROBOD_THIRD_PARTY_VETTING_STALE_AFTER", 1500),
				MaxConcurrency: b.resolver.getEnvIntOrDefault("PROBOD_THIRD_PARTY_VETTING_MAX_CONCURRENCY", 1),
			},
			TrackerMappingWorker: probodconfig.TrackerMappingWorkerConfig{
				Interval:                   b.resolver.getEnvIntOrDefault("PROBOD_TRACKER_MAPPING_INTERVAL", 10),
				MaxConcurrency:             b.resolver.getEnvIntOrDefault("PROBOD_TRACKER_MAPPING_MAX_CONCURRENCY", 3),
				StaleAfter:                 b.resolver.getEnvIntOrDefault("PROBOD_TRACKER_MAPPING_STALE_AFTER", 600),
				AgentTimeout:               b.resolver.getEnvIntOrDefault("PROBOD_TRACKER_MAPPING_AGENT_TIMEOUT", 45),
				AgentMaxTurns:              b.resolver.getEnvIntOrDefault("PROBOD_TRACKER_MAPPING_AGENT_MAX_TURNS", 10),
				DisambiguationAgentTimeout: b.resolver.getEnvIntOrDefault("PROBOD_TRACKER_MAPPING_DISAMBIGUATION_AGENT_TIMEOUT", 45),
			},
			CommonPatternEnrichmentWorker: probodconfig.CommonPatternEnrichmentWorkerConfig{
				Interval:       b.resolver.getEnvIntOrDefault("PROBOD_COMMON_PATTERN_ENRICHMENT_INTERVAL", 10),
				MaxConcurrency: b.resolver.getEnvIntOrDefault("PROBOD_COMMON_PATTERN_ENRICHMENT_MAX_CONCURRENCY", 2),
				StaleAfter:     b.resolver.getEnvIntOrDefault("PROBOD_COMMON_PATTERN_ENRICHMENT_STALE_AFTER", 600),
				AgentTimeout:   b.resolver.getEnvIntOrDefault("PROBOD_COMMON_PATTERN_ENRICHMENT_AGENT_TIMEOUT", 45),
				AgentMaxTurns:  b.resolver.getEnvIntOrDefault("PROBOD_COMMON_PATTERN_ENRICHMENT_AGENT_MAX_TURNS", 10),
			},
			CommonThirdPartyEnrichmentWorker: probodconfig.CommonThirdPartyEnrichmentWorkerConfig{
				Interval:            b.resolver.getEnvIntOrDefault("PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_INTERVAL", 10),
				MaxConcurrency:      b.resolver.getEnvIntOrDefault("PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_MAX_CONCURRENCY", 1),
				StaleAfter:          b.resolver.getEnvIntOrDefault("PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_STALE_AFTER", 900),
				AgentTimeout:        b.resolver.getEnvIntOrDefault("PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_AGENT_TIMEOUT", 90),
				AgentMaxTurns:       b.resolver.getEnvIntOrDefault("PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_AGENT_MAX_TURNS", 12),
				ConfidenceThreshold: b.resolver.getEnvFloatOrDefault("PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_CONFIDENCE_THRESHOLD", 0.7),
				MaxAttempts:         b.resolver.getEnvIntOrDefault("PROBOD_COMMON_THIRD_PARTY_ENRICHMENT_MAX_ATTEMPTS", 3),
			},
			Branding: b.resolver.getEnvBoolOrDefault("PROBOD_BRANDING", true),
		},
	}

	if slackClientID := b.resolver.getEnv("PROBOD_CONNECTOR_SLACK_CLIENT_ID"); slackClientID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "SLACK",
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:     slackClientID,
					ClientSecret: b.resolver.getEnv("PROBOD_CONNECTOR_SLACK_CLIENT_SECRET"),
				},
				RawSettings: map[string]any{
					"signing-secret": b.resolver.getEnv("PROBOD_CONNECTOR_SLACK_SIGNING_SECRET"),
				},
			},
		)
	}

	if hubspotClientID := b.resolver.getEnv("PROBOD_CONNECTOR_HUBSPOT_CLIENT_ID"); hubspotClientID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "HUBSPOT",
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:     hubspotClientID,
					ClientSecret: b.resolver.getEnv("PROBOD_CONNECTOR_HUBSPOT_CLIENT_SECRET"),
				},
			},
		)
	}

	if docusignClientID := b.resolver.getEnv("PROBOD_CONNECTOR_DOCUSIGN_CLIENT_ID"); docusignClientID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "DOCUSIGN",
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:     docusignClientID,
					ClientSecret: b.resolver.getEnv("PROBOD_CONNECTOR_DOCUSIGN_CLIENT_SECRET"),
				},
			},
		)
	}

	if notionClientID := b.resolver.getEnv("PROBOD_CONNECTOR_NOTION_CLIENT_ID"); notionClientID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "NOTION",
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:     notionClientID,
					ClientSecret: b.resolver.getEnv("PROBOD_CONNECTOR_NOTION_CLIENT_SECRET"),
				},
			},
		)
	}

	if githubClientID := b.resolver.getEnv("PROBOD_CONNECTOR_GITHUB_CLIENT_ID"); githubClientID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "GITHUB",
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:     githubClientID,
					ClientSecret: b.resolver.getEnv("PROBOD_CONNECTOR_GITHUB_CLIENT_SECRET"),
				},
			},
		)
	}

	if sentryClientID := b.resolver.getEnv("PROBOD_CONNECTOR_SENTRY_CLIENT_ID"); sentryClientID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "SENTRY",
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:     sentryClientID,
					ClientSecret: b.resolver.getEnv("PROBOD_CONNECTOR_SENTRY_CLIENT_SECRET"),
				},
			},
		)
	}

	if intercomClientID := b.resolver.getEnv("PROBOD_CONNECTOR_INTERCOM_CLIENT_ID"); intercomClientID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "INTERCOM",
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:     intercomClientID,
					ClientSecret: b.resolver.getEnv("PROBOD_CONNECTOR_INTERCOM_CLIENT_SECRET"),
				},
			},
		)
	}

	if brexClientID := b.resolver.getEnv("PROBOD_CONNECTOR_BREX_CLIENT_ID"); brexClientID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "BREX",
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:     brexClientID,
					ClientSecret: b.resolver.getEnv("PROBOD_CONNECTOR_BREX_CLIENT_SECRET"),
				},
			},
		)
	}

	if googleWorkspaceClientID := b.resolver.getEnv("PROBOD_CONNECTOR_GOOGLE_WORKSPACE_CLIENT_ID"); googleWorkspaceClientID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "GOOGLE_WORKSPACE",
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:     googleWorkspaceClientID,
					ClientSecret: b.resolver.getEnv("PROBOD_CONNECTOR_GOOGLE_WORKSPACE_CLIENT_SECRET"),
				},
			},
		)
	}

	if microsoft365ClientID := b.resolver.getEnv("PROBOD_CONNECTOR_MICROSOFT_365_CLIENT_ID"); microsoft365ClientID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "MICROSOFT_365",
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:     microsoft365ClientID,
					ClientSecret: b.resolver.getEnv("PROBOD_CONNECTOR_MICROSOFT_365_CLIENT_SECRET"),
				},
			},
		)
	}

	for _, provider := range []string{
		"GITLAB",
		"BITBUCKET",
		"HEROKU",
		"PAGERDUTY",
		"ASANA",
		"NETLIFY",
		"CLICKUP",
		"MONDAY",
		"DATADOG",
		"ZENDESK",
		"LINEAR",
		"GOOGLE_ANALYTICS",
		"SQUARE",
		"FRONT",
	} {
		clientID := b.resolver.getEnv("PROBOD_CONNECTOR_" + provider + "_CLIENT_ID")
		if clientID == "" {
			continue
		}

		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: provider,
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:     clientID,
					ClientSecret: b.resolver.getEnv("PROBOD_CONNECTOR_" + provider + "_CLIENT_SECRET"),
				},
			},
		)
	}

	// Vercel needs the operator-supplied integration slug to resolve the
	// templated AuthURL ("https://vercel.com/integrations/{integration_slug}/new").
	if vercelClientID := b.resolver.getEnv("PROBOD_CONNECTOR_VERCEL_CLIENT_ID"); vercelClientID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "VERCEL",
				Protocol: "oauth2",
				RawConfig: probodconfig.ConnectorConfigOAuth2{
					ClientID:        vercelClientID,
					ClientSecret:    b.resolver.getEnv("PROBOD_CONNECTOR_VERCEL_CLIENT_SECRET"),
					IntegrationSlug: b.resolver.getEnv("PROBOD_CONNECTOR_VERCEL_INTEGRATION_SLUG"),
				},
			},
		)
	}

	// Crisp is a ManagedAPIKey (Model B) connector: Probo holds one
	// Marketplace plugin token (the verbatim "identifier:key" pair) shared
	// across all customer connections, and each connection carries only a
	// Website ID. The plugin ID is a separate value (the token's Basic
	// identifier is not the plugin ID) required by the per-website plugin API
	// that verifies website ownership at connect time. Both must be set to
	// activate the connector; until then it stays hidden from the driver
	// catalog, so it ships deactivated and activates the moment Crisp
	// validates the production plugin and both values are configured.
	crispPluginToken := b.resolver.getEnv("PROBOD_CONNECTOR_CRISP_PLUGIN_TOKEN")

	crispPluginID := b.resolver.getEnv("PROBOD_CONNECTOR_CRISP_PLUGIN_ID")
	if crispPluginToken != "" && crispPluginID != "" {
		cfg.Probod.Connectors = append(
			cfg.Probod.Connectors,
			probodconfig.ConnectorConfig{
				Provider: "CRISP",
				Protocol: "api_key",
				RawConfig: probodconfig.ConnectorConfigAPIKey{
					APIKey:     crispPluginToken,
					ResourceID: crispPluginID,
				},
			},
		)
	}

	connectorEndpoints, err := b.buildConnectorEndpoints()
	if err != nil {
		return nil, err
	}

	cfg.Probod.ConnectorEndpoints = connectorEndpoints

	if b.resolver.Err() != nil {
		return nil, b.resolver.Err()
	}

	if err := cfg.Probod.Auth.Cookie.Validate(); err != nil {
		return nil, fmt.Errorf("cannot validate auth cookie config: %w", err)
	}

	return cfg, nil
}

func (b *Builder) validateRequired() error {
	var missing []string

	required := []string{
		"PROBOD_ENCRYPTION_KEY",
		"PROBOD_AUTH_COOKIE_SECRET",
		"PROBOD_AUTH_PASSWORD_PEPPER",
	}

	for _, key := range required {
		if b.resolver.getEnv(key) == "" {
			missing = append(missing, key)
		}
	}

	if b.oauth2SigningKey == "" && b.resolver.getEnv("PROBOD_OAUTH2_SERVER_SIGNING_KEY") == "" {
		missing = append(missing, "PROBOD_OAUTH2_SERVER_SIGNING_KEY")
	}

	if slackClientID := b.resolver.getEnv("PROBOD_CONNECTOR_SLACK_CLIENT_ID"); slackClientID != "" {
		slackRequired := []string{
			"PROBOD_CONNECTOR_SLACK_CLIENT_SECRET",
			"PROBOD_CONNECTOR_SLACK_SIGNING_SECRET",
		}
		for _, key := range slackRequired {
			if b.resolver.getEnv(key) == "" {
				missing = append(missing, key+" (required when PROBOD_CONNECTOR_SLACK_CLIENT_ID is set)")
			}
		}
	}

	oauthProviders := []struct {
		envPrefix string
		required  []string
	}{
		{"CONNECTOR_HUBSPOT", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_DOCUSIGN", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_NOTION", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_GITHUB", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_SENTRY", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_INTERCOM", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_BREX", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_GOOGLE_WORKSPACE", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_MICROSOFT_365", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_GITLAB", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_BITBUCKET", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_HEROKU", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_PAGERDUTY", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_ASANA", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_NETLIFY", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_CLICKUP", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_MONDAY", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_DATADOG", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_ZENDESK", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_LINEAR", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_GOOGLE_ANALYTICS", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_SQUARE", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_FRONT", []string{"CLIENT_SECRET"}},
		{"CONNECTOR_VERCEL", []string{"CLIENT_SECRET", "INTEGRATION_SLUG"}},
	}

	for _, p := range oauthProviders {
		clientIDKey := "PROBOD_" + p.envPrefix + "_CLIENT_ID"
		if b.resolver.getEnv(clientIDKey) != "" {
			for _, suffix := range p.required {
				key := "PROBOD_" + p.envPrefix + "_" + suffix
				if b.resolver.getEnv(key) == "" {
					missing = append(missing, key+" (required when "+clientIDKey+" is set)")
				}
			}
		}
	}

	if len(missing) > 0 {
		if err := b.resolver.Err(); err != nil {
			return err
		}

		return fmt.Errorf("missing required environment variables:\n  - %s", strings.Join(missing, "\n  - "))
	}

	return nil
}

func (b *Builder) getSAMLCredentials() (cert, key string, err error) {
	cert = b.samlCertificate
	key = b.samlPrivateKey

	if cert == "" {
		cert = b.resolver.getEnv("PROBOD_SAML_CERTIFICATE")
	}

	if key == "" {
		key = b.resolver.getEnv("PROBOD_SAML_PRIVATE_KEY")
	}

	if cert == "" || key == "" {
		cert, key, err = GenerateSAMLCertificate()
		if err != nil {
			return "", "", fmt.Errorf("cannot generate SAML certificate: %w", err)
		}
	}

	return cert, key, nil
}

func (b *Builder) getOAuth2SigningKey() string {
	if b.oauth2SigningKey != "" {
		return b.oauth2SigningKey
	}

	return b.resolver.getEnv("PROBOD_OAUTH2_SERVER_SIGNING_KEY")
}

func (b *Builder) getPgCACertBundle() string {
	if path := b.resolver.getEnv("PROBOD_PG_CA_BUNDLE_PATH"); path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data)
		}
	}

	return b.resolver.getEnv("PROBOD_PG_CA_BUNDLE")
}

func (b *Builder) buildLLMProviders() map[string]probodconfig.LLMProviderConfig {
	providers := map[string]probodconfig.LLMProviderConfig{}

	if apiKey := b.resolver.getEnv("PROBOD_OPENAI_API_KEY"); apiKey != "" {
		providers["openai"] = probodconfig.LLMProviderConfig{
			Type:   "openai",
			APIKey: apiKey,
		}
	}

	if apiKey := b.resolver.getEnv("PROBOD_ANTHROPIC_API_KEY"); apiKey != "" {
		providers["anthropic"] = probodconfig.LLMProviderConfig{
			Type:   "anthropic",
			APIKey: apiKey,
		}
	}

	if len(providers) == 0 {
		return nil
	}

	return providers
}

func (b *Builder) parseOriginsList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var result []string

	for part := range strings.SplitSeq(s, ",") {
		part = strings.TrimSpace(part)

		part = strings.Trim(part, "\"")
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

// buildConnectorEndpoints collects per-provider endpoint overrides from
// PROBOD_CONNECTOR_<PROVIDER>_ENDPOINT_{AUTH,TOKEN,PROBE,IDENTITY,API_BASE},
// letting a deployment point a connector at a vendor sandbox without a code
// change.
//
// The provider list comes from the connector registry rather than a literal
// here, so a new provider is overridable the moment it is registered. The
// resolver looks up keys one at a time and cannot enumerate the environment,
// which is why the providers are iterated instead of the variables — that
// same blindness is why checkConnectorEndpointEnvTypos below exists as a
// separate, explicit environment scan.
//
// probod rejects an override that names an unknown provider or a field the
// provider does not resolve statically, so a typo fails at startup rather than
// sitting in the config doing nothing. Building the registry can itself fail
// on a bad Registration, which is why this returns an error instead of
// panicking: that failure is an operator-visible startup condition here, not
// the programmer error NewBuiltinRegistry treats it as elsewhere.
func (b *Builder) buildConnectorEndpoints() (map[string]probodconfig.ConnectorEndpointsConfig, error) {
	reg, err := provider.NewBuiltinRegistryWith()
	if err != nil {
		return nil, fmt.Errorf("cannot build connector provider registry: %w", err)
	}

	known := make(map[string]bool, len(reg.All()))
	for _, r := range reg.All() {
		known[string(r.Provider)] = true
	}

	if err := b.checkConnectorEndpointEnvTypos(known); err != nil {
		return nil, err
	}

	var endpoints map[string]probodconfig.ConnectorEndpointsConfig

	for _, r := range reg.All() {
		prefix := "PROBOD_CONNECTOR_" + string(r.Provider) + "_ENDPOINT_"

		e := probodconfig.ConnectorEndpointsConfig{
			Auth:     b.resolver.getEnv(prefix + "AUTH"),
			Token:    b.resolver.getEnv(prefix + "TOKEN"),
			Probe:    b.resolver.getEnv(prefix + "PROBE"),
			Identity: b.resolver.getEnv(prefix + "IDENTITY"),
			APIBase:  b.resolver.getEnv(prefix + "API_BASE"),
		}

		if e == (probodconfig.ConnectorEndpointsConfig{}) {
			continue
		}

		if endpoints == nil {
			endpoints = make(map[string]probodconfig.ConnectorEndpointsConfig)
		}

		endpoints[string(r.Provider)] = e
	}

	return endpoints, nil
}

// checkConnectorEndpointEnvTypos scans the environment (when the resolver has
// an enumerator — production always does, via NewResolver(nil)) for
// PROBOD_CONNECTOR_*_ENDPOINT_* keys and rejects any whose provider segment is
// not in known or whose field segment is not one of connectorEndpointFields.
//
// This exists because buildConnectorEndpoints's getEnv calls above are blind
// to a typo: PROBOD_CONNECTOR_GITHBU_ENDPOINT_API_BASE resolves to "" exactly
// like an unset key, so the misspelled provider's endpoint silently keeps
// pointing at production while the operator believes they repointed it.
// getEnv cannot see a key it never asked for; only a full-environment scan
// can. When the resolver has no enumerator (every test that does not opt in
// via Resolver.SetEnumerator) this is a no-op, matching the pre-existing
// getEnv-only behaviour.
func (b *Builder) checkConnectorEndpointEnvTypos(known map[string]bool) error {
	const (
		prefix = "PROBOD_CONNECTOR_"
		infix  = "_ENDPOINT_"
	)

	var bad []string

	for _, key := range b.resolver.keys() {
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		rest := strings.TrimPrefix(key, prefix)

		before, after, ok := strings.Cut(rest, infix)
		if !ok {
			continue
		}

		providerSeg := before
		fieldSeg := after

		switch {
		case !known[providerSeg]:
			bad = append(bad, fmt.Sprintf("%s: %q is not a registered connector provider", key, providerSeg))
		case !slices.Contains(connectorEndpointFields, fieldSeg):
			bad = append(bad, fmt.Sprintf("%s: %q is not a connector endpoint field", key, fieldSeg))
		}
	}

	if len(bad) > 0 {
		return fmt.Errorf("invalid connector endpoint override environment variable(s):\n  - %s", strings.Join(bad, "\n  - "))
	}

	return nil
}
