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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/malaysiapdpa"
	"go.probo.inc/probo/pkg/validator"
)

type (
	UpdateMalaysiaPDPAProfileRequest struct {
		OrganizationID                    gid.GID
		TotalDataSubjects                 int64
		SensitiveDataSubjects             int64
		RegularSystematicMonitoring       bool
		AssessedByProfileID               gid.GID
		DPOProfileID                      *gid.GID
		DPOAppointedAt                    *time.Time
		CommissionerNotifiedAt            *time.Time
		CommissionerNotificationReference *string
	}
)

func (req *UpdateMalaysiaPDPAProfileRequest) Validate() error {
	v := validator.New()

	v.Check(req.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(req.TotalDataSubjects, "total_data_subjects", validator.Min(0))
	v.Check(req.SensitiveDataSubjects, "sensitive_data_subjects", validator.Min(0))
	v.Check(req.AssessedByProfileID, "assessed_by_profile_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
	v.Check(req.DPOProfileID, "dpo_profile_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(req.CommissionerNotificationReference, "commissioner_notification_reference", validator.SafeTextNoNewLine(TitleMaxLength))

	if req.SensitiveDataSubjects > req.TotalDataSubjects {
		v.Check(
			req.SensitiveDataSubjects,
			"sensitive_data_subjects",
			func(any) *validator.ValidationError {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeCustom,
					Message: "must not exceed total_data_subjects",
				}
			},
		)
	}

	if (req.DPOProfileID == nil) != (req.DPOAppointedAt == nil) {
		v.Check(
			req.DPOProfileID,
			"dpo_profile_id",
			func(any) *validator.ValidationError {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeCustom,
					Message: "and dpo_appointed_at must either both be set or both be omitted",
				}
			},
		)
	}

	if req.CommissionerNotifiedAt != nil && req.DPOAppointedAt == nil {
		v.Check(
			req.CommissionerNotifiedAt,
			"commissioner_notified_at",
			func(any) *validator.ValidationError {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeCustom,
					Message: "requires an appointed DPO",
				}
			},
		)
	}

	if req.CommissionerNotifiedAt != nil && req.DPOAppointedAt != nil && req.CommissionerNotifiedAt.Before(*req.DPOAppointedAt) {
		v.Check(
			req.CommissionerNotifiedAt,
			"commissioner_notified_at",
			func(any) *validator.ValidationError {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeCustom,
					Message: "must not be before dpo_appointed_at",
				}
			},
		)
	}

	if req.CommissionerNotificationReference != nil && req.CommissionerNotifiedAt == nil {
		v.Check(
			req.CommissionerNotificationReference,
			"commissioner_notification_reference",
			func(any) *validator.ValidationError {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeCustom,
					Message: "requires commissioner_notified_at",
				}
			},
		)
	}

	return v.Error()
}

func (s OrganizationService) GetMalaysiaPDPAProfile(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.MalaysiaPDPAProfile, error) {
	profile := &coredata.MalaysiaPDPAProfile{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			err := profile.LoadByOrganizationID(ctx, conn, scope, organizationID)
			if errors.Is(err, coredata.ErrResourceNotFound) {
				profile.OrganizationID = organizationID
				profile.DPORequirementReasons = []string{}

				return nil
			}
			if err != nil {
				return fmt.Errorf("cannot load Malaysia PDPA profile: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s OrganizationService) UpdateMalaysiaPDPAProfile(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateMalaysiaPDPAProfileRequest,
) (*coredata.MalaysiaPDPAProfile, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	assessment, err := malaysiapdpa.AssessDPORequirement(
		malaysiapdpa.DPOAssessmentInput{
			TotalDataSubjects:           req.TotalDataSubjects,
			SensitiveDataSubjects:       req.SensitiveDataSubjects,
			RegularSystematicMonitoring: req.RegularSystematicMonitoring,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot assess DPO requirement: %w", err)
	}

	now := time.Now()
	profile := &coredata.MalaysiaPDPAProfile{
		OrganizationID:                    req.OrganizationID,
		TotalDataSubjects:                 req.TotalDataSubjects,
		SensitiveDataSubjects:             req.SensitiveDataSubjects,
		RegularSystematicMonitoring:       req.RegularSystematicMonitoring,
		DPORequired:                       assessment.Required,
		DPORequirementReasons:             dpoRequirementReasonStrings(assessment.Reasons),
		AssessedByProfileID:               &req.AssessedByProfileID,
		AssessedAt:                        &now,
		DPOProfileID:                      req.DPOProfileID,
		DPOAppointedAt:                    req.DPOAppointedAt,
		CommissionerNotifiedAt:            req.CommissionerNotifiedAt,
		CommissionerNotificationReference: req.CommissionerNotificationReference,
		CreatedAt:                         now,
		UpdatedAt:                         now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := validateProfileOrganization(
				ctx,
				tx,
				scope,
				req.AssessedByProfileID,
				req.OrganizationID,
				"assessor",
			); err != nil {
				return err
			}

			if req.DPOProfileID != nil {
				if err := validateProfileOrganization(
					ctx,
					tx,
					scope,
					*req.DPOProfileID,
					req.OrganizationID,
					"DPO",
				); err != nil {
					return err
				}
			}

			existing := &coredata.MalaysiaPDPAProfile{}
			err := existing.LoadByOrganizationID(ctx, tx, scope, req.OrganizationID)
			if err == nil {
				profile.CreatedAt = existing.CreatedAt
			} else if !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load existing Malaysia PDPA profile: %w", err)
			}

			if err := profile.Upsert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot save Malaysia PDPA profile: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func dpoRequirementReasonStrings(reasons []malaysiapdpa.DPORequirementReason) []string {
	values := make([]string, len(reasons))
	for i, reason := range reasons {
		values[i] = string(reason)
	}

	return values
}

func validateProfileOrganization(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	profileID gid.GID,
	organizationID gid.GID,
	role string,
) error {
	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByID(ctx, conn, scope, profileID); err != nil {
		return fmt.Errorf("cannot load %s profile: %w", role, err)
	}

	if profile.OrganizationID != organizationID {
		return fmt.Errorf("%s profile does not belong to the organization", role)
	}

	return nil
}
