package v1

import (
	"context"
	"fmt"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
)

type (
	addVendorArgs struct {
		OrganizationID                string
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
	// Get MCP context
	mcpCtx := MCPContextFromContext(ctx)
	if mcpCtx == nil {
		r.logger.ErrorCtx(ctx, "AddVendor: missing MCP context")
		return nil, nil, fmt.Errorf("authentication context not found")
	}

	// Parse and validate organization ID
	organizationID, err := gid.ParseGID(args.OrganizationID)
	if err != nil {
		r.logger.WarnCtx(ctx, "AddVendor: invalid organization_id",
			log.Error(err),
			log.String("user_id", mcpCtx.UserID.String()),
			log.String("organization_id", args.OrganizationID),
		)
		return nil, nil, NewValidationError("organizationID", "invalid organization ID format")
	}

	// Validate user has access to the organization
	if err := ValidateOrganizationAccess(ctx, organizationID); err != nil {
		r.logger.WarnCtx(ctx, "AddVendor: access denied",
			log.Error(err),
			log.String("user_id", mcpCtx.UserID.String()),
			log.String("organization_id", organizationID.String()),
		)
		return nil, nil, err
	}

	tenantID := organizationID.TenantID()

	r.logger.InfoCtx(ctx, "AddVendor: creating vendor",
		log.String("tenant_id", tenantID.String()),
		log.String("organization_id", organizationID.String()),
		log.String("user_id", mcpCtx.UserID.String()),
		log.String("vendor_name", args.Name),
	)

	// Note: name is validated by the MCP SDK (required + minLength: 1)

	// Validate URLs if provided
	// Note: While the schema has format: "uri", we validate URLs manually for
	// stricter validation and better error messages
	var validationErrs ValidationErrors
	if args.WebsiteURL != nil {
		if err := ValidateURL("websiteURL", *args.WebsiteURL); err != nil {
			validationErrs = append(validationErrs, err)
		}
	}
	if args.PrivacyPolicyURL != nil {
		if err := ValidateURL("privacyPolicyURL", *args.PrivacyPolicyURL); err != nil {
			validationErrs = append(validationErrs, err)
		}
	}
	if args.ServiceLevelAgreementURL != nil {
		if err := ValidateURL("serviceLevelAgreementURL", *args.ServiceLevelAgreementURL); err != nil {
			validationErrs = append(validationErrs, err)
		}
	}
	if args.DataProcessingAgreementURL != nil {
		if err := ValidateURL("dataProcessingAgreementURL", *args.DataProcessingAgreementURL); err != nil {
			validationErrs = append(validationErrs, err)
		}
	}
	if args.BusinessAssociateAgreementURL != nil {
		if err := ValidateURL("businessAssociateAgreementURL", *args.BusinessAssociateAgreementURL); err != nil {
			validationErrs = append(validationErrs, err)
		}
	}
	if args.SubprocessorsListURL != nil {
		if err := ValidateURL("subprocessorsListURL", *args.SubprocessorsListURL); err != nil {
			validationErrs = append(validationErrs, err)
		}
	}
	if args.SecurityPageURL != nil {
		if err := ValidateURL("securityPageURL", *args.SecurityPageURL); err != nil {
			validationErrs = append(validationErrs, err)
		}
	}
	if args.TrustPageURL != nil {
		if err := ValidateURL("trustPageURL", *args.TrustPageURL); err != nil {
			validationErrs = append(validationErrs, err)
		}
	}
	if args.TermsOfServiceURL != nil {
		if err := ValidateURL("termsOfServiceURL", *args.TermsOfServiceURL); err != nil {
			validationErrs = append(validationErrs, err)
		}
	}
	if args.StatusPageURL != nil {
		if err := ValidateURL("statusPageURL", *args.StatusPageURL); err != nil {
			validationErrs = append(validationErrs, err)
		}
	}

	if validationErrs.HasErrors() {
		r.logger.WarnCtx(ctx, "AddVendor: validation failed",
			log.Error(validationErrs),
			log.String("tenant_id", tenantID.String()),
		)
		return nil, nil, validationErrs
	}

	// Get tenant-scoped service
	svc := r.proboSvc.WithTenant(tenantID)

	vendor, err := svc.Vendors.Create(
		ctx,
		probo.CreateVendorRequest{
			OrganizationID:                organizationID,
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
		r.logger.ErrorCtx(ctx, "AddVendor: failed to create vendor",
			log.Error(err),
			log.String("tenant_id", tenantID.String()),
			log.String("organization_id", organizationID.String()),
			log.String("vendor_name", args.Name),
		)
		return nil, nil, fmt.Errorf("failed to create vendor: %w", err)
	}

	r.logger.InfoCtx(ctx, "AddVendor: vendor created successfully",
		log.String("tenant_id", tenantID.String()),
		log.String("organization_id", organizationID.String()),
		log.String("vendor_id", vendor.ID.String()),
		log.String("vendor_name", vendor.Name),
	)

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
