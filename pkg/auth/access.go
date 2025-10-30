package auth

import (
	"fmt"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
)

// AuthMethod represents a method of authentication
type AuthMethod int

const (
	AuthMethodPassword AuthMethod = iota
	AuthMethodSAML
	AuthMethodAny // Used when either password or SAML would work
)

// OrgAuthRequirement encapsulates the authentication requirements for accessing an organization
type OrgAuthRequirement struct {
	OrganizationID gid.GID
	EmailDomain    string
	SAMLConfig     *coredata.SAMLConfiguration // nil if no SAML config applies to this org+domain
}

// AccessResult represents the result of an organization access check
type AccessResult struct {
	OrganizationID gid.GID
	Allowed        bool
	MissingAuth    AuthMethod                  // Which auth method is missing (if not allowed)
	SAMLConfig     *coredata.SAMLConfiguration // The SAML config involved (if any)
}

// Check performs the access control decision based on session state
// This is pure business logic with no side effects - easily testable
func (r OrgAuthRequirement) Check(session coredata.SessionData) AccessResult {
	// No SAML config or disabled → requires password authentication
	if r.SAMLConfig == nil || !r.SAMLConfig.Enabled || !r.SAMLConfig.DomainVerified {
		return AccessResult{
			OrganizationID: r.OrganizationID,
			Allowed:        session.PasswordAuthenticated,
			MissingAuth:    AuthMethodPassword,
			SAMLConfig:     nil,
		}
	}

	// Check if user has SAML authentication for this specific organization
	orgKey := r.OrganizationID.String()
	_, hasSAML := session.SAMLAuthenticatedOrgs[orgKey]

	// SAML enforcement: REQUIRED → must have SAML auth for this org
	if r.SAMLConfig.EnforcementPolicy == coredata.SAMLEnforcementPolicyRequired {
		return AccessResult{
			OrganizationID: r.OrganizationID,
			Allowed:        hasSAML,
			MissingAuth:    AuthMethodSAML,
			SAMLConfig:     r.SAMLConfig,
		}
	}

	// SAML enforcement: OPTIONAL → needs either password (global) OR SAML (for this org)
	hasAnyAuth := session.PasswordAuthenticated || hasSAML
	missingAuth := AuthMethodAny
	if hasAnyAuth {
		missingAuth = AuthMethodPassword // Not actually missing, but need a value
	}

	return AccessResult{
		OrganizationID: r.OrganizationID,
		Allowed:        hasAnyAuth,
		MissingAuth:    missingAuth,
		SAMLConfig:     r.SAMLConfig,
	}
}

// ToError converts an AccessResult to an error if access is denied
// This handles the presentation layer concern of generating appropriate errors and redirect URLs
func (r AccessResult) ToError(baseURL string) error {
	if r.Allowed {
		return nil
	}

	switch r.MissingAuth {
	case AuthMethodPassword:
		return ErrPasswordAuthRequired{
			OrganizationID: r.OrganizationID,
			RedirectURL:    fmt.Sprintf("%s/authentication/login?method=password", baseURL),
		}
	case AuthMethodSAML, AuthMethodAny:
		return ErrSAMLAuthRequired{
			ConfigID:       r.SAMLConfig.ID,
			OrganizationID: r.OrganizationID,
			RedirectURL:    fmt.Sprintf("%s/auth/saml/login/%s", baseURL, r.SAMLConfig.ID),
		}
	default:
		return fmt.Errorf("access denied to organization %s", r.OrganizationID)
	}
}
