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

package probodconfig

type (
	// FullConfig represents the complete configuration file structure.
	// This is used by bootstrap to generate the YAML config file.
	FullConfig struct {
		Unit   UnitConfig `json:"unit"`
		Probod Config     `json:"probod"`
	}

	// UnitConfig contains unit framework configuration.
	UnitConfig struct {
		Metrics MetricsConfig `json:"metrics"`
		Tracing TracingConfig `json:"tracing"`
	}

	// MetricsConfig contains metrics server configuration.
	MetricsConfig struct {
		Addr string `json:"addr"`
	}

	// TracingConfig contains tracing configuration.
	TracingConfig struct {
		Addr          string  `json:"addr"`
		MaxBatchSize  int     `json:"max-batch-size"`
		BatchTimeout  int     `json:"batch-timeout"`
		ExportTimeout int     `json:"export-timeout"`
		MaxQueueSize  int     `json:"max-queue-size"`
		SamplingRatio float64 `json:"sampling-ratio"`
	}

	// ESignConfig contains electronic signature configuration.
	ESignConfig struct {
		TSAURL string `json:"tsa-url"`
	}

	// Config represents the probod application configuration.
	Config struct {
		BaseURL           string                        `json:"base-url"`
		EncryptionKey     string                        `json:"encryption-key"`
		Pg                PgConfig                      `json:"pg"`
		Api               APIConfig                     `json:"api"`
		Auth              AuthConfig                    `json:"auth"`
		TrustCenter       TrustCenterConfig             `json:"trust-center"`
		AWS               AWSConfig                     `json:"aws"`
		Notifications     NotificationsConfig           `json:"notifications"`
		Connectors        []ConnectorConfig             `json:"connectors"`
		Agents            AgentsConfig                  `json:"llm"`
		EvidenceDescriber EvidenceDescriberConfig       `json:"evidence-describer"`
		ThirdPartyVetting ThirdPartyVettingWorkerConfig `json:"third-party-vetting-worker"`

		TrackerMappingWorker          TrackerMappingWorkerConfig          `json:"tracker-mapping-worker"`
		CommonPatternEnrichmentWorker CommonPatternEnrichmentWorkerConfig `json:"common-pattern-enrichment-worker"`

		ChromeDPAddr  string              `json:"chrome-dp-addr"`
		CustomDomains CustomDomainsConfig `json:"custom-domains"`
		SCIMBridge    SCIMBridgeConfig    `json:"scim-bridge"`
		ESign         ESignConfig         `json:"esign"`
		Branding      bool                `json:"branding"`
	}

	// TrustCenterConfig contains trust center server configuration.
	TrustCenterConfig struct {
		HTTPAddr      string              `json:"http-addr"`
		HTTPSAddr     string              `json:"https-addr"`
		ProxyProtocol ProxyProtocolConfig `json:"proxy-protocol"`
	}
)
