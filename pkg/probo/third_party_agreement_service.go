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

package probo

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	ThirdPartyAgreementService struct {
		svc           *Service
		fileValidator *filevalidation.FileValidator
	}

	ThirdPartyAgreementCreateRequest struct {
		File       FileUpload
		Type       coredata.ThirdPartyAgreementType
		ValidFrom  *time.Time
		ValidUntil *time.Time
	}

	ThirdPartyAgreementUpdateRequest struct {
		Type       *coredata.ThirdPartyAgreementType
		ValidFrom  **time.Time
		ValidUntil **time.Time
	}
)

func (r *ThirdPartyAgreementCreateRequest) Validate() error {
	v := validator.New()

	v.Check(r.Type, "type", validThirdPartyAgreementType())
	v.Check(r.ValidUntil, "valid_until", validator.After(r.ValidFrom))

	return v.Error()
}

func (r *ThirdPartyAgreementUpdateRequest) Validate() error {
	v := validator.New()

	if r.Type != nil {
		v.Check(*r.Type, "type", validThirdPartyAgreementType())
	}

	return v.Error()
}

func validThirdPartyAgreementType() validator.ValidatorFunc {
	return func(value any) *validator.ValidationError {
		agreementType, ok := value.(coredata.ThirdPartyAgreementType)
		if !ok || agreementType.IsValid() {
			return nil
		}

		return &validator.ValidationError{
			Code:    validator.ErrorCodeInvalidEnum,
			Message: "must be a known agreement type",
		}
	}
}

func (s ThirdPartyAgreementService) ListForThirdPartyID(
	ctx context.Context,
	scope coredata.Scoper,
	thirdPartyID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyAgreementOrderField],
) (*page.Page[*coredata.ThirdPartyAgreement, coredata.ThirdPartyAgreementOrderField], error) {
	var agreements coredata.ThirdPartyAgreements

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return agreements.LoadForThirdPartyID(ctx, conn, scope, thirdPartyID, cursor)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(agreements, cursor), nil
}

func (s ThirdPartyAgreementService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	agreementID gid.GID,
) (*coredata.ThirdPartyAgreement, error) {
	agreement := &coredata.ThirdPartyAgreement{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return agreement.LoadByID(ctx, conn, scope, agreementID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load third party agreement: %w", err)
	}

	return agreement, nil
}

func (s ThirdPartyAgreementService) Upload(
	ctx context.Context,
	scope coredata.Scoper,
	thirdPartyID gid.GID,
	req *ThirdPartyAgreementCreateRequest,
) (*coredata.ThirdPartyAgreement, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	thirdParty, err := s.svc.ThirdParties.Get(ctx, scope, thirdPartyID)
	if err != nil {
		return nil, fmt.Errorf("cannot get third party: %w", err)
	}

	file, err := s.svc.Files.UploadAndSaveFile(
		ctx,
		scope,
		s.fileValidator,
		map[string]string{
			"type":            "third-party-agreement",
			"third-party-id":  thirdPartyID.String(),
			"organization-id": thirdParty.OrganizationID.String(),
		},
		&req.File,
	)
	if err != nil {
		if validationErrors, ok := fileUploadValidationErrors(err); ok {
			return nil, validationErrors
		}

		return nil, err
	}

	now := time.Now()
	agreement := &coredata.ThirdPartyAgreement{
		ID:             gid.New(scope.GetTenantID(), coredata.ThirdPartyAgreementEntityType),
		OrganizationID: thirdParty.OrganizationID,
		ThirdPartyID:   thirdPartyID,
		Type:           req.Type,
		ValidFrom:      req.ValidFrom,
		ValidUntil:     req.ValidUntil,
		FileID:         file.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return agreement.Insert(ctx, tx, scope)
		},
	)
	if err != nil {
		return nil, err
	}

	return agreement, nil
}

func (s ThirdPartyAgreementService) Update(
	ctx context.Context,
	scope coredata.Scoper,
	agreementID gid.GID,
	req *ThirdPartyAgreementUpdateRequest,
) (*coredata.ThirdPartyAgreement, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	agreement := &coredata.ThirdPartyAgreement{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := agreement.LoadByID(ctx, tx, scope, agreementID); err != nil {
				return fmt.Errorf("cannot load third party agreement: %w", err)
			}

			if req.Type != nil {
				agreement.Type = *req.Type
			}

			if req.ValidFrom != nil {
				agreement.ValidFrom = *req.ValidFrom
			}

			if req.ValidUntil != nil {
				agreement.ValidUntil = *req.ValidUntil
			}

			v := validator.New()
			v.Check(agreement.ValidUntil, "valid_until", validator.After(agreement.ValidFrom))

			if err := v.Error(); err != nil {
				return err
			}

			agreement.UpdatedAt = time.Now()

			if err := agreement.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update third party agreement: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return agreement, nil
}

func (s ThirdPartyAgreementService) Delete(
	ctx context.Context,
	scope coredata.Scoper,
	agreementID gid.GID,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			agreement := &coredata.ThirdPartyAgreement{}
			if err := agreement.LoadByID(ctx, tx, scope, agreementID); err != nil {
				return fmt.Errorf("cannot load third party agreement: %w", err)
			}

			if err := agreement.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete third party agreement: %w", err)
			}

			return nil
		},
	)
}
