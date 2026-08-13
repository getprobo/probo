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

package slack

import (
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/statelesstoken"
)

const (
	TokenTypeSlackbotInstall = "slackbot_install"

	installStateExpiry = 10 * time.Minute
)

type InstallStateData struct {
	OrganizationID gid.GID `json:"organization_id"`
	IdentityID     gid.GID `json:"identity_id"`
	Nonce          string  `json:"nonce"`
	ContinueURL    string  `json:"continue_url,omitempty"`
}

func newInstallState(
	secret string,
	organizationID gid.GID,
	identityID gid.GID,
	continueURL string,
) (string, error) {
	nonce, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("cannot generate Slack install state nonce: %w", err)
	}

	token, err := statelesstoken.NewToken(
		secret,
		TokenTypeSlackbotInstall,
		installStateExpiry,
		InstallStateData{
			OrganizationID: organizationID,
			IdentityID:     identityID,
			Nonce:          nonce.String(),
			ContinueURL:    continueURL,
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot create Slack install state: %w", err)
	}

	return token, nil
}

func validateInstallState(
	secret string,
	state string,
) (*statelesstoken.Payload[InstallStateData], error) {
	payload, err := statelesstoken.ValidateToken[InstallStateData](
		secret,
		TokenTypeSlackbotInstall,
		state,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot validate Slack install state: %w", err)
	}

	return payload, nil
}
