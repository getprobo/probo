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

package dnsclient

import (
	"context"
	"fmt"
	"strings"

	"codeberg.org/miekg/dns"
)

// CheckCAA verifies that CAA policy from hostname up through each parent
// toward the DNS root permits non-wildcard issuance by permittedIssuer
// (RFC 8659). Evaluation stops at the first non-empty CAA RRset.
func (c *Client) CheckCAA(ctx context.Context, hostname, permittedIssuer string) error {
	checkNames, err := HostnamesForCAA(hostname)
	if err != nil {
		return err
	}

	for _, checkName := range checkNames {
		fqdn := ToFQDN(checkName)

		msg := &dns.Msg{MsgHeader: dns.MsgHeader{ID: dns.ID(), RecursionDesired: true}}
		msg.Question = []dns.RR{&dns.CAA{Hdr: dns.Header{Name: fqdn, Class: dns.ClassINET}}}

		// Each label gets its own exchange budget so a slow empty answer at a
		// child name cannot starve the parent lookup that holds the policy.
		queryCtx, cancel := c.withExchangeTimeout(ctx)
		resp, err := c.query(queryCtx, msg)
		cancel()
		if err != nil {
			return fmt.Errorf("cannot exchange dns message for caa records: %w", err)
		}

		if resp.Rcode != dns.RcodeSuccess {
			return fmt.Errorf(
				"cannot query caa records for %q: %s",
				checkName,
				dns.RcodeToString[resp.Rcode],
			)
		}

		caaRecords := caaRecordsOwnedBy(resp, fqdn)
		if len(caaRecords) == 0 {
			continue
		}

		if caaPermitsIssuer(caaRecords, permittedIssuer) {
			return nil
		}

		return fmt.Errorf("%w: domain %q by %q", ErrCAADenied, hostname, permittedIssuer)
	}

	return nil
}

func caaRecordsOwnedBy(resp *dns.Msg, owner string) []*dns.CAA {
	var records []*dns.CAA

	for _, rr := range resp.Answer {
		caa, ok := rr.(*dns.CAA)
		if !ok || !EqualNames(caa.Hdr.Name, owner) {
			continue
		}

		records = append(records, caa)
	}

	return records
}

// caaPermitsIssuer reports whether a non-empty CAA RRset permits ordinary
// (non-wildcard) issuance by permittedIssuer per RFC 8659.
//
// Relevant rules for non-wildcard requests:
//   - Critical unrecognized tags deny issuance.
//   - Only "issue" properties authorize (case-insensitive tag match).
//   - "issuewild" is ignored (neither authorizes nor denies).
//   - Other non-critical tags (e.g. iodef) are ignored.
//   - If no "issue" property is present, issuance is permitted.
//   - An empty or malformed "issue" value is treated as an empty
//     issuer-domain-name and does not authorize any issuer.
func caaPermitsIssuer(records []*dns.CAA, permittedIssuer string) bool {
	var issueValues []string

	for _, caa := range records {
		switch {
		case strings.EqualFold(caa.Tag, "issue"):
			issueValues = append(issueValues, caa.Value)
		case strings.EqualFold(caa.Tag, "issuewild"):
			// issuewild applies only to wildcard issuance.
		default:
			if caa.Flag&1 != 0 {
				return false
			}
		}
	}

	if len(issueValues) == 0 {
		return true
	}

	for _, value := range issueValues {
		issuer, ok := parseCAAIssueValue(value)
		if !ok || issuer == "" {
			continue
		}

		if strings.EqualFold(issuer, permittedIssuer) {
			return true
		}
	}

	return false
}

// parseCAAIssueValue parses an RFC 8659 issue / issuewild property value.
// Malformed values return ok=false and must be treated like an empty
// issuer-domain-name (no authorization from that property).
func parseCAAIssueValue(value string) (issuer string, ok bool) {
	s := trimCAALeadingWSP(value)

	if s != "" && s[0] != ';' {
		end := 0
		for end < len(s) && !isCAAWSP(s[end]) && s[end] != ';' {
			end++
		}

		issuer = s[:end]
		if !isCAAIssuerDomainName(issuer) {
			return "", false
		}

		s = trimCAALeadingWSP(s[end:])
	}

	if s == "" {
		return issuer, true
	}

	if s[0] != ';' {
		return "", false
	}

	s = trimCAALeadingWSP(s[1:])
	if s == "" {
		return issuer, true
	}

	if !consumeCAAParameters(s) {
		return "", false
	}

	return issuer, true
}

func consumeCAAParameters(s string) bool {
	for {
		rest, ok := consumeCAAParameter(s)
		if !ok {
			return false
		}

		s = trimCAALeadingWSP(rest)
		if s == "" {
			return true
		}

		if s[0] != ';' {
			return false
		}

		s = trimCAALeadingWSP(s[1:])
		if s == "" {
			// Trailing ";" with no following parameter is not in the ABNF.
			return false
		}
	}
}

func consumeCAAParameter(s string) (string, bool) {
	if s == "" {
		return "", false
	}

	tagEnd := 0
	for tagEnd < len(s) && !isCAAWSP(s[tagEnd]) && s[tagEnd] != '=' {
		tagEnd++
	}

	if tagEnd == 0 || !isCAAIssuerLabel(s[:tagEnd]) {
		return "", false
	}

	s = trimCAALeadingWSP(s[tagEnd:])
	if s == "" || s[0] != '=' {
		return "", false
	}

	s = trimCAALeadingWSP(s[1:])

	valueEnd := 0
	for valueEnd < len(s) && isCAAParameterValueByte(s[valueEnd]) {
		valueEnd++
	}

	return s[valueEnd:], true
}

func isCAAIssuerDomainName(name string) bool {
	if name == "" {
		return false
	}

	labels := strings.Split(name, ".")
	for _, label := range labels {
		if !isCAAIssuerLabel(label) {
			return false
		}
	}

	return true
}

func isCAAIssuerLabel(label string) bool {
	if label == "" {
		return false
	}

	// label = (ALPHA / DIGIT) *( *("-") (ALPHA / DIGIT) )
	if !isCAAAlphaNum(label[0]) {
		return false
	}

	i := 1
	for i < len(label) {
		for i < len(label) && label[i] == '-' {
			i++
		}

		if i >= len(label) || !isCAAAlphaNum(label[i]) {
			return false
		}

		i++
	}

	return true
}

func isCAAParameterValueByte(b byte) bool {
	// value = *(%x21-3A / %x3C-7E) — printable ASCII except space and ";".
	return (b >= 0x21 && b <= 0x3A) || (b >= 0x3C && b <= 0x7E)
}

func isCAAAlphaNum(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}

func isCAAWSP(b byte) bool {
	return b == ' ' || b == '\t'
}

func trimCAALeadingWSP(s string) string {
	i := 0
	for i < len(s) && isCAAWSP(s[i]) {
		i++
	}

	return s[i:]
}
