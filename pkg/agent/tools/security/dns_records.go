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
	dnsRecordsParams struct {
		Domain string `json:"domain" jsonschema:"The domain to query DNS records for (e.g. example.com)"`
	}

	dnsRecordsResult struct {
		A           []string `json:"a_records,omitempty"`
		AAAA        []string `json:"aaaa_records,omitempty"`
		MX          []string `json:"mx_records,omitempty"`
		CNAME       []string `json:"cname_records,omitempty"`
		TXT         []string `json:"txt_records,omitempty"`
		NS          []string `json:"ns_records,omitempty"`
		ErrorDetail string   `json:"error_detail,omitempty"`
	}

	queryOption func(*dns.MsgHeader)
)

func CheckDNSRecordsTool() agent.Tool {
	return agent.FunctionTool(
		"check_dns_records",
		"Query DNS records for a domain (A, AAAA, MX, CNAME, TXT, NS). Reveals hosting provider, email provider, and additional security signals.",
		func(ctx context.Context, p dnsRecordsParams) (agent.ToolResult, error) {
			fqdn := p.Domain
			if !strings.HasSuffix(fqdn, ".") {
				fqdn = fqdn + "."
			}

			hdr := dns.Header{Name: fqdn, Class: dns.ClassINET}
			client := dns.NewClient()
			var result dnsRecordsResult
			var errs []string

			// A records.
			if answers, err := queryDNS(ctx, client, &dns.A{Hdr: hdr}); err != nil {
				errs = append(errs, fmt.Sprintf("A query failed: %s", err))
			} else {
				for _, rr := range answers {
					if a, ok := rr.(*dns.A); ok {
						result.A = append(result.A, a.A.String())
					}
				}
			}

			// AAAA records.
			if answers, err := queryDNS(ctx, client, &dns.AAAA{Hdr: hdr}); err != nil {
				errs = append(errs, fmt.Sprintf("AAAA query failed: %s", err))
			} else {
				for _, rr := range answers {
					if aaaa, ok := rr.(*dns.AAAA); ok {
						result.AAAA = append(result.AAAA, aaaa.AAAA.String())
					}
				}
			}

			// MX records.
			if answers, err := queryDNS(ctx, client, &dns.MX{Hdr: hdr}); err != nil {
				errs = append(errs, fmt.Sprintf("MX query failed: %s", err))
			} else {
				for _, rr := range answers {
					if mx, ok := rr.(*dns.MX); ok {
						result.MX = append(result.MX, fmt.Sprintf("%d %s", mx.Preference, strings.TrimSuffix(mx.Mx, ".")))
					}
				}
			}

			// CNAME records.
			if answers, err := queryDNS(ctx, client, &dns.CNAME{Hdr: hdr}); err != nil {
				errs = append(errs, fmt.Sprintf("CNAME query failed: %s", err))
			} else {
				for _, rr := range answers {
					if cname, ok := rr.(*dns.CNAME); ok {
						result.CNAME = append(result.CNAME, strings.TrimSuffix(cname.Target, "."))
					}
				}
			}

			// TXT records.
			if answers, err := queryDNS(ctx, client, &dns.TXT{Hdr: hdr}); err != nil {
				errs = append(errs, fmt.Sprintf("TXT query failed: %s", err))
			} else {
				for _, rr := range answers {
					if txt, ok := rr.(*dns.TXT); ok {
						result.TXT = append(result.TXT, strings.Join(txt.Txt, ""))
					}
				}
			}

			// NS records.
			if answers, err := queryDNS(ctx, client, &dns.NS{Hdr: hdr}); err != nil {
				errs = append(errs, fmt.Sprintf("NS query failed: %s", err))
			} else {
				for _, rr := range answers {
					if ns, ok := rr.(*dns.NS); ok {
						result.NS = append(result.NS, strings.TrimSuffix(ns.Ns, "."))
					}
				}
			}

			if len(errs) > 0 {
				result.ErrorDetail = strings.Join(errs, "; ")
			}

			return agent.ResultJSON(result), nil
		},
	)
}

func withDNSSEC() queryOption {
	return func(h *dns.MsgHeader) {
		h.UDPSize = 4096
		h.Security = true
	}
}

func queryDNS(ctx context.Context, client *dns.Client, question dns.RR, opts ...queryOption) ([]dns.RR, error) {
	msg := &dns.Msg{
		MsgHeader: dns.MsgHeader{
			ID:               dns.ID(),
			RecursionDesired: true,
		},
	}
	for _, opt := range opts {
		opt(&msg.MsgHeader)
	}
	msg.Question = []dns.RR{question}

	resp, _, err := client.Exchange(ctx, msg, "udp", defaultResolverAddr)
	if err == nil && resp.Truncated {
		resp, _, err = client.Exchange(ctx, msg, "tcp", defaultResolverAddr)
	}
	if err != nil {
		return nil, err
	}

	if resp.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("cannot execute DNS query: %s", dns.RcodeToString[resp.Rcode])
	}

	return resp.Answer, nil
}
