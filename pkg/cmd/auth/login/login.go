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

package login

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/version"
)

const (
	hostEU = "eu.console.getprobo.com"
	hostUS = "us.console.getprobo.com"

	regionEU     = "eu"
	regionUS     = "us"
	regionCustom = "custom"
)

type (
	oidcDiscovery struct {
		DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
		TokenEndpoint               string `json:"token_endpoint"`
	}

	deviceAuthResponse struct {
		DeviceCode              string `json:"device_code"`
		UserCode                string `json:"user_code"`
		VerificationURI         string `json:"verification_uri"`
		VerificationURIComplete string `json:"verification_uri_complete"`
		ExpiresIn               int    `json:"expires_in"`
		Interval                int    `json:"interval"`
	}

	tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
		Scope        string `json:"scope,omitempty"`
	}

	tokenErrorResponse struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description,omitempty"`
	}
)

func NewCmdLogin(f *cmdutil.Factory) *cobra.Command {
	var (
		flagHost         string
		flagOrganization string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a Probo host",
		Example: `  # Interactive login (select region, opens browser for device authorization)
  prb auth login

  # Login to Probo EU
  prb auth login --hostname eu.console.getprobo.com

  # Login to Probo US
  prb auth login --hostname us.console.getprobo.com

  # Login to a self-hosted instance
  prb auth login --hostname probo.example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if f.IOStreams.IsInteractive() && flagHost == "" {
				var region string

				err := huh.NewSelect[string]().
					Title("Where is your Probo account hosted?").
					Options(
						huh.NewOption("Probo EU (eu.console.getprobo.com)", regionEU),
						huh.NewOption("Probo US (us.console.getprobo.com)", regionUS),
						huh.NewOption("Other (custom domain)", regionCustom),
					).
					Value(&region).
					Run()
				if err != nil {
					return err
				}

				switch region {
				case regionEU:
					flagHost = hostEU
				case regionUS:
					flagHost = hostUS
				case regionCustom:
					err := huh.NewInput().
						Title("Probo hostname").
						Placeholder("probo.example.com").
						Value(&flagHost).
						Run()
					if err != nil {
						return err
					}

					if flagHost == "" {
						return fmt.Errorf("hostname is required")
					}
				}
			}

			if flagHost == "" {
				flagHost = hostEU
			}

			baseURL := normalizeHostToURL(flagHost)
			httpClient := &http.Client{Timeout: 30 * time.Second}

			_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "Discovering OAuth2 endpoints on %s...\n", flagHost)

			discovery, err := fetchDiscovery(httpClient, baseURL)
			if err != nil {
				return fmt.Errorf("cannot discover OAuth2 endpoints: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			deviceAuth, err := requestDeviceCode(
				httpClient,
				discovery.DeviceAuthorizationEndpoint,
				config.CLIClientID,
			)
			if err != nil {
				return fmt.Errorf("cannot start device authorization: %w", err)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.ErrOut,
				"\nOpen the following URL in your browser and enter the code:\n\n  %s\n\n  Code: %s\n\n",
				deviceAuth.VerificationURI,
				deviceAuth.UserCode,
			)

			if f.IOStreams.IsInteractive() {
				openBrowser(deviceAuth.VerificationURIComplete, cfg.Browser)
			}

			_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "Waiting for authorization...")

			token, err := pollForToken(
				httpClient,
				discovery.TokenEndpoint,
				config.CLIClientID,
				deviceAuth,
			)
			if err != nil {
				_, _ = fmt.Fprintln(f.IOStreams.ErrOut)
				return fmt.Errorf("cannot complete device authorization: %w", err)
			}

			_, _ = fmt.Fprintln(f.IOStreams.ErrOut)

			if f.IOStreams.IsInteractive() && flagOrganization == "" {
				_, _ = fmt.Fprintln(f.IOStreams.ErrOut, "Loading organizations...")
				orgs, orgsErr := fetchOrganizations(baseURL, token.AccessToken)

				if orgsErr == nil && len(orgs) > 0 {
					selected := orgs[0].ID
					options := make([]huh.Option[string], 0, len(orgs)+1)
					for _, org := range orgs {
						options = append(
							options,
							huh.NewOption(
								fmt.Sprintf("%s (%s)", org.Name, org.ID),
								org.ID,
							),
						)
					}
					options = append(
						options,
						huh.NewOption("Skip (no default)", ""),
					)

					err = huh.NewSelect[string]().
						Title("Default organization").
						Value(&selected).
						Options(options...).
						Run()
					if err != nil {
						return err
					}

					flagOrganization = selected
				}
			}

			cfg.Hosts[flagHost] = &config.HostConfig{
				Token:         token.AccessToken,
				RefreshToken:  token.RefreshToken,
				TokenEndpoint: discovery.TokenEndpoint,
				Organization:  flagOrganization,
			}
			cfg.ActiveHost = flagHost

			if err := cfg.Save(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.ErrOut,
				"Logged in to %s\n",
				flagHost,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(
		&flagHost,
		"hostname",
		"",
		"Probo hostname (e.g. eu.console.getprobo.com, us.console.getprobo.com)",
	)
	cmd.Flags().StringVar(
		&flagOrganization,
		"org",
		"",
		"Default organization ID",
	)

	return cmd
}

func normalizeHostToURL(host string) string {
	lower := strings.ToLower(host)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return strings.TrimRight(host, "/")
	}
	return "https://" + strings.TrimRight(host, "/")
}

func fetchDiscovery(client *http.Client, baseURL string) (*oidcDiscovery, error) {
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

	if discovery.DeviceAuthorizationEndpoint == "" {
		return nil, fmt.Errorf("server does not support device authorization")
	}

	if discovery.TokenEndpoint == "" {
		return nil, fmt.Errorf("server does not advertise a token endpoint")
	}

	return &discovery, nil
}

func requestDeviceCode(
	client *http.Client,
	endpoint string,
	clientID string,
) (*deviceAuthResponse, error) {
	values := url.Values{
		"client_id": {clientID},
		"scope":     {"openid profile email offline_access"},
	}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", version.UserAgent("prb"))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot request device code: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device authorization returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var deviceAuth deviceAuthResponse
	if err := json.Unmarshal(body, &deviceAuth); err != nil {
		return nil, fmt.Errorf("cannot decode device authorization response: %w", err)
	}

	return &deviceAuth, nil
}

func pollForToken(
	client *http.Client,
	tokenEndpoint string,
	clientID string,
	deviceAuth *deviceAuthResponse,
) (*tokenResponse, error) {
	interval := time.Duration(deviceAuth.Interval) * time.Second
	if interval < 1*time.Second {
		interval = 5 * time.Second
	}

	deadline := time.Now().Add(time.Duration(deviceAuth.ExpiresIn) * time.Second)

	for {
		time.Sleep(interval)

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("device code expired, please try again")
		}

		values := url.Values{
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"client_id":   {clientID},
			"device_code": {deviceAuth.DeviceCode},
		}

		req, err := http.NewRequest(
			http.MethodPost,
			tokenEndpoint,
			strings.NewReader(values.Encode()),
		)
		if err != nil {
			return nil, fmt.Errorf("cannot create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", version.UserAgent("prb"))

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("cannot poll token endpoint: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("cannot read token response: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			var token tokenResponse
			if err := json.Unmarshal(body, &token); err != nil {
				return nil, fmt.Errorf("cannot decode token response: %w", err)
			}
			return &token, nil
		}

		var errResp tokenErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("cannot decode error response: %w", err)
		}

		switch errResp.Error {
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5 * time.Second
			continue
		case "expired_token":
			return nil, fmt.Errorf("device code expired, please try again")
		case "access_denied":
			return nil, fmt.Errorf("authorization denied by user")
		default:
			return nil, fmt.Errorf("token error: %s: %s", errResp.Error, errResp.ErrorDescription)
		}
	}
}

const viewerOrganizationsQuery = `
query($first: Int, $filter: ProfileFilter) {
  viewer {
    profiles(first: $first, filter: $filter) {
      edges {
        node {
          organization {
            id
            name
          }
        }
      }
    }
  }
}
`

type viewerOrganization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func fetchOrganizations(baseURL string, token string) ([]viewerOrganization, error) {
	client := api.NewClient(
		baseURL,
		token,
		"/api/connect/v1/graphql",
		config.DefaultHTTPTimeout,
	)

	variables := map[string]any{
		"first": 100,
		"filter": map[string]any{
			"state": "ACTIVE",
		},
	}

	data, err := client.Do(viewerOrganizationsQuery, variables)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch organizations: %w", err)
	}

	var resp struct {
		Viewer struct {
			Profiles struct {
				Edges []struct {
					Node struct {
						Organization *viewerOrganization `json:"organization"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"profiles"`
		} `json:"viewer"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("cannot parse organizations response: %w", err)
	}

	orgs := make([]viewerOrganization, 0, len(resp.Viewer.Profiles.Edges))
	for _, edge := range resp.Viewer.Profiles.Edges {
		if edge.Node.Organization != nil {
			orgs = append(orgs, *edge.Node.Organization)
		}
	}

	return orgs, nil
}

func openBrowser(url, browser string) {
	if browser != "" {
		_ = exec.Command("sh", "-c", browser+" \"$0\"", url).Start()
		return
	}

	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", url).Start()
	case "linux":
		_ = exec.Command("xdg-open", url).Start()
	case "windows":
		_ = exec.Command(
			"rundll32",
			"url.dll,FileProtocolHandler",
			url,
		).Start()
	}
}
