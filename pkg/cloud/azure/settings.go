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

package azure

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	errInvalidTenantID       = errors.New("tenantId is not a GUID")
	errInvalidClientID       = errors.New("clientId is not a GUID")
	errInvalidSubscriptionID = errors.New("subscriptionId is not a GUID")
	errInvalidEnvironment    = errors.New("environment is not a supported Azure environment")
)

// ConnectorSettings is the customer-supplied Entra application, subscription,
// and cloud. Every value is public knowledge.
type ConnectorSettings struct {
	TenantID       string
	ClientID       string
	SubscriptionID string
	Environment    Environment
}

// NewConnectorSettings validates the customer-supplied tenant, client, and
// subscription identifiers. Values are stored trimmed and lowercased. Returned
// errors are safe to show a client: they never echo the identifier.
func NewConnectorSettings(
	tenantID, clientID, subscriptionID, environment string,
) (ConnectorSettings, error) {
	tenantID, err := parseGUID(tenantID, errInvalidTenantID)
	if err != nil {
		return ConnectorSettings{}, fmt.Errorf("cannot create azure connector: %w", err)
	}

	clientID, err = parseGUID(clientID, errInvalidClientID)
	if err != nil {
		return ConnectorSettings{}, fmt.Errorf("cannot create azure connector: %w", err)
	}

	subscriptionID, err = parseGUID(subscriptionID, errInvalidSubscriptionID)
	if err != nil {
		return ConnectorSettings{}, fmt.Errorf("cannot create azure connector: %w", err)
	}

	env, err := parseEnvironment(environment)
	if err != nil {
		return ConnectorSettings{}, fmt.Errorf("cannot create azure connector: %w", err)
	}

	return ConnectorSettings{
		TenantID:       tenantID,
		ClientID:       clientID,
		SubscriptionID: subscriptionID,
		Environment:    env,
	}, nil
}

func parseGUID(raw string, invalid error) (string, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == uuid.Nil {
		return "", invalid
	}

	return parsed.String(), nil
}

func parseEnvironment(raw string) (Environment, error) {
	var env Environment
	if err := env.UnmarshalText([]byte(strings.TrimSpace(raw))); err != nil {
		return "", err
	}

	return env, nil
}
