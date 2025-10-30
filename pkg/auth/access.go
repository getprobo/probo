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
	AuthMethodAny
)

type OrgAuthRequirement struct {
	OrganizationID gid.GID
	EmailDomain    string
	SAMLConfig     *coredata.SAMLConfiguration
}

type AccessResult struct {
	OrganizationID gid.GID
	Allowed        bool
	MissingAuth    AuthMethod
	SAMLConfig     *coredata.SAMLConfiguration
}

func (r OrgAuthRequirement) Check(session coredata.SessionData) AccessResult {
	if r.SAMLConfig == nil || !r.SAMLConfig.Enabled || !r.SAMLConfig.DomainVerified {
		return AccessResult{
			OrganizationID: r.OrganizationID,
			Allowed:        session.PasswordAuthenticated,
			MissingAuth:    AuthMethodPassword,
			SAMLConfig:     nil,
		}
	}

	orgKey := r.OrganizationID.String()
	_, hasSAML := session.SAMLAuthenticatedOrgs[orgKey]

	if r.SAMLConfig.EnforcementPolicy == coredata.SAMLEnforcementPolicyRequired {
		return AccessResult{
			OrganizationID: r.OrganizationID,
			Allowed:        hasSAML,
			MissingAuth:    AuthMethodSAML,
			SAMLConfig:     r.SAMLConfig,
		}
	}

	hasAnyAuth := session.PasswordAuthenticated || hasSAML
	missingAuth := AuthMethodAny
	if hasAnyAuth {
		missingAuth = AuthMethodPassword
	}

	return AccessResult{
		OrganizationID: r.OrganizationID,
		Allowed:        hasAnyAuth,
		MissingAuth:    missingAuth,
		SAMLConfig:     r.SAMLConfig,
	}
}

func (r AccessResult) ToError(baseURL string) error {
	if r.Allowed {
		return nil
	}

	switch r.MissingAuth {
	case AuthMethodPassword:
		return ErrPasswordAuthRequired{
			OrganizationID: r.OrganizationID,
			RedirectURL:    fmt.Sprintf("%s/auth/login?method=password", baseURL),
		}
	case AuthMethodSAML, AuthMethodAny:
		return ErrSAMLAuthRequired{
			ConfigID:       r.SAMLConfig.ID,
			OrganizationID: r.OrganizationID,
			RedirectURL:    fmt.Sprintf("%s/connect/saml/login/%s", baseURL, r.SAMLConfig.ID),
		}
	default:
		return fmt.Errorf("access denied to organization %s", r.OrganizationID)
	}
}
