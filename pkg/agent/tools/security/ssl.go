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

package security

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"go.probo.inc/probo/pkg/agent"
)

type sslParams struct {
	Domain string `json:"domain" jsonschema:"description=The domain to check the SSL certificate for (e.g. example.com)"`
}

type sslResult struct {
	Valid       bool     `json:"valid"`
	Issuer      string   `json:"issuer"`
	Subject     string   `json:"subject"`
	NotBefore   string   `json:"not_before"`
	NotAfter    string   `json:"not_after"`
	DaysLeft    int      `json:"days_left"`
	Protocol    string   `json:"protocol"`
	DNSNames    []string `json:"dns_names"`
	IsExpired   bool     `json:"is_expired"`
	ErrorDetail string   `json:"error_detail,omitempty"`
}

func protocolName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("unknown (0x%04x)", version)
	}
}

func CheckSSLCertificateTool() (agent.Tool, error) {
	return agent.FunctionTool[sslParams](
		"check_ssl_certificate",
		"Check the SSL/TLS certificate for a domain, returning issuer, expiry, protocol version, and validity.",
		func(ctx context.Context, p sslParams) (agent.ToolResult, error) {
			conn, err := tls.DialWithDialer(
				&net.Dialer{Timeout: 10 * time.Second},
				"tcp",
				p.Domain+":443",
				&tls.Config{
					InsecureSkipVerify: false,
				},
			)
			if err != nil {
				data, _ := json.Marshal(sslResult{
					Valid:       false,
					ErrorDetail: err.Error(),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}
			defer conn.Close()

			state := conn.ConnectionState()
			if len(state.PeerCertificates) == 0 {
				data, _ := json.Marshal(sslResult{
					Valid:       false,
					ErrorDetail: "no peer certificates",
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			cert := state.PeerCertificates[0]
			now := time.Now()

			result := sslResult{
				Valid:     now.Before(cert.NotAfter) && now.After(cert.NotBefore),
				Issuer:    cert.Issuer.String(),
				Subject:   cert.Subject.String(),
				NotBefore: cert.NotBefore.Format(time.RFC3339),
				NotAfter:  cert.NotAfter.Format(time.RFC3339),
				DaysLeft:  int(time.Until(cert.NotAfter).Hours() / 24),
				Protocol:  protocolName(state.Version),
				DNSNames:  cert.DNSNames,
				IsExpired: now.After(cert.NotAfter),
			}

			data, _ := json.Marshal(result)

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
