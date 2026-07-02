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

package logout

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/version"
)

type oidcDiscovery struct {
	RevocationEndpoint string `json:"revocation_endpoint"`
}

func NewCmdLogout(f *cmdutil.Factory) *cobra.Command {
	var flagHost string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out of a Probo host",
		Example: `  # Log out of the active host
  prb auth logout

  # Log out of a specific host
  prb auth logout --hostname app.getprobo.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			if flagHost == "" {
				host, _, err := cfg.DefaultHost()
				if err != nil {
					return err
				}

				hosts := slices.Sorted(maps.Keys(cfg.Hosts))
				if len(hosts) > 1 && f.IOStreams.IsInteractive() {
					options := make([]huh.Option[string], len(hosts))
					for i, h := range hosts {
						options[i] = huh.NewOption(h, h)
					}

					err := huh.NewSelect[string]().
						Title("Select a host to log out of").
						Options(options...).
						Value(&flagHost).
						Run()
					if err != nil {
						return err
					}
				} else {
					flagHost = host
				}
			}

			hc, ok := cfg.Hosts[flagHost]
			if !ok {
				return fmt.Errorf("not logged in to %s", flagHost)
			}

			revokeTokens(flagHost, hc, f)

			delete(cfg.Hosts, flagHost)

			if cfg.ActiveHost == flagHost {
				cfg.ActiveHost = ""
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.ErrOut,
				"Logged out of %s\n",
				flagHost,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagHost, "hostname", "", "Probo hostname to log out of")

	return cmd
}

func revokeTokens(host string, hc *config.HostConfig, f *cmdutil.Factory) {
	baseURL := normalizeHostToURL(host)
	httpClient := &http.Client{Timeout: 10 * time.Second}

	discovery, err := fetchRevocationEndpoint(httpClient, baseURL)
	if err != nil || discovery.RevocationEndpoint == "" {
		return
	}

	if hc.RefreshToken != "" {
		_ = revokeToken(
			httpClient,
			discovery.RevocationEndpoint,
			hc.RefreshToken,
			"refresh_token",
		)
	}

	if hc.Token != "" {
		_ = revokeToken(
			httpClient,
			discovery.RevocationEndpoint,
			hc.Token,
			"access_token",
		)
	}
}

func fetchRevocationEndpoint(client *http.Client, baseURL string) (*oidcDiscovery, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		baseURL+"/.well-known/openid-configuration",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("User-Agent", version.UserAgent("prb"))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch discovery document: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery endpoint returned HTTP %d", resp.StatusCode)
	}

	var discovery oidcDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("cannot decode discovery document: %w", err)
	}

	return &discovery, nil
}

func revokeToken(
	client *http.Client,
	endpoint string,
	token string,
	tokenTypeHint string,
) error {
	data := url.Values{
		"token":           {token},
		"token_type_hint": {tokenTypeHint},
		"client_id":       {config.CLIClientID},
	}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return fmt.Errorf("cannot create revocation request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", version.UserAgent("prb"))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot send revocation request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revocation endpoint returned HTTP %d", resp.StatusCode)
	}

	return nil
}

func normalizeHostToURL(host string) string {
	lower := strings.ToLower(host)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return strings.TrimRight(host, "/")
	}

	return "https://" + strings.TrimRight(host, "/")
}
