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
	"time"

	"go.gearno.de/crypto/uuid"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/statelesstoken"
)

const (
	// TokenTypeConnectorInstall namespaces install state against every other
	// stateless token this deployment signs. The provider is inside the signed
	// data rather than in the type, so one shared validator serves every
	// install provider; the callback handler rejects a state whose provider
	// does not match the route it arrived on.
	TokenTypeConnectorInstall = "probo/connector/install"

	// installStateTTL is the house window for a redirect the customer walks
	// through by hand, the same one OAuth2 state and the Slack install use.
	installStateTTL = 10 * time.Minute
)

// InstallState is what an app-install ceremony carries across the vendor. It is
// SIGNED, NOT ENCRYPTED: the vendor and anything logging the URL can read the
// organization and identity GIDs, which is true of every connector state today
// and is the reason nothing secret goes in here.
type InstallState struct {
	Provider       string  `json:"provider"`
	OrganizationID gid.GID `json:"organization_id"`
	// IdentityID BINDS THE TWO LEGS TO ONE HUMAN and IS re-checked on the
	// callback: the session cookie on the callback must belong to this identity
	// or the request is 401'd with nothing claimed. It binds the human, NOT the
	// permission -- the callback separately re-authorizes that identity against
	// OrganizationID, because a role downgrade does not revoke a live session.
	//
	// It is not audit metadata. Server-side verification of the vendor's proof
	// establishes only that the browser came from a real install of that vendor
	// tenant; it says nothing about which Probo organization the tenant belongs
	// in. Without this check, anyone holding ActionConnectorInitiate in any
	// organization can mint a state, hand the vendor's own install link to an
	// administrator of an unrelated tenant, and capture that tenant on a wholly
	// GENUINE vendor token.
	IdentityID gid.GID `json:"identity_id"`
	// Nonce makes each minted state string unique, which is what lets the
	// single-use claim ledger key on its digest. Two tabs are therefore two
	// claims, which is why row dedupe needs its own lock (see
	// coredata.LockConnectorInstallResource).
	Nonce string `json:"nonce"`
}

// NewInstallState mints the signed state an app-install redirect carries to the
// vendor and back.
func NewInstallState(
	secret string,
	provider string,
	organizationID gid.GID,
	identityID gid.GID,
) (string, error) {
	nonce, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("cannot generate connector install state nonce: %w", err)
	}

	token, err := statelesstoken.NewToken(
		secret,
		TokenTypeConnectorInstall,
		installStateTTL,
		InstallState{
			Provider:       provider,
			OrganizationID: organizationID,
			IdentityID:     identityID,
			Nonce:          nonce.String(),
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot create connector install state: %w", err)
	}

	return token, nil
}

// ValidateInstallState checks the signature, the token type and the expiry. It
// proves only that Probo minted this state for someone: the caller must still
// match the identity against the session cookie, re-authorize it against the
// organization, and spend the state exactly once.
func ValidateInstallState(
	secret string,
	state string,
) (*statelesstoken.Payload[InstallState], error) {
	payload, err := statelesstoken.ValidateToken[InstallState](
		secret,
		TokenTypeConnectorInstall,
		state,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot validate connector install state: %w", err)
	}

	return payload, nil
}
