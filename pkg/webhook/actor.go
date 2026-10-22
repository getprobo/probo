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

package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

// Actor is the membership profile that performed the action in a webhook
// delivery.
type Actor struct {
	ID           gid.GID                 `json:"id"`
	FullName     string                  `json:"fullName"`
	EmailAddress mail.Addr               `json:"emailAddress"`
	Role         coredata.MembershipRole `json:"role"`
}

func principalID(scope coredata.Scoper) gid.GID {
	if scope == nil {
		return gid.Nil
	}

	return scope.PrincipalID()
}

func marshalActor(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
) (json.RawMessage, error) {
	identityID := principalID(scope)
	if identityID == gid.Nil {
		return nil, nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByIdentityIDAndOrganizationID(
		ctx,
		tx,
		scope,
		identityID,
		organizationID,
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load actor profile: %w", err)
	}

	membership := &coredata.Membership{}
	if err := membership.LoadByIdentityIDAndOrganizationID(
		ctx,
		tx,
		scope,
		identityID,
		organizationID,
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load actor membership: %w", err)
	}

	raw, err := json.Marshal(
		Actor{
			ID:           profile.ID,
			FullName:     profile.FullName,
			EmailAddress: profile.EmailAddress,
			Role:         membership.Role,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal webhook actor: %w", err)
	}

	return raw, nil
}
