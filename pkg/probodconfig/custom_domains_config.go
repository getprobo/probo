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

package probodconfig

type CustomDomainsConfig struct {
	RenewalInterval   int    `json:"renewal-interval"`
	ProvisionInterval int    `json:"provision-interval"`
	ResolverAddr      string `json:"resolver-addr,omitempty"`
	CnameTarget       string `json:"cname-target"`
	CAAIssuerDomain   string `json:"caa-issuer-domain"`
	// SkipCNAMECheck turns the CNAME pre-check into a warning instead of a
	// hard block. A hostname served through a proxying CDN never exposes its
	// origin CNAME in public DNS, so the check can never pass even when HTTP
	// routing works. The ACME HTTP-01 challenge still proves control of the
	// hostname; what is lost is the cheap upfront filter that keeps obviously
	// misconfigured domains from consuming ACME rate limits, so only enable
	// this on a self-hosted, single-tenant deployment.
	SkipCNAMECheck bool       `json:"skip-cname-check,omitempty"`
	ACME           ACMEConfig `json:"acme,omitzero"`
}

type ACMEConfig struct {
	Directory  string `json:"directory,omitempty"`
	Email      string `json:"email,omitempty"`
	KeyType    string `json:"key-type,omitempty"`
	AccountKey string `json:"account-key,omitempty"`
	RootCA     string `json:"root-ca,omitempty"`
}
