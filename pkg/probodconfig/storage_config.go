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
	"fmt"
	"net/url"
	"strings"
)

type (
	StorageProvider string

	StorageConfig struct {
		Provider        StorageProvider `json:"provider,omitempty"`
		Bucket          string          `json:"bucket,omitempty"`
		AzureAccountURL string          `json:"azure-account-url,omitempty"`
	}
)

const (
	StorageProviderS3    StorageProvider = "s3"
	StorageProviderAzure StorageProvider = "azure"
)

func (c StorageConfig) Validate() error {
	provider := c.Provider
	if provider == "" {
		provider = StorageProviderS3
	}

	switch provider {
	case StorageProviderS3:
		return nil
	case StorageProviderAzure:
		// The Azure SDK receives the value as configured.
		if c.AzureAccountURL != strings.TrimSpace(c.AzureAccountURL) {
			return fmt.Errorf("invalid Azure Storage account URL: must not have surrounding whitespace")
		}

		accountURL, err := url.Parse(c.AzureAccountURL)
		if err != nil {
			return fmt.Errorf("invalid Azure Storage account URL: %w", err)
		}

		if accountURL.Scheme != "https" || accountURL.Hostname() == "" ||
			accountURL.User != nil || accountURL.RawQuery != "" || accountURL.Fragment != "" {
			return fmt.Errorf("invalid Azure Storage account URL: must be HTTPS without userinfo, query, or fragment")
		}

		return nil
	default:
		return fmt.Errorf("unsupported object storage provider %q", provider)
	}
}
