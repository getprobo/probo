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

package identityname

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// CopyFromUniqueProfile copies fullName onto the identity when that identity
// has exactly one membership profile and fullName is not empty.
func CopyFromUniqueProfile(
	ctx context.Context,
	tx pg.Tx,
	identityID gid.GID,
	fullName string,
) error {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return nil
	}

	profiles := &coredata.MembershipProfiles{}

	count, err := profiles.CountByIdentityID(
		ctx,
		tx,
		identityID,
		coredata.NewMembershipProfileFilter(nil),
	)
	if err != nil {
		return fmt.Errorf("cannot count membership profiles: %w", err)
	}

	if count != 1 {
		return nil
	}

	identity := &coredata.Identity{}
	if err := identity.LoadByID(ctx, tx, identityID); err != nil {
		return fmt.Errorf("cannot load identity: %w", err)
	}

	if identity.FullName == fullName {
		return nil
	}

	identity.FullName = fullName
	identity.UpdatedAt = time.Now()

	if err := identity.Update(ctx, tx); err != nil {
		return fmt.Errorf("cannot update identity: %w", err)
	}

	return nil
}
