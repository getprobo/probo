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

type CorsConfig struct {
	AllowedOrigins []string `json:"allowed-origins,omitzero,omitempty"`
}

func (c CorsConfig) IsZero() bool {
	return len(c.AllowedOrigins) == 0
}

type ProxyProtocolConfig struct {
	TrustedProxies []string `json:"trusted-proxies,omitzero,omitempty"`
}

func (c ProxyProtocolConfig) IsZero() bool {
	return len(c.TrustedProxies) == 0
}

type GraphQLConfig struct {
	ParserTokenLimit  int  `json:"parser-token-limit"`
	ComplexityLimit   int  `json:"complexity-limit"`
	QueryCacheSize    int  `json:"query-cache-size"`
	DisableSuggestion bool `json:"disable-suggestion"`
}

type APIConfig struct {
	Addr              string              `json:"addr,omitempty"`
	ProxyProtocol     ProxyProtocolConfig `json:"proxy-protocol,omitzero"`
	Cors              CorsConfig          `json:"cors,omitzero"`
	ExtraHeaderFields map[string]string   `json:"extra-header-fields,omitzero,omitempty"`
	GraphQL           GraphQLConfig       `json:"graphql,omitzero"`
}
