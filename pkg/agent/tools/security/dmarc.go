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
	dmarcParams struct {
		Domain string `json:"domain" jsonschema:"The domain to check DMARC record for (e.g. example.com)"`
	}

	dmarcResult struct {
		Found       bool   `json:"found"`
		RawRecord   string `json:"raw_record,omitempty"`
		Policy      string `json:"policy,omitempty"`
		Percentage  string `json:"pct,omitempty"`
		RUA         string `json:"rua,omitempty"`
		RUF         string `json:"ruf,omitempty"`
		ErrorDetail string `json:"error_detail,omitempty"`
	}
)

func parseDMARCTag(record, tag string) string {
	for part := range strings.SplitSeq(record, ";") {
		part = strings.TrimSpace(part)
		if after, ok := strings.CutPrefix(part, tag+"="); ok {
			return after
		}
	}
	return ""
}

func CheckDMARCTool() agent.Tool {
	return agent.FunctionTool(
		"check_dmarc",
		"Check the DMARC DNS record for a domain, returning the policy, percentage, and reporting addresses.",
		func(ctx context.Context, p dmarcParams) (agent.ToolResult, error) {
			fqdn := "_dmarc." + p.Domain
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
				return agent.ResultJSON(dmarcResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("cannot lookup DMARC record: %s", err),
				}), nil
			}

			for _, answer := range answers {
				txt, ok := answer.(*dns.TXT)
				if !ok {
					continue
				}

				record := strings.Join(txt.Txt, "")
				if !strings.HasPrefix(record, "v=DMARC1") {
					continue
				}

				result := dmarcResult{
					Found:      true,
					RawRecord:  record,
					Policy:     parseDMARCTag(record, "p"),
					Percentage: parseDMARCTag(record, "pct"),
					RUA:        parseDMARCTag(record, "rua"),
					RUF:        parseDMARCTag(record, "ruf"),
				}

				return agent.ResultJSON(result), nil
			}

			return agent.ResultJSON(dmarcResult{Found: false}), nil
		},
	)
}
