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

package probod

import "go.probo.inc/probo/pkg/probodconfig"

type (
	FullConfig              = probodconfig.FullConfig
	Config                  = probodconfig.Config
	UnitConfig              = probodconfig.UnitConfig
	MetricsConfig           = probodconfig.MetricsConfig
	TracingConfig           = probodconfig.TracingConfig
	ESignConfig             = probodconfig.ESignConfig
	TrustCenterConfig       = probodconfig.TrustCenterConfig
	APIConfig               = probodconfig.APIConfig
	CorsConfig              = probodconfig.CorsConfig
	ProxyProtocolConfig     = probodconfig.ProxyProtocolConfig
	AuthConfig              = probodconfig.AuthConfig
	OAuth2ServerConfig      = probodconfig.OAuth2ServerConfig
	OAuth2SigningKeyConfig  = probodconfig.OAuth2SigningKeyConfig
	CookieConfig            = probodconfig.CookieConfig
	PasswordConfig          = probodconfig.PasswordConfig
	AWSConfig               = probodconfig.AWSConfig
	ConnectorConfig         = probodconfig.ConnectorConfig
	ConnectorConfigOAuth2   = probodconfig.ConnectorConfigOAuth2
	CustomDomainsConfig     = probodconfig.CustomDomainsConfig
	ACMEConfig              = probodconfig.ACMEConfig
	LLMProviderConfig       = probodconfig.LLMProviderConfig
	LLMAgentConfig          = probodconfig.LLMAgentConfig
	EvidenceDescriberConfig = probodconfig.EvidenceDescriberConfig
	AgentsConfig            = probodconfig.AgentsConfig
	MailerConfig            = probodconfig.MailerConfig
	SMTPConfig              = probodconfig.SMTPConfig
	NotificationsConfig     = probodconfig.NotificationsConfig
	WebhookConfig           = probodconfig.WebhookConfig
	OIDCProviderConfig      = probodconfig.OIDCProviderConfig
	PgConfig                = probodconfig.PgConfig
	SAMLConfig              = probodconfig.SAMLConfig
	SCIMBridgeConfig        = probodconfig.SCIMBridgeConfig
	SlackConfig             = probodconfig.SlackConfig
)
