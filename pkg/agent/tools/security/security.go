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
	"os"

	"go.probo.inc/probo/pkg/agent"
)

var defaultResolverAddr = resolverAddr()

func resolverAddr() string {
	if addr := os.Getenv("DNS_RESOLVER_ADDR"); addr != "" {
		return addr
	}
	return "8.8.8.8:53"
}

func BuildTools() ([]agent.Tool, error) {
	sslTool, err := CheckSSLCertificateTool()
	if err != nil {
		return nil, err
	}

	headersTool, err := CheckSecurityHeadersTool()
	if err != nil {
		return nil, err
	}

	dmarcTool, err := CheckDMARCTool()
	if err != nil {
		return nil, err
	}

	hibpTool, err := CheckBreachesTool()
	if err != nil {
		return nil, err
	}

	dnssecTool, err := CheckDNSSECTool()
	if err != nil {
		return nil, err
	}

	cspTool, err := AnalyzeCSPTool()
	if err != nil {
		return nil, err
	}

	corsTool, err := CheckCORSTool()
	if err != nil {
		return nil, err
	}

	spfTool, err := CheckSPFTool()
	if err != nil {
		return nil, err
	}

	whoisTool, err := CheckWhoisTool()
	if err != nil {
		return nil, err
	}

	dnsRecordsTool, err := CheckDNSRecordsTool()
	if err != nil {
		return nil, err
	}

	return []agent.Tool{
		sslTool,
		headersTool,
		dmarcTool,
		spfTool,
		hibpTool,
		dnssecTool,
		cspTool,
		corsTool,
		whoisTool,
		dnsRecordsTool,
	}, nil
}
