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

package mcp_v1

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/validator"
)

func (r *Resolver) malaysiaPDPABreachActor(ctx context.Context, organizationID gid.GID) (*coredata.MembershipProfile, error) {
	identity := authn.IdentityFromContext(ctx)
	actor, err := r.iamSvc.OrganizationService.GetProfileForIdentityAndOrganization(ctx, identity.ID, organizationID)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot load Malaysia PDPA breach actor profile", log.Error(err))
		return nil, fmt.Errorf("internal server error")
	}

	return actor, nil
}

func (r *Resolver) malaysiaPDPABreachMutationError(ctx context.Context, message string, err error) error {
	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		return validationErrors
	}

	r.logger.ErrorCtx(ctx, message, log.Error(err))
	return fmt.Errorf("internal server error")
}

func intPointerFromMCP(value *int) *int64 {
	if value == nil {
		return nil
	}

	converted := int64(*value)
	return &converted
}
