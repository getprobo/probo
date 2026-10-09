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

package catalog

import (
	"encoding/json"
	"fmt"

	"go.probo.inc/probo/pkg/cli/api"
)

const providersQuery = `
query {
  connectorProviders {
    provider
    displayName
    apiKeySupported
    apiKeyManaged
    installSupported
    apiKeyExtraSettings { key }
    clientCredentialsSupported
    clientCredentialsTokenUrl
    clientCredentialsExtraSettings { key }
  }
}
`

type (
	Setting struct {
		Key string `json:"key"`
	}

	Provider struct {
		Provider                       string    `json:"provider"`
		DisplayName                    string    `json:"displayName"`
		APIKeySupported                bool      `json:"apiKeySupported"`
		APIKeyManaged                  bool      `json:"apiKeyManaged"`
		InstallSupported               bool      `json:"installSupported"`
		APIKeyExtraSettings            []Setting `json:"apiKeyExtraSettings"`
		ClientCredentialsSupported     bool      `json:"clientCredentialsSupported"`
		ClientCredentialsTokenURL      *string   `json:"clientCredentialsTokenUrl"`
		ClientCredentialsExtraSettings []Setting `json:"clientCredentialsExtraSettings"`
	}
)

func List(client *api.Client) ([]Provider, error) {
	data, err := client.Do(providersQuery, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		ConnectorProviders []Provider `json:"connectorProviders"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("cannot parse response: %w", err)
	}

	return resp.ConnectorProviders, nil
}

func APIKeyCreatable(provider Provider) bool {
	return provider.APIKeySupported &&
		!provider.APIKeyManaged &&
		!provider.InstallSupported &&
		len(provider.APIKeyExtraSettings) == 0
}

func ClientCredentialsCreatable(provider Provider) bool {
	return provider.ClientCredentialsSupported && len(provider.ClientCredentialsExtraSettings) == 0
}

func Find(providers []Provider, name string) (Provider, bool) {
	for _, provider := range providers {
		if provider.Provider == name {
			return provider, true
		}
	}

	return Provider{}, false
}
