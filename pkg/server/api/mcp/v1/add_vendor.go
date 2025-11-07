package v1

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

var (
	AddVendorTool = &mcp.Tool{
		Name:         "addVendor",
		Title:        "Add Vendor",
		Description:  "Add a new vendor to the organization",
		Annotations:  &mcp.ToolAnnotations{ReadOnlyHint: false},
		InputSchema:  types.AddVendorInputSchema,
		OutputSchema: types.AddVendorOutputSchema,
	}
)

func (r *resolver) AddVendor(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args types.AddVendorInput,
) (*mcp.CallToolResult, types.AddVendorOutput, error) {
	tenantID := args.OrganizationID.TenantID()
	svc := r.ProboService(ctx, tenantID)

	vendor, err := svc.Vendors.Create(
		ctx,
		probo.CreateVendorRequest{
			OrganizationID:                args.OrganizationID,
			Name:                          args.Name,
			Description:                   args.Description,
			HeadquarterAddress:            args.HeadquarterAddress,
			LegalName:                     args.LegalName,
			WebsiteURL:                    args.WebsiteURL,
			Category:                      args.Category,
			PrivacyPolicyURL:              args.PrivacyPolicyURL,
			ServiceLevelAgreementURL:      args.ServiceLevelAgreementURL,
			DataProcessingAgreementURL:    args.DataProcessingAgreementURL,
			BusinessAssociateAgreementURL: args.BusinessAssociateAgreementURL,
			SubprocessorsListURL:          args.SubprocessorsListURL,
			Certifications:                args.Certifications,
			SecurityPageURL:               args.SecurityPageURL,
			TrustPageURL:                  args.TrustPageURL,
			TermsOfServiceURL:             args.TermsOfServiceURL,
			StatusPageURL:                 args.StatusPageURL,
			BusinessOwnerID:               args.BusinessOwnerID,
			SecurityOwnerID:               args.SecurityOwnerID,
		},
	)
	if err != nil {
		return nil, types.AddVendorOutput{}, fmt.Errorf("failed to create vendor: %w", err)
	}

	return nil, types.NewAddVendorOutput(vendor), nil
}
