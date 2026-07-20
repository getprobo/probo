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
		Addr          string `json:"addr,omitempty"`
		MaxBatchSize  int    `json:"max-batch-size"`
		BatchTimeout  int    `json:"batch-timeout"`
		ExportTimeout int    `json:"export-timeout"`
		MaxQueueSize  int    `json:"max-queue-size"`
	}

	// ESignConfig contains electronic signature configuration.
	ESignConfig struct {
		TSAURL string `json:"tsa-url,omitempty"`
	}

	// Config represents the probod application configuration.
	Config struct {
		BaseURL           string                        `json:"base-url,omitempty"`
		EncryptionKey     string                        `json:"encryption-key"`
		Pg                PgConfig                      `json:"pg"`
		Api               APIConfig                     `json:"api"`
		Auth              AuthConfig                    `json:"auth"`
		CompliancePortal  CompliancePortalConfig        `json:"trust-center"`
		AWS               AWSConfig                     `json:"aws"`
		Notifications     NotificationsConfig           `json:"notifications"`
		Connectors        []ConnectorConfig             `json:"connectors,omitempty"`
		Agents            AgentsConfig                  `json:"llm"`
		EvidenceDescriber EvidenceDescriberConfig       `json:"evidence-describer"`
		ThirdPartyVetting ThirdPartyVettingWorkerConfig `json:"third-party-vetting-worker"`

		TrackerMappingWorker             TrackerMappingWorkerConfig             `json:"tracker-mapping-worker"`
		CommonPatternEnrichmentWorker    CommonPatternEnrichmentWorkerConfig    `json:"common-pattern-enrichment-worker"`
		CommonThirdPartyEnrichmentWorker CommonThirdPartyEnrichmentWorkerConfig `json:"common-third-party-enrichment-worker"`

		ChromeDPAddr  string              `json:"chrome-dp-addr,omitempty"`
		CustomDomains CustomDomainsConfig `json:"custom-domains"`
		SCIMBridge    SCIMBridgeConfig    `json:"scim-bridge"`
		ESign         ESignConfig         `json:"esign,omitzero"`
		Branding      bool                `json:"branding"`
	}

	// CompliancePortalConfig contains compliance portal server configuration.
	CompliancePortalConfig struct {
		HTTPAddr      string              `json:"http-addr,omitempty"`
		HTTPSAddr     string              `json:"https-addr,omitempty"`
		BaseDomain    string              `json:"base-domain,omitempty"`
		ProxyProtocol ProxyProtocolConfig `json:"proxy-protocol,omitzero"`
	}
)
