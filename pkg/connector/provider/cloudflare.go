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

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func cloudflareRegistration() *Registration {
	return &Registration{
		Provider: coredata.ConnectorProviderCloudflare,
		InitialAccountFunc: initialAccount(
			func(s coredata.CloudflareConnectorSettings) string {
				return s.AccountID
			},
		),
		DisplayName:      "Cloudflare",
		DocumentationURL: accessReviewDocsURL("cloudflare"),
		Endpoints: Endpoints{
			// Kept so an APIBase/Probe override moves this check with the
			// driver. probeCloudflare reads it; a plain GET would treat
			// Cloudflare's 400 (a malformed token) and a 200 whose token is
			// not active as connected.
			Probe:   "https://api.cloudflare.com/client/v4/user/tokens/verify",
			APIBase: "https://api.cloudflare.com/client/v4",
		},
		Probe:  probeCloudflare,
		APIKey: &APIKeyConfig{},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.CloudflareConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read cloudflare connector settings: %w", err)
			}

			if s.AccountID == "" {
				return nil, fmt.Errorf("cannot create cloudflare driver: account_id is required")
			}

			return drivers.NewCloudflareDriver(c, s.AccountID, ep.APIBase), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, ep Endpoints) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.CloudflareConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read cloudflare connector settings", log.Error(err))
				return nil
			}

			return drivers.NewCloudflareNameResolver(c, s.AccountID, ep.APIBase)
		},
		SetOrganizationSettings: func(c *coredata.Connector, accountID string) error {
			return c.SetSettings(&coredata.CloudflareConnectorSettings{AccountID: accountID})
		},
	}
}

// probeCloudflare checks the token verify endpoint and reads whether the
// token is active.
//
// A plain GET is not enough. Cloudflare answers a malformed Authorization
// header with 400, which the default probe reads as connected, and it
// answers an expired or disabled token with 200 and result.status set to
// something other than "active". Listing accounts then fails, so the access
// review connections page reports the credential as invalid while
// connectionStatus stays connected. Only "active" counts. The body is not
// returned: provider text must not leave this function.
func probeCloudflare(
	ctx context.Context,
	httpClient *http.Client,
	_ *coredata.Connector,
	ep Endpoints,
) error {
	if ep.Probe == "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep.Probe, nil)
	if err != nil {
		return fmt.Errorf("cannot create cloudflare probe request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cloudflare probe request failed: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	// 400 is the malformed-token answer. 401 is a well-formed dead token.
	// 403 means Cloudflare accepted the credential and refused the call.
	if resp.StatusCode == http.StatusBadRequest ||
		resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusForbidden {
		return newCredentialRejected(resp.StatusCode)
	}

	// Any other non-2xx keeps the default probe contract: it is not a
	// credential verdict, so it counts as connected. A 2xx still has to say
	// the token is active.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, rejectionBodyLimit))
	if err != nil {
		return fmt.Errorf("cannot read cloudflare probe response: %w", err)
	}

	if respondsWithHTML(bytes.NewReader(body)) {
		return &NotAnAPIEndpointError{StatusCode: resp.StatusCode}
	}

	var parsed struct {
		Success bool `json:"success"`
		Result  struct {
			Status string `json:"status"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Errorf("cannot decode cloudflare probe response: %w", err)
	}

	if !parsed.Success || parsed.Result.Status != "active" {
		return newCredentialRejected(http.StatusUnauthorized)
	}

	return nil
}
