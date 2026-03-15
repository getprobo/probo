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
	"encoding/json"
	"fmt"
	"strings"

	"codeberg.org/miekg/dns"
	"go.probo.inc/probo/pkg/agent"
)

type spfParams struct {
	Domain string `json:"domain" jsonschema:"description=The domain to check SPF record for (e.g. example.com)"`
}

type spfResult struct {
	Found       bool   `json:"found"`
	RawRecord   string `json:"raw_record,omitempty"`
	Policy      string `json:"policy,omitempty"`
	Mechanisms  string `json:"mechanisms,omitempty"`
	ErrorDetail string `json:"error_detail,omitempty"`
}

func parseSPFPolicy(record string) string {
	for _, part := range strings.Fields(strings.ToLower(record)) {
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
	return agent.FunctionTool[spfParams](
		"check_spf",
		"Check the SPF (Sender Policy Framework) DNS record for a domain, returning the raw record and its policy qualifier.",
		func(ctx context.Context, p spfParams) (agent.ToolResult, error) {
			fqdn := p.Domain
			if !strings.HasSuffix(fqdn, ".") {
				fqdn = fqdn + "."
			}

			msg := &dns.Msg{
				MsgHeader: dns.MsgHeader{
					ID:               dns.ID(),
					RecursionDesired: true,
				},
			}
			msg.Question = []dns.RR{
				&dns.TXT{
					Hdr: dns.Header{
						Name:  fqdn,
						Class: dns.ClassINET,
					},
				},
			}

			client := dns.NewClient()
			resp, _, err := client.Exchange(ctx, msg, "udp", defaultResolverAddr)
			if err != nil {
				data, _ := json.Marshal(spfResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("cannot lookup SPF record: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			if resp.Rcode != dns.RcodeSuccess {
				data, _ := json.Marshal(spfResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("DNS query failed: %s", dns.RcodeToString[resp.Rcode]),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			var spfRecords []string
			for _, answer := range resp.Answer {
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
				data, _ := json.Marshal(spfResult{
					Found:       true,
					ErrorDetail: fmt.Sprintf("multiple SPF records found (%d); this is an invalid configuration per RFC 7208", len(spfRecords)),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			if len(spfRecords) == 1 {
				record := spfRecords[0]
				result := spfResult{
					Found:      true,
					RawRecord:  record,
					Policy:     parseSPFPolicy(record),
					Mechanisms: record,
				}
				data, _ := json.Marshal(result)
				return agent.ToolResult{Content: string(data)}, nil
			}

			data, _ := json.Marshal(spfResult{Found: false})

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
