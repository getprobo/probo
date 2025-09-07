package v1

import (
	"context"
	"fmt"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	addVendorArgs struct {
		Name                          string
		Description                   *string
		HeadquarterAddress            *string
		LegalName                     *string
		WebsiteURL                    *string
		Category                      *coredata.VendorCategory
		PrivacyPolicyURL              *string
		ServiceLevelAgreementURL      *string
		DataProcessingAgreementURL    *string
		BusinessAssociateAgreementURL *string
		SubprocessorsListURL          *string
		Certifications                []string
		SecurityPageURL               *string
		TrustPageURL                  *string
		TermsOfServiceURL             *string
		StatusPageURL                 *string
		BusinessOwnerID               *gid.GID
		SecurityOwnerID               *gid.GID
	}

	addVendorResult struct {
		Result struct {
			Name string
			ID   string
		}
	}
)

func (r *resolver) AddVendor(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args *addVendorArgs,
) (*mcp.CallToolResult, *addVendorResult, error) {
	vendor, err := r.proboSvc.Vendors.Create(
		ctx,
		probo.CreateVendorRequest{
			OrganizationID:                r.organizationID,
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
		return nil, nil, fmt.Errorf("failed to list vendors: %w", err)
	}

	result := &addVendorResult{
		Result: struct {
			Name string
			ID   string
		}{
			Name: vendor.Name,
			ID:   vendor.ID.String(),
		},
	}

	return nil, result, nil
}
