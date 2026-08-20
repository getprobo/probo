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
	"fmt"
	"net/url"
)

// OrganizationSelector is implemented by connections whose auth model can
// participate in a multi-org picker when the provider registers ListOrgs.
type OrganizationSelector interface {
	SupportsOrganizationPicker() bool
}

// ScopeGranter is implemented by connections that carry an OAuth-style scope
// grant that can be compared against provider requirements.
type ScopeGranter interface {
	SupportsScopeGrantCheck() bool
}

// Reconnector is implemented by connections that can be refreshed through the
// connector initiate/complete flow.
type Reconnector interface {
	SupportsReconnect() bool
}

// ProbeURLProvider optionally overrides the provider's default health-check
// URL for this connection's auth model.
type ProbeURLProvider interface {
	ProbeURL(apiBase, defaultProbe string) (string, error)
}

// CapabilityProbe returns a zero-value Connection used only to answer
// capability questions for a stored protocol when the live connection is not
// loaded. Behavior lives on the connection types; this switch is the factory.
func CapabilityProbe(protocol ProtocolType) Connection {
	switch protocol {
	case ProtocolOAuth2:
		return &OAuth2Connection{}
	case ProtocolAPIKey:
		return &APIKeyConnection{}
	case ProtocolGitHubApp:
		return &GitHubAppConnection{}
	default:
		return nil
	}
}

func connectionOrProbe(conn Connection, protocol ProtocolType) Connection {
	if conn != nil {
		return conn
	}

	return CapabilityProbe(protocol)
}

// SupportsOrganizationPicker reports whether conn's auth model can drive a
// provider ListOrgs picker.
func SupportsOrganizationPicker(conn Connection) bool {
	selector, ok := conn.(OrganizationSelector)
	return ok && selector.SupportsOrganizationPicker()
}

// SupportsOrganizationPickerForProtocol answers the picker capability when
// only the stored protocol is available.
func SupportsOrganizationPickerForProtocol(protocol ProtocolType) bool {
	return SupportsOrganizationPicker(CapabilityProbe(protocol))
}

// SupportsScopeGrantCheck reports whether missing-scope checks apply to conn.
func SupportsScopeGrantCheck(conn Connection) bool {
	granter, ok := conn.(ScopeGranter)
	return ok && granter.SupportsScopeGrantCheck()
}

// SupportsScopeGrantCheckFor reports whether missing-scope checks apply,
// using protocol as a fallback when conn is nil.
func SupportsScopeGrantCheckFor(conn Connection, protocol ProtocolType) bool {
	return SupportsScopeGrantCheck(connectionOrProbe(conn, protocol))
}

// SupportsReconnect reports whether conn can be refreshed via initiate.
func SupportsReconnect(conn Connection) bool {
	reconnector, ok := conn.(Reconnector)
	return ok && reconnector.SupportsReconnect()
}

// SupportsReconnectFor reports reconnect support, using protocol when conn is nil.
func SupportsReconnectFor(conn Connection, protocol ProtocolType) bool {
	return SupportsReconnect(connectionOrProbe(conn, protocol))
}

// ResolveProbeURL returns conn's health-check URL, or defaultProbe when the
// connection does not override it.
func ResolveProbeURL(conn Connection, apiBase, defaultProbe string) (string, error) {
	if provider, ok := conn.(ProbeURLProvider); ok {
		return provider.ProbeURL(apiBase, defaultProbe)
	}

	return defaultProbe, nil
}

// ResolveProbeURLFor resolves a probe URL from a live connection or protocol probe.
func ResolveProbeURLFor(
	conn Connection,
	protocol ProtocolType,
	apiBase, defaultProbe string,
) (string, error) {
	return ResolveProbeURL(connectionOrProbe(conn, protocol), apiBase, defaultProbe)
}

func (c *OAuth2Connection) SupportsOrganizationPicker() bool { return true }
func (c *OAuth2Connection) SupportsScopeGrantCheck() bool    { return true }
func (c *OAuth2Connection) SupportsReconnect() bool          { return true }

func (c *APIKeyConnection) SupportsOrganizationPicker() bool { return false }
func (c *APIKeyConnection) SupportsScopeGrantCheck() bool    { return false }
func (c *APIKeyConnection) SupportsReconnect() bool          { return false }

func (c *GitHubAppConnection) SupportsOrganizationPicker() bool { return false }
func (c *GitHubAppConnection) SupportsScopeGrantCheck() bool    { return false }
func (c *GitHubAppConnection) SupportsReconnect() bool          { return true }

func (c *GitHubAppConnection) ProbeURL(apiBase, defaultProbe string) (string, error) {
	if apiBase == "" {
		return "", fmt.Errorf("cannot build install probe URL: missing API base")
	}

	probeURL, err := url.JoinPath(apiBase, "installation/repositories")
	if err != nil {
		return "", fmt.Errorf("cannot build install probe URL: %w", err)
	}

	return probeURL, nil
}

var (
	_ OrganizationSelector = (*OAuth2Connection)(nil)
	_ ScopeGranter         = (*OAuth2Connection)(nil)
	_ Reconnector          = (*OAuth2Connection)(nil)

	_ OrganizationSelector = (*APIKeyConnection)(nil)
	_ ScopeGranter         = (*APIKeyConnection)(nil)
	_ Reconnector          = (*APIKeyConnection)(nil)

	_ OrganizationSelector = (*GitHubAppConnection)(nil)
	_ ScopeGranter         = (*GitHubAppConnection)(nil)
	_ Reconnector          = (*GitHubAppConnection)(nil)
	_ ProbeURLProvider     = (*GitHubAppConnection)(nil)
)
