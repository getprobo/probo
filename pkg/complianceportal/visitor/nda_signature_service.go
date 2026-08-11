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

package visitor

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/gid"
)

func (s *Service) AcceptPortalNDASignature(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	identityID gid.GID,
	req *esign.AcceptSignatureRequest,
) (*coredata.ElectronicSignature, error) {
	if err := s.authorizePortalNDASignature(ctx, scope, compliancePortalID, identityID, req.SignatureID); err != nil {
		return nil, err
	}

	if s.esign == nil {
		return nil, fmt.Errorf("electronic signature service is not configured")
	}

	return s.esign.AcceptSignature(ctx, scope, req)
}

func (s *Service) RecordPortalNDASigningEvent(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	identityID gid.GID,
	req *esign.RecordEventRequest,
) error {
	if err := s.authorizePortalNDASignature(ctx, scope, compliancePortalID, identityID, req.SignatureID); err != nil {
		return err
	}

	if s.esign == nil {
		return fmt.Errorf("electronic signature service is not configured")
	}

	return s.esign.RecordEvent(ctx, scope, req)
}

func (s *Service) authorizePortalNDASignature(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	identityID gid.GID,
	signatureID gid.GID,
) error {
	access := &coredata.CompliancePortalAccess{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return access.LoadByCompliancePortalIDAndIdentityID(
				ctx,
				conn,
				scope,
				compliancePortalID,
				identityID,
			)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return esign.ErrSignatureAccessDenied
		}

		return fmt.Errorf("cannot load compliance portal access: %w", err)
	}

	if access.ElectronicSignatureID == nil || *access.ElectronicSignatureID != signatureID {
		return esign.ErrSignatureAccessDenied
	}

	return nil
}
