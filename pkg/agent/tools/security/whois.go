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
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/internal/netcheck"
)

type (
	whoisParams struct {
		Domain string `json:"domain" jsonschema:"The domain to perform a WHOIS lookup on (e.g. example.com)"`
	}

	whoisResult struct {
		Registrar     string   `json:"registrar,omitempty"`
		CreationDate  string   `json:"creation_date,omitempty"`
		ExpiryDate    string   `json:"expiry_date,omitempty"`
		UpdatedDate   string   `json:"updated_date,omitempty"`
		RegistrantOrg string   `json:"registrant_org,omitempty"`
		RegistrantCC  string   `json:"registrant_country,omitempty"`
		NameServers   []string `json:"name_servers,omitempty"`
		DomainAge     string   `json:"domain_age,omitempty"`
		ErrorDetail   string   `json:"error_detail,omitempty"`
	}
)

func CheckWhoisTool() (agent.Tool, error) {
	return agent.FunctionTool(
		"check_whois",
		"Perform a WHOIS lookup on a domain to retrieve registration details including registrar, creation date, expiry date, registrant organization, and name servers.",
		func(ctx context.Context, p whoisParams) (agent.ToolResult, error) {
			if err := netcheck.ValidatePublicDomain(p.Domain); err != nil {
				return agent.ResultJSON(whoisResult{
					ErrorDetail: fmt.Sprintf("domain not allowed: %s", err),
				}), nil
			}

			// Step 1: query IANA to find the referral WHOIS server.
			referral, err := queryWhois(ctx, "whois.iana.org:43", p.Domain)
			if err != nil {
				return agent.ResultJSON(whoisResult{
					ErrorDetail: fmt.Sprintf("cannot query IANA WHOIS: %s", err),
				}), nil
			}

			whoisServer := parseWhoisField(referral, "refer")
			if whoisServer == "" {
				whoisServer = parseWhoisField(referral, "whois")
			}
			if whoisServer == "" {
				// Try common TLD WHOIS servers as fallback.
				parts := strings.Split(p.Domain, ".")
				tld := parts[len(parts)-1]
				whoisServer = "whois." + tld + ".com"
			}

			if !strings.Contains(whoisServer, ":") {
				whoisServer = whoisServer + ":43"
			}

			// Validate the referral WHOIS server resolves to a public IP
			// to prevent SSRF via crafted IANA responses.
			whoisHost, _, _ := net.SplitHostPort(whoisServer)
			if whoisHost == "" {
				whoisHost = whoisServer
			}
			if err := netcheck.ValidatePublicDomain(whoisHost); err != nil {
				return agent.ResultJSON(whoisResult{
					ErrorDetail: fmt.Sprintf("WHOIS referral server not allowed: %s", err),
				}), nil
			}

			// Step 2: query the registrar's WHOIS server.
			raw, err := queryWhois(ctx, whoisServer, p.Domain)
			if err != nil {
				return agent.ResultJSON(whoisResult{
					ErrorDetail: fmt.Sprintf("cannot query WHOIS server %s: %s", whoisServer, err),
				}), nil
			}

			result := parseWhoisResponse(raw)

			// Compute domain age from creation date.
			if result.CreationDate != "" {
				for _, layout := range []string{
					"2006-01-02T15:04:05Z",
					"2006-01-02",
					"02-Jan-2006",
					"2006-01-02 15:04:05",
					time.RFC3339,
				} {
					if t, err := time.Parse(layout, result.CreationDate); err == nil {
						age := time.Since(t)
						years := int(age.Hours() / 24 / 365)
						months := int(age.Hours()/24/30) % 12
						result.DomainAge = fmt.Sprintf("%d years, %d months", years, months)
						break
					}
				}
			}

			return agent.ResultJSON(result), nil
		},
	)
}

func queryWhois(ctx context.Context, server, domain string) (string, error) {
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", server)
	if err != nil {
		return "", fmt.Errorf("cannot connect to %s: %w", server, err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	_, err = fmt.Fprintf(conn, "%s\r\n", domain)
	if err != nil {
		return "", fmt.Errorf("cannot write to %s: %w", server, err)
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("cannot read from %s: %w", server, err)
	}

	return sb.String(), nil
}

func parseWhoisField(raw, field string) string {
	field = strings.ToLower(field)
	for line := range strings.SplitSeq(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.ToLower(strings.TrimSpace(k)) == field {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

var (
	whoisFieldMap = map[string]string{
		"registrar":                              "registrar",
		"registrar name":                         "registrar",
		"sponsoring registrar":                   "registrar",
		"creation date":                          "creation_date",
		"created":                                "creation_date",
		"created on":                             "creation_date",
		"registration date":                      "creation_date",
		"domain name commencement date":          "creation_date",
		"registry expiry date":                   "expiry_date",
		"registrar registration expiration date": "expiry_date",
		"expiry date":                            "expiry_date",
		"paid-till":                              "expiry_date",
		"updated date":                           "updated_date",
		"last updated":                           "updated_date",
		"last modified":                          "updated_date",
		"registrant organization":                "registrant_org",
		"registrant organisation":                "registrant_org",
		"org":                                    "registrant_org",
		"registrant country":                     "registrant_cc",
		"registrant country/economy":             "registrant_cc",
		"name server":                            "name_server",
		"nserver":                                "name_server",
	}
)

func parseWhoisResponse(raw string) whoisResult {
	var result whoisResult
	for line := range strings.SplitSeq(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(k))
		val := strings.TrimSpace(v)
		if val == "" {
			continue
		}

		field, ok := whoisFieldMap[key]
		if !ok {
			continue
		}

		switch field {
		case "registrar":
			if result.Registrar == "" {
				result.Registrar = val
			}
		case "creation_date":
			if result.CreationDate == "" {
				result.CreationDate = val
			}
		case "expiry_date":
			if result.ExpiryDate == "" {
				result.ExpiryDate = val
			}
		case "updated_date":
			if result.UpdatedDate == "" {
				result.UpdatedDate = val
			}
		case "registrant_org":
			if result.RegistrantOrg == "" {
				result.RegistrantOrg = val
			}
		case "registrant_cc":
			if result.RegistrantCC == "" {
				result.RegistrantCC = val
			}
		case "name_server":
			result.NameServers = append(result.NameServers, strings.ToLower(val))
		}
	}

	return result
}
