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

package shared

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

type (
	// ServiceAccount contains fields displayed by service-account commands.
	ServiceAccount struct {
		ID             string   `json:"id"`
		OrganizationID string   `json:"organizationId"`
		Name           string   `json:"name"`
		Description    *string  `json:"description"`
		Scopes         []string `json:"scopes"`
		DisabledAt     *string  `json:"disabledAt"`
		CreatedAt      string   `json:"createdAt"`
		UpdatedAt      string   `json:"updatedAt"`
	}

	// Credential contains non-secret service-account credential metadata.
	Credential struct {
		ID               string   `json:"id"`
		ServiceAccountID string   `json:"serviceAccountId"`
		Name             string   `json:"name"`
		Scopes           []string `json:"scopes"`
		ExpiresAt        string   `json:"expiresAt"`
		LastUsedAt       *string  `json:"lastUsedAt"`
		RevokedAt        *string  `json:"revokedAt"`
		CreatedAt        string   `json:"createdAt"`
	}
)

// NewClient creates a client for the Connect GraphQL API.
func NewClient(cfg *config.Config, host string, hc *config.HostConfig) *api.Client {
	return api.NewClient(
		host,
		hc.Token,
		"/api/connect/v1/graphql",
		cfg.HTTPTimeoutDuration(),
		cmdutil.TokenRefreshOption(cfg, host, hc),
	)
}

// ResolveOrganization selects an explicit organization or the configured default.
func ResolveOrganization(flagOrg string, hc *config.HostConfig) (string, error) {
	if flagOrg != "" {
		return flagOrg, nil
	}

	if hc.Organization == "" {
		return "", fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
	}

	return hc.Organization, nil
}

// PrintServiceAccount writes a human-readable service account.
func PrintServiceAccount(out io.Writer, account *ServiceAccount) {
	bold := lipgloss.NewStyle().Bold(true)
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)
	description := ""
	disabledAt := "Active"

	if account.Description != nil {
		description = *account.Description
	}
	if account.DisabledAt != nil {
		disabledAt = cmdutil.FormatTime(*account.DisabledAt)
	}

	_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(account.Name))
	_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), account.ID)
	_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Organization ID:"), account.OrganizationID)
	_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Description:"), description)
	_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Scopes:"), strings.Join(account.Scopes, ", "))
	_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Disabled:"), disabledAt)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(account.CreatedAt))
	_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(account.UpdatedAt))
}
