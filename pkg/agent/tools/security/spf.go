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
	"fmt"
	"strings"

	"codeberg.org/miekg/dns"
	"go.probo.inc/probo/pkg/agent"
)

type (
	spfParams struct {
		Domain string `json:"domain" jsonschema:"The domain to check SPF record for (e.g. example.com)"`
	}

	spfResult struct {
		Found       bool   `json:"found"`
		RawRecord   string `json:"raw_record,omitempty"`
		Policy      string `json:"policy,omitempty"`
		Mechanisms  string `json:"mechanisms,omitempty"`
		ErrorDetail string `json:"error_detail,omitempty"`
	}
)

func parseSPFPolicy(record string) string {
	for part := range strings.FieldsSeq(strings.ToLower(record)) {
		switch part {
		case "-all":
			return "fail"
		case "~all":
			return "softfail"
		case "?all":
			return "neutral"
		case "+all":
			return "pass"
		}
	}

	return ""
}

func CheckSPFTool() (agent.Tool, error) {
	return agent.FunctionTool(
		"check_spf",
		"Check the SPF (Sender Policy Framework) DNS record for a domain, returning the raw record and its policy qualifier.",
		func(ctx context.Context, p spfParams) (agent.ToolResult, error) {
			fqdn := p.Domain
			if !strings.HasSuffix(fqdn, ".") {
				fqdn = fqdn + "."
			}

			client := dns.NewClient()
			answers, err := queryDNS(
				ctx,
				client,
				&dns.TXT{
					Hdr: dns.Header{
						Name:  fqdn,
						Class: dns.ClassINET,
					},
				},
			)
			if err != nil {
				return agent.ResultJSON(spfResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("cannot lookup SPF record: %s", err),
				}), nil
			}

			var spfRecords []string
			for _, answer := range answers {
				txt, ok := answer.(*dns.TXT)
				if !ok {
					continue
				}

				record := strings.Join(txt.Txt, "")
				if !strings.HasPrefix(strings.ToLower(record), "v=spf1") {
					continue
				}

				spfRecords = append(spfRecords, record)
			}

			if len(spfRecords) > 1 {
				return agent.ResultJSON(spfResult{
					Found:       true,
					ErrorDetail: fmt.Sprintf("multiple SPF records found (%d); this is an invalid configuration per RFC 7208", len(spfRecords)),
				}), nil
			}

			if len(spfRecords) == 1 {
				record := spfRecords[0]
				return agent.ResultJSON(spfResult{
					Found:      true,
					RawRecord:  record,
					Policy:     parseSPFPolicy(record),
					Mechanisms: record,
				}), nil
			}

			return agent.ResultJSON(spfResult{Found: false}), nil
		},
	)
}
