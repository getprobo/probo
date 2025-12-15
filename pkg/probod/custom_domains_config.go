// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package probod

type customDomainsConfig struct {
	RenewalInterval   int        `json:"renewal-interval"`
	ProvisionInterval int        `json:"provision-interval"`
	ResolverAddr      string     `json:"resolver-addr"`
	CnameTarget       string     `json:"cname-target"`
	ACME              acmeConfig `json:"acme"`
}

type acmeConfig struct {
	Directory  string `json:"directory"`
	Email      string `json:"email"`
	KeyType    string `json:"key-type"`
	AccountKey string `json:"account-key"`
	RootCA     string `json:"root-ca"`
}
