// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/malaysiapdpa"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type TransferImpactAssessmentService struct {
	svc *Service
}

type (
	MalaysiaPDPATransferRequest struct {
		Basis                      coredata.MalaysiaPDPATransferBasis
		DestinationCountry         coredata.CountryCode
		RecipientThirdPartyID      gid.GID
		ReceiverRegistrationNumber *string
		ReceiverContact            string
		TransferPurpose            string
		PersonalDataCategories     string
		Safeguards                 string
		ApprovalStatus             coredata.MalaysiaPDPATransferApprovalStatus
		ApprovalNotes              *string
		ReviewEvidence             *string
		ApprovedByProfileID        gid.GID
	}

	CreateTransferImpactAssessmentRequest struct {
		ProcessingActivityID  gid.GID
		DataSubjects          *string
		LegalMechanism        *string
		Transfer              *string
		LocalLawRisk          *string
		SupplementaryMeasures *string
		MalaysiaPDPA          *MalaysiaPDPATransferRequest
	}

	UpdateTransferImpactAssessmentRequest struct {
		ID                    gid.GID
		DataSubjects          **string
		LegalMechanism        **string
		Transfer              **string
		LocalLawRisk          **string
		SupplementaryMeasures **string
		MalaysiaPDPA          *MalaysiaPDPATransferRequest
	}
)

func (req *CreateTransferImpactAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(req.ProcessingActivityID, "processing_activity_id", validator.Required(), validator.GID(coredata.ProcessingActivityEntityType))
	v.Check(req.DataSubjects, "data_subjects", validator.SafeText(ContentMaxLength))
	v.Check(req.LegalMechanism, "legal_mechanism", validator.SafeText(ContentMaxLength))
	v.Check(req.Transfer, "transfer", validator.SafeText(ContentMaxLength))
	v.Check(req.LocalLawRisk, "local_law_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.SupplementaryMeasures, "supplementary_measures", validator.SafeText(ContentMaxLength))
	validateMalaysiaPDPATransferRequest(v, req.MalaysiaPDPA)

	return v.Error()
}

func (req *UpdateTransferImpactAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.TransferImpactAssessmentEntityType))
	v.Check(req.DataSubjects, "data_subjects", validator.SafeText(ContentMaxLength))
	v.Check(req.LegalMechanism, "legal_mechanism", validator.SafeText(ContentMaxLength))
	v.Check(req.Transfer, "transfer", validator.SafeText(ContentMaxLength))
	v.Check(req.LocalLawRisk, "local_law_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.SupplementaryMeasures, "supplementary_measures", validator.SafeText(ContentMaxLength))
	validateMalaysiaPDPATransferRequest(v, req.MalaysiaPDPA)

	return v.Error()
}

func (s TransferImpactAssessmentService) Get(
	ctx context.Context, scope coredata.Scoper,
	tiaID gid.GID,
) (*coredata.TransferImpactAssessment, error) {
	tia := &coredata.TransferImpactAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tia.LoadByID(ctx, conn, scope, tiaID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s TransferImpactAssessmentService) GetByProcessingActivityID(
	ctx context.Context, scope coredata.Scoper,
	processingActivityID gid.GID,
) (*coredata.TransferImpactAssessment, error) {
	tia := &coredata.TransferImpactAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tia.LoadByProcessingActivityID(ctx, conn, scope, processingActivityID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s TransferImpactAssessmentService) ListForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.TransferImpactAssessmentOrderField],
) (*page.Page[*coredata.TransferImpactAssessment, coredata.TransferImpactAssessmentOrderField], error) {
	var tias coredata.TransferImpactAssessments

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := tias.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load transfer impact assessments: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(tias, cursor), nil
}

func (s TransferImpactAssessmentService) CountForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			tias := coredata.TransferImpactAssessments{}
			count, err = tias.CountByOrganizationID(ctx, conn, scope, organizationID)

			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *TransferImpactAssessmentService) Create(
	ctx context.Context, scope coredata.Scoper,
	req *CreateTransferImpactAssessmentRequest,
) (*coredata.TransferImpactAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	tia := &coredata.TransferImpactAssessment{
		ID:                    gid.New(scope.GetTenantID(), coredata.TransferImpactAssessmentEntityType),
		ProcessingActivityID:  req.ProcessingActivityID,
		DataSubjects:          req.DataSubjects,
		LegalMechanism:        req.LegalMechanism,
		Transfer:              req.Transfer,
		LocalLawRisk:          req.LocalLawRisk,
		SupplementaryMeasures: req.SupplementaryMeasures,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			processingActivity := &coredata.ProcessingActivity{}
			if err := processingActivity.LoadByID(ctx, conn, scope, req.ProcessingActivityID); err != nil {
				return fmt.Errorf("cannot load processing activity: %w", err)
			}

			tia.OrganizationID = processingActivity.OrganizationID

			if req.MalaysiaPDPA != nil {
				if err := applyMalaysiaPDPATransfer(ctx, conn, scope, tia, processingActivity, req.MalaysiaPDPA); err != nil {
					return err
				}
			}

			if err := tia.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert transfer impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s *TransferImpactAssessmentService) Update(
	ctx context.Context, scope coredata.Scoper,
	req *UpdateTransferImpactAssessmentRequest,
) (*coredata.TransferImpactAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	tia := &coredata.TransferImpactAssessment{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := tia.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			if req.DataSubjects != nil {
				tia.DataSubjects = *req.DataSubjects
			}

			if req.LegalMechanism != nil {
				tia.LegalMechanism = *req.LegalMechanism
			}

			if req.Transfer != nil {
				tia.Transfer = *req.Transfer
			}

			if req.LocalLawRisk != nil {
				tia.LocalLawRisk = *req.LocalLawRisk
			}

			if req.SupplementaryMeasures != nil {
				tia.SupplementaryMeasures = *req.SupplementaryMeasures
			}

			if req.MalaysiaPDPA != nil {
				processingActivity := &coredata.ProcessingActivity{}
				if err := processingActivity.LoadByID(ctx, conn, scope, tia.ProcessingActivityID); err != nil {
					return fmt.Errorf("cannot load processing activity for Malaysia PDPA transfer: %w", err)
				}

				if err := applyMalaysiaPDPATransfer(ctx, conn, scope, tia, processingActivity, req.MalaysiaPDPA); err != nil {
					return err
				}
			}

			tia.UpdatedAt = time.Now()

			if err := tia.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update transfer impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tia, nil
}

func validateMalaysiaPDPATransferRequest(v *validator.Validator, req *MalaysiaPDPATransferRequest) {
	if req == nil {
		return
	}

	v.Check(req.Basis, "malaysia_pdpa.basis", validator.Required(), validator.OneOfSlice(coredata.MalaysiaPDPATransferBases()))
	v.Check(req.RecipientThirdPartyID, "malaysia_pdpa.recipient_third_party_id", validator.Required(), validator.GID(coredata.ThirdPartyEntityType))
	if !isBlank(req.ReceiverRegistrationNumber) {
		v.Check(req.ReceiverRegistrationNumber, "malaysia_pdpa.receiver_registration_number", validator.SafeTextNoNewLine(TitleMaxLength))
	}
	v.Check(req.ReceiverContact, "malaysia_pdpa.receiver_contact", validator.Required(), validator.SafeText(ContentMaxLength))
	v.Check(req.TransferPurpose, "malaysia_pdpa.transfer_purpose", validator.Required(), validator.SafeText(ContentMaxLength))
	v.Check(req.PersonalDataCategories, "malaysia_pdpa.personal_data_categories", validator.Required(), validator.SafeText(ContentMaxLength))
	v.Check(req.Safeguards, "malaysia_pdpa.safeguards", validator.Required(), validator.SafeText(ContentMaxLength))
	v.Check(req.ApprovalStatus, "malaysia_pdpa.approval_status", validator.Required(), validator.OneOfSlice(coredata.MalaysiaPDPATransferApprovalStatuses()))
	if !isBlank(req.ApprovalNotes) {
		v.Check(req.ApprovalNotes, "malaysia_pdpa.approval_notes", validator.SafeText(ContentMaxLength))
	}
	if !isBlank(req.ReviewEvidence) {
		v.Check(req.ReviewEvidence, "malaysia_pdpa.review_evidence", validator.SafeText(ContentMaxLength))
	}

	if !req.DestinationCountry.IsValid() || req.DestinationCountry == coredata.CountryCodeMY || req.DestinationCountry == coredata.CountryCodeGlobal {
		v.Check(req.DestinationCountry, "malaysia_pdpa.destination_country", func(any) *validator.ValidationError {
			return &validator.ValidationError{Code: validator.ErrorCodeInvalidEnum, Message: "must be a foreign country"}
		})
	}

	if req.ApprovalStatus == coredata.MalaysiaPDPATransferApprovalStatusApproved {
		v.Check(req.ApprovedByProfileID, "malaysia_pdpa.approved_by_profile_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
		if isBlank(req.ReviewEvidence) {
			v.Check(req.ReviewEvidence, "malaysia_pdpa.review_evidence", func(any) *validator.ValidationError {
				return &validator.ValidationError{Code: validator.ErrorCodeRequired, Message: "is required for approval"}
			})
		}
	}

	if req.ApprovalStatus == coredata.MalaysiaPDPATransferApprovalStatusRejected {
		v.Check(req.ApprovedByProfileID, "malaysia_pdpa.approved_by_profile_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
		if isBlank(req.ApprovalNotes) {
			v.Check(req.ApprovalNotes, "malaysia_pdpa.approval_notes", func(any) *validator.ValidationError {
				return &validator.ValidationError{Code: validator.ErrorCodeRequired, Message: "is required for rejection"}
			})
		}
	}
}

func applyMalaysiaPDPATransfer(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	tia *coredata.TransferImpactAssessment,
	processingActivity *coredata.ProcessingActivity,
	req *MalaysiaPDPATransferRequest,
) error {
	if !processingActivity.InternationalTransfers {
		return validator.ValidationErrors{
			&validator.ValidationError{
				Field:   "malaysia_pdpa",
				Code:    validator.ErrorCodeCustom,
				Message: "requires processing activity international_transfers to be true",
				Value:   req,
			},
		}
	}

	recipient := &coredata.ThirdParty{}
	if err := recipient.LoadByID(ctx, conn, scope, req.RecipientThirdPartyID); err != nil {
		return fmt.Errorf("cannot load Malaysia PDPA transfer recipient: %w", err)
	}
	if recipient.OrganizationID != tia.OrganizationID {
		return fmt.Errorf("Malaysia PDPA transfer recipient does not belong to the organization")
	}

	var approvedByProfileID *gid.GID
	var reviewedAt *time.Time
	var nextReviewAt *time.Time
	if req.ApprovalStatus != coredata.MalaysiaPDPATransferApprovalStatusPending {
		if err := validateProfileOrganization(ctx, conn, scope, req.ApprovedByProfileID, tia.OrganizationID, "transfer approver"); err != nil {
			return err
		}

		now := time.Now()
		approvedByProfileID = &req.ApprovedByProfileID
		reviewedAt = &now
		if req.ApprovalStatus == coredata.MalaysiaPDPATransferApprovalStatusApproved {
			nextReviewAt = malaysiapdpa.TransferNextReviewAt(req.Basis, now)
		}
	}

	basis := req.Basis
	destination := req.DestinationCountry
	recipientID := req.RecipientThirdPartyID
	receiverContact := req.ReceiverContact
	transferPurpose := req.TransferPurpose
	personalDataCategories := req.PersonalDataCategories
	safeguards := req.Safeguards
	approvalStatus := req.ApprovalStatus
	ruleVersion := malaysiapdpa.TransferRuleVersion
	ruleSource := malaysiapdpa.TransferRuleSource
	tia.MalaysiaTransferBasis = &basis
	tia.MalaysiaDestinationCountry = &destination
	tia.MalaysiaRecipientThirdPartyID = &recipientID
	tia.MalaysiaReceiverRegistrationNumber = nilIfBlank(req.ReceiverRegistrationNumber)
	tia.MalaysiaReceiverContact = &receiverContact
	tia.MalaysiaTransferPurpose = &transferPurpose
	tia.MalaysiaPersonalDataCategories = &personalDataCategories
	tia.MalaysiaSafeguards = &safeguards
	tia.MalaysiaApprovalStatus = &approvalStatus
	tia.MalaysiaApprovedByProfileID = approvedByProfileID
	tia.MalaysiaApprovalNotes = nilIfBlank(req.ApprovalNotes)
	tia.MalaysiaReviewedAt = reviewedAt
	tia.MalaysiaNextReviewAt = nextReviewAt
	tia.MalaysiaReviewEvidence = nilIfBlank(req.ReviewEvidence)
	tia.MalaysiaRuleVersion = &ruleVersion
	tia.MalaysiaRuleSource = &ruleSource

	return nil
}

func nilIfBlank(value *string) *string {
	if isBlank(value) {
		return nil
	}

	return value
}

func (s *TransferImpactAssessmentService) Delete(
	ctx context.Context, scope coredata.Scoper,
	tiaID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			tia := &coredata.TransferImpactAssessment{}
			if err := tia.LoadByID(ctx, conn, scope, tiaID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			if err := tia.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete transfer impact assessment: %w", err)
			}

			return nil
		},
	)

	return err
}
