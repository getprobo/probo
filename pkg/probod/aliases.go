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

package probod

import "go.probo.inc/probo/pkg/probodconfig"

type (
	FullConfig                    = probodconfig.FullConfig
	Config                        = probodconfig.Config
	UnitConfig                    = probodconfig.UnitConfig
	MetricsConfig                 = probodconfig.MetricsConfig
	TracingConfig                 = probodconfig.TracingConfig
	ESignConfig                   = probodconfig.ESignConfig
	TrustCenterConfig             = probodconfig.TrustCenterConfig
	APIConfig                     = probodconfig.APIConfig
	CorsConfig                    = probodconfig.CorsConfig
	GraphQLConfig                 = probodconfig.GraphQLConfig
	ProxyProtocolConfig           = probodconfig.ProxyProtocolConfig
	AuthConfig                    = probodconfig.AuthConfig
	OAuth2ServerConfig            = probodconfig.OAuth2ServerConfig
	OAuth2SigningKeyConfig        = probodconfig.OAuth2SigningKeyConfig
	CookieConfig                  = probodconfig.CookieConfig
	PasswordConfig                = probodconfig.PasswordConfig
	AWSConfig                     = probodconfig.AWSConfig
	ConnectorConfig               = probodconfig.ConnectorConfig
	ConnectorConfigOAuth2         = probodconfig.ConnectorConfigOAuth2
	CustomDomainsConfig           = probodconfig.CustomDomainsConfig
	ACMEConfig                    = probodconfig.ACMEConfig
	LLMProviderConfig             = probodconfig.LLMProviderConfig
	LLMAgentConfig                = probodconfig.LLMAgentConfig
	EvidenceDescriberConfig       = probodconfig.EvidenceDescriberConfig
	ThirdPartyVettingWorkerConfig = probodconfig.ThirdPartyVettingWorkerConfig
	AgentsConfig                  = probodconfig.AgentsConfig

	TrackerMappingWorkerConfig             = probodconfig.TrackerMappingWorkerConfig
	CommonPatternEnrichmentWorkerConfig    = probodconfig.CommonPatternEnrichmentWorkerConfig
	CommonThirdPartyEnrichmentWorkerConfig = probodconfig.CommonThirdPartyEnrichmentWorkerConfig

	MailerConfig        = probodconfig.MailerConfig
	SMTPConfig          = probodconfig.SMTPConfig
	NotificationsConfig = probodconfig.NotificationsConfig
	WebhookConfig       = probodconfig.WebhookConfig

	DocumentNotificationConfig = probodconfig.DocumentNotificationConfig
	OIDCProviderConfig         = probodconfig.OIDCProviderConfig
	PgConfig                   = probodconfig.PgConfig
	SAMLConfig                 = probodconfig.SAMLConfig
	SCIMBridgeConfig           = probodconfig.SCIMBridgeConfig
	SlackConfig                = probodconfig.SlackConfig
)
