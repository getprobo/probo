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
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/probod"
)

type EnvGetter func(key string) string

type Builder struct {
	getEnv           EnvGetter
	samlCertificate  string
	samlPrivateKey   string
	oauth2SigningKey string
}

func NewBuilder(getEnv EnvGetter) *Builder {
	if getEnv == nil {
		getEnv = os.Getenv
	}
	return &Builder{getEnv: getEnv}
}

func (b *Builder) Build() (*probod.FullConfig, error) {
	if err := b.validateRequired(); err != nil {
		return nil, err
	}

	samlCert, samlKey, err := b.getSAMLCredentials()
	if err != nil {
		return nil, fmt.Errorf("cannot get SAML credentials: %w", err)
	}

	oauth2SigningKey := b.getOAuth2SigningKey()

	pgCACertBundle := b.getPgCACertBundle()

	cfg := &probod.FullConfig{
		Unit: probod.UnitConfig{
			Metrics: probod.MetricsConfig{
				Addr: b.getEnvOrDefault("METRICS_ADDR", "localhost:8081"),
			},
			Tracing: probod.TracingConfig{
				Addr:          b.getEnvOrDefault("TRACING_ADDR", "localhost:4317"),
				MaxBatchSize:  b.getEnvIntOrDefault("TRACING_MAX_BATCH_SIZE", 512),
				BatchTimeout:  b.getEnvIntOrDefault("TRACING_BATCH_TIMEOUT", 5),
				ExportTimeout: b.getEnvIntOrDefault("TRACING_EXPORT_TIMEOUT", 30),
				MaxQueueSize:  b.getEnvIntOrDefault("TRACING_MAX_QUEUE_SIZE", 2048),
			},
		},
		Probod: probod.Config{
			BaseURL:       b.getEnvOrDefault("PROBOD_BASE_URL", "http://localhost:8080"),
			EncryptionKey: b.getEnv("PROBOD_ENCRYPTION_KEY"),
			ChromeDPAddr:  b.getEnvOrDefault("CHROME_DP_ADDR", "localhost:9222"),
			Api: probod.APIConfig{
				Addr: b.getEnvOrDefault("API_ADDR", ":8080"),
				ProxyProtocol: probod.ProxyProtocolConfig{
					TrustedProxies: b.parseOriginsList(b.getEnv("API_PROXY_PROTOCOL_TRUSTED_PROXIES")),
				},
				Cors: probod.CorsConfig{
					AllowedOrigins: b.parseOriginsList(b.getEnvOrDefault("API_CORS_ALLOWED_ORIGINS", "http://localhost:8080")),
				},
				ExtraHeaderFields: make(map[string]string),
			},
			Pg: probod.PgConfig{
				Addr:                   b.getEnvOrDefault("PG_ADDR", "localhost:5432"),
				Username:               b.getEnvOrDefault("PG_USERNAME", "postgres"),
				Password:               b.getEnvOrDefault("PG_PASSWORD", "postgres"),
				Database:               b.getEnvOrDefault("PG_DATABASE", "probod"),
				PoolSize:               int32(b.getEnvIntOrDefault("PG_POOL_SIZE", 100)),
				MinPoolSize:            int32(b.getEnvIntOrDefault("PG_MIN_POOL_SIZE", 10)),
				MaxConnIdleTimeSeconds: b.getEnvIntOrDefault("PG_MAX_CONN_IDLE_TIME_SECONDS", 1800),
				MaxConnLifetimeSeconds: b.getEnvIntOrDefault("PG_MAX_CONN_LIFETIME_SECONDS", 3600),
				CACertBundle:           pgCACertBundle,
				Debug:                  b.getEnvBoolOrDefault("PG_DEBUG", false),
			},
			Auth: probod.AuthConfig{
				DisableSignup:                       b.getEnvBoolOrDefault("AUTH_DISABLE_SIGNUP", false),
				InvitationConfirmationTokenValidity: b.getEnvIntOrDefault("AUTH_INVITATION_TOKEN_VALIDITY", 3600),
				PasswordResetTokenValidity:          b.getEnvIntOrDefault("AUTH_PASSWORD_RESET_TOKEN_VALIDITY", 3600),
				MagicLinkTokenValidity:              b.getEnvIntOrDefault("AUTH_MAGIC_LINK_TOKEN_VALIDITY", 900),
				Cookie: probod.CookieConfig{
					Name:     b.getEnvOrDefault("AUTH_COOKIE_NAME", "SSID"),
					Domain:   b.getEnvOrDefault("AUTH_COOKIE_DOMAIN", "localhost"),
					Secret:   b.getEnv("AUTH_COOKIE_SECRET"),
					Duration: b.getEnvIntOrDefault("AUTH_COOKIE_DURATION", 24),
					Secure:   b.getEnvBoolOrDefault("AUTH_COOKIE_SECURE", true),
				},
				Password: probod.PasswordConfig{
					Pepper:     b.getEnv("AUTH_PASSWORD_PEPPER"),
					Iterations: b.getEnvIntOrDefault("AUTH_PASSWORD_ITERATIONS", 1000000),
				},
				SAML: probod.SAMLConfig{
					SessionDuration:                   b.getEnvIntOrDefault("SAML_SESSION_DURATION", 604800),
					CleanupIntervalSeconds:            b.getEnvIntOrDefault("SAML_CLEANUP_INTERVAL_SECONDS", 0),
					Certificate:                       samlCert,
					PrivateKey:                        samlKey,
					DomainVerificationIntervalSeconds: b.getEnvIntOrDefault("SAML_DOMAIN_VERIFICATION_INTERVAL_SECONDS", 60),
					DomainVerificationResolverAddr:    b.getEnvOrDefault("SAML_DOMAIN_VERIFICATION_RESOLVER_ADDR", "8.8.8.8:53"),
				},
				Google: probod.OIDCProviderConfig{
					ClientID:     b.getEnv("AUTH_GOOGLE_CLIENT_ID"),
					ClientSecret: b.getEnv("AUTH_GOOGLE_CLIENT_SECRET"),
					Enabled:      b.getEnv("AUTH_GOOGLE_CLIENT_ID") != "" && b.getEnv("AUTH_GOOGLE_CLIENT_SECRET") != "",
				},
				Microsoft: probod.OIDCProviderConfig{
					ClientID:     b.getEnv("AUTH_MICROSOFT_CLIENT_ID"),
					ClientSecret: b.getEnv("AUTH_MICROSOFT_CLIENT_SECRET"),
					Enabled:      b.getEnv("AUTH_MICROSOFT_CLIENT_ID") != "" && b.getEnv("AUTH_MICROSOFT_CLIENT_SECRET") != "",
				},
				OAuth2Server: probod.OAuth2ServerConfig{
					SigningKeys: []probod.OAuth2SigningKeyConfig{{
						PrivateKey: oauth2SigningKey,
						KID:        b.getEnvOrDefault("OAUTH2_SERVER_SIGNING_KEY_KID", "default"),
						Active:     true,
					}},
					AccessTokenDuration:       b.getEnvIntOrDefault("OAUTH2_SERVER_ACCESS_TOKEN_DURATION", 3600),
					RefreshTokenDuration:      b.getEnvIntOrDefault("OAUTH2_SERVER_REFRESH_TOKEN_DURATION", 2592000),
					AuthorizationCodeDuration: b.getEnvIntOrDefault("OAUTH2_SERVER_AUTHORIZATION_CODE_DURATION", 600),
					DeviceCodeDuration:        b.getEnvIntOrDefault("OAUTH2_SERVER_DEVICE_CODE_DURATION", 600),
				},
			},
			TrustCenter: probod.TrustCenterConfig{
				HTTPAddr:  b.getEnvOrDefault("TRUST_CENTER_HTTP_ADDR", ":80"),
				HTTPSAddr: b.getEnvOrDefault("TRUST_CENTER_HTTPS_ADDR", ":443"),
				ProxyProtocol: probod.ProxyProtocolConfig{
					TrustedProxies: b.parseOriginsList(b.getEnv("TRUST_CENTER_PROXY_PROTOCOL_TRUSTED_PROXIES")),
				},
			},
			AWS: probod.AWSConfig{
				Region:          b.getEnvOrDefault("AWS_REGION", "us-east-1"),
				Bucket:          b.getEnvOrDefault("AWS_BUCKET", "probod"),
				AccessKeyID:     b.getEnv("AWS_ACCESS_KEY_ID"),
				SecretAccessKey: b.getEnv("AWS_SECRET_ACCESS_KEY"),
				Endpoint:        b.getEnv("AWS_ENDPOINT"),
				UsePathStyle:    b.getEnvBoolOrDefault("AWS_USE_PATH_STYLE", false),
			},
			Notifications: probod.NotificationsConfig{
				Mailer: probod.MailerConfig{
					SenderName:     b.getEnvOrDefault("MAILER_SENDER_NAME", "Probo"),
					SenderEmail:    b.getEnvOrDefault("MAILER_SENDER_EMAIL", "no-reply@notification.getprobo.com"),
					MailerInterval: b.getEnvIntOrDefault("MAILER_INTERVAL", 60),
					SMTP: probod.SMTPConfig{
						Addr:        b.getEnvOrDefault("SMTP_ADDR", "localhost:1025"),
						User:        b.getEnv("SMTP_USER"),
						Password:    b.getEnv("SMTP_PASSWORD"),
						TLSRequired: b.getEnvBoolOrDefault("SMTP_TLS_REQUIRED", false),
					},
				},
				Slack: probod.SlackConfig{
					SenderInterval: b.getEnvIntOrDefault("SLACK_SENDER_INTERVAL", 60),
					SigningSecret:  b.getEnv("CONNECTOR_SLACK_SIGNING_SECRET"),
				},
				Webhook: probod.WebhookConfig{
					SenderInterval: b.getEnvIntOrDefault("WEBHOOK_SENDER_INTERVAL", 5),
					CacheTTL:       b.getEnvIntOrDefault("WEBHOOK_CACHE_TTL", 86400),
				},
			},
			Agents: probod.AgentsConfig{
				Providers: map[string]probod.LLMProviderConfig{
					"openai": {
						Type:   "openai",
						APIKey: b.getEnv("OPENAI_API_KEY"),
					},
					"anthropic": {
						Type:   "anthropic",
						APIKey: b.getEnv("ANTHROPIC_API_KEY"),
					},
				},
				Default: probod.LLMAgentConfig{
					Provider:    b.getEnvOrDefault("AGENT_DEFAULT_PROVIDER", "openai"),
					ModelName:   b.getEnvOrDefault("AGENT_DEFAULT_MODEL_NAME", "gpt-4o"),
					Temperature: new(b.getEnvFloatOrDefault("AGENT_DEFAULT_TEMPERATURE", 0.1)),
					MaxTokens:   new(b.getEnvIntOrDefault("AGENT_DEFAULT_MAX_TOKENS", 4096)),
				},
				Probo: probod.LLMAgentConfig{
					Provider:    b.getEnvOrDefault("AGENT_PROBO_PROVIDER", ""),
					ModelName:   b.getEnvOrDefault("AGENT_PROBO_MODEL_NAME", ""),
					Temperature: b.getEnvFloatPtr("AGENT_PROBO_TEMPERATURE"),
					MaxTokens:   b.getEnvIntPtr("AGENT_PROBO_MAX_TOKENS"),
				},
				EvidenceDescriber: probod.LLMAgentConfig{
					Provider:    b.getEnvOrDefault("AGENT_EVIDENCE_DESCRIBER_PROVIDER", ""),
					ModelName:   b.getEnvOrDefault("AGENT_EVIDENCE_DESCRIBER_MODEL_NAME", ""),
					Temperature: b.getEnvFloatPtr("AGENT_EVIDENCE_DESCRIBER_TEMPERATURE"),
					MaxTokens:   b.getEnvIntPtr("AGENT_EVIDENCE_DESCRIBER_MAX_TOKENS"),
				},
			},
			CustomDomains: probod.CustomDomainsConfig{
				RenewalInterval:   b.getEnvIntOrDefault("CUSTOM_DOMAINS_RENEWAL_INTERVAL", 3600),
				ProvisionInterval: b.getEnvIntOrDefault("CUSTOM_DOMAINS_PROVISION_INTERVAL", 30),
				CnameTarget:       b.getEnvOrDefault("CUSTOM_DOMAINS_CNAME_TARGET", "custom.getprobo.com"),
				ResolverAddr:      b.getEnvOrDefault("CUSTOM_DOMAINS_RESOLVER_ADDR", "8.8.8.8:53"),
				CAAIssuerDomain:   b.getEnvOrDefault("CUSTOM_DOMAINS_CAA_ISSUER_DOMAIN", "letsencrypt.org"),
				ACME: probod.ACMEConfig{
					Directory:  b.getEnvOrDefault("ACME_DIRECTORY", "https://acme-v02.api.letsencrypt.org/directory"),
					Email:      b.getEnvOrDefault("ACME_EMAIL", "admin@getprobo.com"),
					KeyType:    b.getEnvOrDefault("ACME_KEY_TYPE", "EC256"),
					RootCA:     b.getEnv("ACME_ROOT_CA"),
					AccountKey: b.getEnv("ACME_ACCOUNT_KEY"),
				},
			},
			SCIMBridge: probod.SCIMBridgeConfig{
				SyncInterval: b.getEnvIntOrDefault("SCIM_BRIDGE_SYNC_INTERVAL", 900),
				PollInterval: b.getEnvIntOrDefault("SCIM_BRIDGE_POLL_INTERVAL", 30),
			},
			ESign: probod.ESignConfig{
				TSAURL: b.getEnvOrDefault("ESIGN_TSA_URL", "http://timestamp.digicert.com"),
			},
			EvidenceDescriber: probod.EvidenceDescriberConfig{
				Interval:       b.getEnvIntOrDefault("EVIDENCE_DESCRIBER_INTERVAL", 10),
				StaleAfter:     b.getEnvIntOrDefault("EVIDENCE_DESCRIBER_STALE_AFTER", 300),
				MaxConcurrency: b.getEnvIntOrDefault("EVIDENCE_DESCRIBER_MAX_CONCURRENCY", 10),
			},
			Branding: b.getEnvBoolOrDefault("BRANDING", true),
		},
	}

	if slackClientID := b.getEnv("CONNECTOR_SLACK_CLIENT_ID"); slackClientID != "" {
		cfg.Probod.Connectors = append(cfg.Probod.Connectors, probod.ConnectorConfig{
			Provider: "SLACK",
			Protocol: "oauth2",
			RawConfig: probod.ConnectorConfigOAuth2{
				ClientID:     slackClientID,
				ClientSecret: b.getEnv("CONNECTOR_SLACK_CLIENT_SECRET"),
			},
			RawSettings: map[string]any{
				"signing-secret": b.getEnv("CONNECTOR_SLACK_SIGNING_SECRET"),
			},
		})
	}

	if hubspotClientID := b.getEnv("CONNECTOR_HUBSPOT_CLIENT_ID"); hubspotClientID != "" {
		cfg.Probod.Connectors = append(cfg.Probod.Connectors, probod.ConnectorConfig{
			Provider: "HUBSPOT",
			Protocol: "oauth2",
			RawConfig: probod.ConnectorConfigOAuth2{
				ClientID:     hubspotClientID,
				ClientSecret: b.getEnv("CONNECTOR_HUBSPOT_CLIENT_SECRET"),
			},
		})
	}

	if docusignClientID := b.getEnv("CONNECTOR_DOCUSIGN_CLIENT_ID"); docusignClientID != "" {
		cfg.Probod.Connectors = append(cfg.Probod.Connectors, probod.ConnectorConfig{
			Provider: "DOCUSIGN",
			Protocol: "oauth2",
			RawConfig: probod.ConnectorConfigOAuth2{
				ClientID:     docusignClientID,
				ClientSecret: b.getEnv("CONNECTOR_DOCUSIGN_CLIENT_SECRET"),
			},
		})
	}

	if notionClientID := b.getEnv("CONNECTOR_NOTION_CLIENT_ID"); notionClientID != "" {
		cfg.Probod.Connectors = append(cfg.Probod.Connectors, probod.ConnectorConfig{
			Provider: "NOTION",
			Protocol: "oauth2",
			RawConfig: probod.ConnectorConfigOAuth2{
				ClientID:     notionClientID,
				ClientSecret: b.getEnv("CONNECTOR_NOTION_CLIENT_SECRET"),
			},
		})
	}

	if githubClientID := b.getEnv("CONNECTOR_GITHUB_CLIENT_ID"); githubClientID != "" {
		cfg.Probod.Connectors = append(cfg.Probod.Connectors, probod.ConnectorConfig{
			Provider: "GITHUB",
			Protocol: "oauth2",
			RawConfig: probod.ConnectorConfigOAuth2{
				ClientID:     githubClientID,
				ClientSecret: b.getEnv("CONNECTOR_GITHUB_CLIENT_SECRET"),
			},
		})
	}

	if sentryClientID := b.getEnv("CONNECTOR_SENTRY_CLIENT_ID"); sentryClientID != "" {
		cfg.Probod.Connectors = append(cfg.Probod.Connectors, probod.ConnectorConfig{
			Provider: "SENTRY",
			Protocol: "oauth2",
			RawConfig: probod.ConnectorConfigOAuth2{
				ClientID:     sentryClientID,
				ClientSecret: b.getEnv("CONNECTOR_SENTRY_CLIENT_SECRET"),
			},
		})
	}

	if intercomClientID := b.getEnv("CONNECTOR_INTERCOM_CLIENT_ID"); intercomClientID != "" {
		cfg.Probod.Connectors = append(cfg.Probod.Connectors, probod.ConnectorConfig{
			Provider: "INTERCOM",
			Protocol: "oauth2",
			RawConfig: probod.ConnectorConfigOAuth2{
				ClientID:     intercomClientID,
				ClientSecret: b.getEnv("CONNECTOR_INTERCOM_CLIENT_SECRET"),
			},
		})
	}

	if brexClientID := b.getEnv("CONNECTOR_BREX_CLIENT_ID"); brexClientID != "" {
		cfg.Probod.Connectors = append(cfg.Probod.Connectors, probod.ConnectorConfig{
			Provider: "BREX",
			Protocol: "oauth2",
			RawConfig: probod.ConnectorConfigOAuth2{
				ClientID:     brexClientID,
				ClientSecret: b.getEnv("CONNECTOR_BREX_CLIENT_SECRET"),
			},
		})
	}

	return cfg, nil
}

func (b *Builder) validateRequired() error {
	var missing []string

	required := []string{
		"PROBOD_ENCRYPTION_KEY",
		"AUTH_COOKIE_SECRET",
		"AUTH_PASSWORD_PEPPER",
	}

	for _, key := range required {
		if b.getEnv(key) == "" {
			missing = append(missing, key)
		}
	}

	if b.oauth2SigningKey == "" && b.getEnv("OAUTH2_SERVER_SIGNING_KEY") == "" {
		missing = append(missing, "OAUTH2_SERVER_SIGNING_KEY")
	}

	if slackClientID := b.getEnv("CONNECTOR_SLACK_CLIENT_ID"); slackClientID != "" {
		slackRequired := []string{
			"CONNECTOR_SLACK_CLIENT_SECRET",
			"CONNECTOR_SLACK_SIGNING_SECRET",
		}
		for _, key := range slackRequired {
			if b.getEnv(key) == "" {
				missing = append(missing, key+" (required when CONNECTOR_SLACK_CLIENT_ID is set)")
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
	}

	for _, p := range oauthProviders {
		clientIDKey := p.envPrefix + "_CLIENT_ID"
		if b.getEnv(clientIDKey) != "" {
			for _, suffix := range p.required {
				key := p.envPrefix + "_" + suffix
				if b.getEnv(key) == "" {
					missing = append(missing, key+" (required when "+clientIDKey+" is set)")
				}
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables:\n  - %s", strings.Join(missing, "\n  - "))
	}

	return nil
}

func (b *Builder) getSAMLCredentials() (cert, key string, err error) {
	cert = b.samlCertificate
	key = b.samlPrivateKey

	if cert == "" {
		cert = b.getEnv("SAML_CERTIFICATE")
	}
	if key == "" {
		key = b.getEnv("SAML_PRIVATE_KEY")
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
	return b.getEnv("OAUTH2_SERVER_SIGNING_KEY")
}

func (b *Builder) getPgCACertBundle() string {
	if path := b.getEnv("PG_CA_BUNDLE_PATH"); path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data)
		}
	}

	return b.getEnv("PG_CA_BUNDLE")
}

func (b *Builder) getEnvOrDefault(key, defaultValue string) string {
	if value := b.getEnv(key); value != "" {
		return value
	}
	return defaultValue
}

func (b *Builder) getEnvIntOrDefault(key string, defaultValue int) int {
	if value := b.getEnv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			return int(intValue)
		}
	}
	return defaultValue
}

func (b *Builder) getEnvFloatOrDefault(key string, defaultValue float64) float64 {
	if value := b.getEnv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func (b *Builder) getEnvFloatPtr(key string) *float64 {
	if value := b.getEnv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return &floatValue
		}
	}
	return nil
}

func (b *Builder) getEnvIntPtr(key string) *int {
	if value := b.getEnv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			v := int(intValue)
			return &v
		}
	}
	return nil
}

func (b *Builder) getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := b.getEnv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
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
