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

import (
	"encoding/json"
	"fmt"
)

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

type GraphQLRateLimitBudget struct {
	Limit  int `json:"limit,omitempty"`
	Window int `json:"window,omitempty"`
}

type GraphQLRateLimitConfig struct {
	Default GraphQLRateLimitBudget            `json:"-"`
	Budgets map[string]GraphQLRateLimitBudget `json:"-"`
}

func (c GraphQLRateLimitConfig) IsZero() bool {
	return c.Default == (GraphQLRateLimitBudget{}) && len(c.Budgets) == 0
}

func (c GraphQLRateLimitConfig) MarshalJSON() ([]byte, error) {
	out := make(map[string]any, len(c.Budgets)+1)
	if c.Default != (GraphQLRateLimitBudget{}) {
		out["default"] = c.Default
	}

	for name, budget := range c.Budgets {
		out[name] = budget
	}

	return json.Marshal(out)
}

func (c *GraphQLRateLimitConfig) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("cannot parse rate limit: %w", err)
	}

	if _, ok := raw["enabled"]; ok {
		return fmt.Errorf("rate limit %q is unknown", "enabled")
	}

	if value, ok := raw["default"]; ok {
		budget, err := parseRateLimitBudget("default", value)
		if err != nil {
			return err
		}

		c.Default = budget

		delete(raw, "default")
	}

	if len(raw) == 0 {
		return nil
	}

	c.Budgets = make(map[string]GraphQLRateLimitBudget, len(raw))
	for name, value := range raw {
		budget, err := parseRateLimitBudget(name, value)
		if err != nil {
			return err
		}

		if budget == (GraphQLRateLimitBudget{}) {
			continue
		}

		c.Budgets[name] = budget
	}

	return nil
}

func parseRateLimitBudget(name string, value json.RawMessage) (GraphQLRateLimitBudget, error) {
	var budget GraphQLRateLimitBudget
	if err := json.Unmarshal(value, &budget); err != nil {
		return GraphQLRateLimitBudget{}, fmt.Errorf("cannot parse rate limit %q: %w", name, err)
	}

	if budget.Limit == 0 && budget.Window == 0 {
		return GraphQLRateLimitBudget{}, nil
	}

	if budget.Limit < 1 || budget.Window < 1 {
		return GraphQLRateLimitBudget{}, fmt.Errorf("rate limit %q requires limit and window of at least 1", name)
	}

	return budget, nil
}

type GraphQLConfig struct {
	ParserTokenLimit  int                    `json:"parser-token-limit"`
	ComplexityLimit   int                    `json:"complexity-limit"`
	QueryCacheSize    int                    `json:"query-cache-size"`
	DisableSuggestion bool                   `json:"disable-suggestion"`
	RateLimit         GraphQLRateLimitConfig `json:"rate-limit,omitzero"`
}

type APIConfig struct {
	Addr              string              `json:"addr,omitempty"`
	ProxyProtocol     ProxyProtocolConfig `json:"proxy-protocol,omitzero"`
	Cors              CorsConfig          `json:"cors,omitzero"`
	ExtraHeaderFields map[string]string   `json:"extra-header-fields,omitzero,omitempty"`
	GraphQL           GraphQLConfig       `json:"graphql,omitzero"`
}
