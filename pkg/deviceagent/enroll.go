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

package deviceagent

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// ErrServerURLMismatch is returned when install is re-run with a --server
// that differs from the persisted config.
var ErrServerURLMismatch = errors.New("server URL does not match persisted config; uninstall first")

// LoadOrExchangeAPIKey returns a persisted API key for install.
// If agent.key already exists, it is returned without contacting the server
// only when serverURL matches the server_url persisted in config.json.
// Otherwise the enrollment token is exchanged, the key and server binding
// are saved to disk immediately, and then the key is returned.
func LoadOrExchangeAPIKey(
	ctx context.Context,
	dir string,
	client *Client,
	serverURL string,
	enrollmentToken string,
) (string, error) {
	normalized, err := NormalizeServerURL(serverURL)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}

	apiKey, err := LoadAPIKey(dir)
	if err == nil {
		if err := validatePersistedServerURL(dir, normalized); err != nil {
			return "", err
		}

		return apiKey, nil
	}

	if !errors.Is(err, ErrKeyNotFound) {
		return "", fmt.Errorf("cannot load device api key: %w", err)
	}

	apiKey, err = client.ExchangeEnrollmentToken(ctx, enrollmentToken)
	if err != nil {
		return "", fmt.Errorf("cannot exchange enrollment token: %w", err)
	}

	if err := SaveAPIKey(dir, apiKey); err != nil {
		return "", fmt.Errorf("cannot save device api key: %w", err)
	}

	if err := SaveConfig(dir, &Config{ServerURL: normalized}); err != nil {
		if delErr := DeleteAPIKey(dir); delErr != nil {
			return "", fmt.Errorf(
				"cannot save device config: %w (also cannot delete device api key: %v)",
				err,
				delErr,
			)
		}

		return "", fmt.Errorf("cannot save device config: %w", err)
	}

	return apiKey, nil
}

func validatePersistedServerURL(dir, serverURL string) error {
	normalized, err := NormalizeServerURL(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrServerURLMismatch
		}

		return fmt.Errorf("cannot load device config: %w", err)
	}

	if cfg.ServerURL == "" {
		return ErrServerURLMismatch
	}

	persisted, err := NormalizeServerURL(cfg.ServerURL)
	if err != nil {
		return fmt.Errorf("invalid persisted server URL: %w", err)
	}

	if persisted != normalized {
		return ErrServerURLMismatch
	}

	return nil
}
