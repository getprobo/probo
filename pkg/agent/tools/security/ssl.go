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
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/internal/netcheck"
)

type (
	sslParams struct {
		Domain string `json:"domain" jsonschema:"The domain to check the SSL certificate for (e.g. example.com)"`
	}

	sslResult struct {
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
)

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

func CheckSSLCertificateTool() agent.Tool {
	return agent.FunctionTool(
		"check_ssl_certificate",
		"Check the SSL/TLS certificate for a domain, returning issuer, expiry, protocol version, and validity.",
		func(ctx context.Context, p sslParams) (agent.ToolResult, error) {
			if err := netcheck.ValidatePublicDomain(p.Domain); err != nil {
				return agent.ResultJSON(sslResult{
					Valid:       false,
					ErrorDetail: fmt.Sprintf("domain not allowed: %s", err),
				}), nil
			}

			// This is a certificate inspection tool: we intentionally
			// connect to servers whose certificates may be expired,
			// self-signed, or otherwise invalid, because the whole
			// point is to report back on the certificate state.
			// InsecureSkipVerify disables the handshake's built-in
			// verification; we then perform the verification manually
			// below (x509.Verify) and surface the result in Valid.
			// This pattern is safe here because we never send any
			// credentials or confidential data over the connection.
			dialer := &tls.Dialer{
				NetDialer: &net.Dialer{Timeout: 10 * time.Second},
				Config: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec // cert inspector; verification happens manually below
					ServerName:         p.Domain,
				},
			}
			netConn, err := dialer.DialContext(ctx, "tcp", p.Domain+":443")
			var conn *tls.Conn
			if netConn != nil {
				conn = netConn.(*tls.Conn)
			}
			if err != nil {
				return agent.ResultJSON(sslResult{
					Valid:       false,
					ErrorDetail: err.Error(),
				}), nil
			}
			defer func() { _ = conn.Close() }()

			state := conn.ConnectionState()
			if len(state.PeerCertificates) == 0 {
				return agent.ResultJSON(sslResult{
					Valid:       false,
					ErrorDetail: "no peer certificates",
				}), nil
			}

			cert := state.PeerCertificates[0]
			now := time.Now()

			// Manually verify the certificate since we connected
			// with InsecureSkipVerify to retrieve cert details
			// even for expired/invalid certificates.
			valid := now.Before(cert.NotAfter) && now.After(cert.NotBefore)
			if valid {
				opts := x509.VerifyOptions{
					DNSName:       p.Domain,
					Intermediates: x509.NewCertPool(),
				}
				for _, ic := range state.PeerCertificates[1:] {
					opts.Intermediates.AddCert(ic)
				}
				if _, err := cert.Verify(opts); err != nil {
					valid = false
				}
			}

			result := sslResult{
				Valid:     valid,
				Issuer:    cert.Issuer.String(),
				Subject:   cert.Subject.String(),
				NotBefore: cert.NotBefore.Format(time.RFC3339),
				NotAfter:  cert.NotAfter.Format(time.RFC3339),
				DaysLeft:  int(time.Until(cert.NotAfter).Hours() / 24),
				Protocol:  protocolName(state.Version),
				DNSNames:  cert.DNSNames,
				IsExpired: now.After(cert.NotAfter),
			}

			return agent.ResultJSON(result), nil
		},
	)
}
