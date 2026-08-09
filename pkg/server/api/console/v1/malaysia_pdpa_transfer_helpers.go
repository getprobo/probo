package console_v1

import (
	"context"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func (r *mutationResolver) newMalaysiaPDPATransferRequest(
	ctx context.Context,
	organizationID gid.GID,
	input *types.MalaysiaPDPATransferInput,
) (*probo.MalaysiaPDPATransferRequest, error) {
	if input == nil {
		return nil, nil
	}

	var approverProfileID gid.GID
	if input.ApprovalStatus != coredata.MalaysiaPDPATransferApprovalStatusPending {
		identity := authn.IdentityFromContext(ctx)
		approver, err := r.iam.OrganizationService.GetProfileForIdentityAndOrganization(ctx, identity.ID, organizationID)
		if err != nil {
			r.logger.ErrorCtx(ctx, "cannot get Malaysia PDPA transfer approver profile", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}
		approverProfileID = approver.ID
	}

	return &probo.MalaysiaPDPATransferRequest{
		Basis:                      input.Basis,
		DestinationCountry:         input.DestinationCountry,
		RecipientThirdPartyID:      input.RecipientThirdPartyID,
		ReceiverRegistrationNumber: input.ReceiverRegistrationNumber,
		ReceiverContact:            input.ReceiverContact,
		TransferPurpose:            input.TransferPurpose,
		PersonalDataCategories:     input.PersonalDataCategories,
		Safeguards:                 input.Safeguards,
		ApprovalStatus:             input.ApprovalStatus,
		ApprovalNotes:              input.ApprovalNotes,
		ReviewEvidence:             input.ReviewEvidence,
		ApprovedByProfileID:        approverProfileID,
	}, nil
}
