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
				return agent.ResultJSON(
					dmarcResult{
						Found:       false,
						ErrorDetail: fmt.Sprintf("cannot lookup DMARC record: %s", err),
					},
				), nil
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
