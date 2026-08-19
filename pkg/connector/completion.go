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

package connector

import (
	"context"
	"fmt"
	"net/http"
)

const (
	// CompletionMetadataGitHubOrganization is the installed GitHub org login
	// surfaced by a GitHub App callback for connector settings persistence.
	CompletionMetadataGitHubOrganization = "github_organization"
)

// CompletionState is the protocol-neutral result of finishing a connector
// install or reconnect callback. HTTP handlers use it to create or update the
// connector row and build the post-auth redirect.
type CompletionState struct {
	Provider         string
	Protocol         ProtocolType
	Connection       Connection
	OrganizationID   string
	ConnectorID      string
	ContinueURL      string
	Site             string
	ProviderMetadata map[string]string
}

// CompleteFromState routes the callback to the connector registered for the
// signed state token's protocol and returns a normalized completion payload.
func (r *ConnectorRegistry) CompleteFromState(
	ctx context.Context,
	req *http.Request,
) (*CompletionState, error) {
	stateToken := req.URL.Query().Get("state")
	if stateToken == "" {
		return nil, fmt.Errorf("missing state parameter")
	}

	if IsGitHubAppState(stateToken) {
		return r.completeGitHubAppFromState(ctx, req)
	}

	return r.completeOAuth2FromState(ctx, req, stateToken)
}

func (r *ConnectorRegistry) completeGitHubAppFromState(
	ctx context.Context,
	req *http.Request,
) (*CompletionState, error) {
	connection, state, err := r.CompleteGitHubApp(ctx, req)
	if err != nil {
		return nil, err
	}

	metadata := map[string]string{}
	if state.Organization != "" {
		metadata[CompletionMetadataGitHubOrganization] = state.Organization
	}

	return &CompletionState{
		Provider:         GitHubProvider,
		Protocol:         ProtocolGitHubApp,
		Connection:       connection,
		OrganizationID:   state.OrganizationID,
		ConnectorID:      state.ConnectorID,
		ContinueURL:      state.ContinueURL,
		ProviderMetadata: metadata,
	}, nil
}

func (r *ConnectorRegistry) completeOAuth2FromState(
	ctx context.Context,
	req *http.Request,
	stateToken string,
) (*CompletionState, error) {
	provider, err := ExtractProviderFromState(stateToken)
	if err != nil {
		return nil, fmt.Errorf("cannot extract provider from state: %w", err)
	}

	connection, state, err := r.CompleteWithState(ctx, provider, req)
	if err != nil {
		return nil, err
	}

	return &CompletionState{
		Provider:         provider,
		Protocol:         connection.Type(),
		Connection:       connection,
		OrganizationID:   state.OrganizationID,
		ConnectorID:      state.ConnectorID,
		ContinueURL:      state.ContinueURL,
		Site:             state.Site,
		ProviderMetadata: state.ProviderMetadata,
	}, nil
}
