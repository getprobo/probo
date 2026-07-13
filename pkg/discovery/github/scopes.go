// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package github

import (
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

// RequiredDiscoveryScopes are the minimum OAuth2 scopes needed to start a
// discovery run on the existing GitHub connector. Additional discovery
// scopes are requested through reconnect when deeper checks need them.
func RequiredDiscoveryScopes() []string {
	return []string{"read:org"}
}

// EscalationDiscoveryScopes are optional scopes requested on reconnect to
// unlock deeper discovery checks. The scanner degrades gracefully when
// they are absent.
func EscalationDiscoveryScopes() []string {
	return []string{
		"repo",
		"security_events",
		"read:enterprise",
		"read:audit_log",
	}
}

// DiscoveryReconnectScopes returns the full scope set to pass to the
// connector reconnect flow for GitHub discovery escalation.
func DiscoveryReconnectScopes() []string {
	return connector.UnionScopes(RequiredDiscoveryScopes(), EscalationDiscoveryScopes())
}

// InsufficientScopesError is returned when a GitHub connector lacks the
// scopes required to start discovery.
type InsufficientScopesError struct {
	Missing []string
}

func (e *InsufficientScopesError) Error() string {
	return fmt.Sprintf(
		"github connector is missing required oauth scopes: %s",
		strings.Join(e.Missing, ", "),
	)
}

// MissingRequiredScopes returns the required discovery scopes that are
// not present in granted.
func MissingRequiredScopes(granted []string) []string {
	return missingScopes(granted, RequiredDiscoveryScopes())
}

func missingScopes(granted, required []string) []string {
	present := map[string]struct{}{}

	for _, scope := range granted {
		present[scope] = struct{}{}
	}

	missing := make([]string, 0, len(required))

	for _, scope := range required {
		if _, ok := present[scope]; ok {
			continue
		}

		missing = append(missing, scope)
	}

	return missing
}

func validateDiscoveryScopes(protocol coredata.ConnectorProtocol, conn connector.Connection) error {
	if protocol != coredata.ConnectorProtocolOAuth2 || conn == nil {
		return nil
	}

	if missing := MissingRequiredScopes(conn.Scopes()); len(missing) > 0 {
		return &InsufficientScopesError{Missing: missing}
	}

	return nil
}
