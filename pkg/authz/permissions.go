package authz

import (
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
)

type Action string

const (
	ActionGet    Action = "get"
	ActionCount  Action = "count"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

type ActionPermissions map[Action]bool

type EntityPermissions map[uint16]ActionPermissions

type Role string

const (
	RoleOwner  Role = "OWNER"
	RoleAdmin  Role = "ADMIN"
	RoleViewer Role = "VIEWER"
	RoleFull   Role = "FULL"
)

var (
	fullAccess = ActionPermissions{
		ActionGet:    true,
		ActionCount:  true,
		ActionCreate: true,
		ActionUpdate: true,
		ActionDelete: true,
	}

	fullExceptDelete = ActionPermissions{
		ActionGet:    true,
		ActionCount:  true,
		ActionCreate: true,
		ActionUpdate: true,
		ActionDelete: false,
	}

	readOnly = ActionPermissions{
		ActionGet:    true,
		ActionCount:  true,
		ActionCreate: false,
		ActionUpdate: false,
		ActionDelete: false,
	}

	noAccess = ActionPermissions{
		ActionGet:    false,
		ActionCount:  false,
		ActionCreate: false,
		ActionUpdate: false,
		ActionDelete: false,
	}

	RolePermissions = map[Role]EntityPermissions{
		RoleOwner: {
			coredata.OrganizationEntityType:             fullAccess,
			coredata.MembershipEntityType:               fullAccess,
			coredata.InvitationEntityType:               fullAccess,
			coredata.DocumentEntityType:                 fullAccess,
			coredata.DocumentVersionEntityType:          fullAccess,
			coredata.DocumentVersionSignatureEntityType: fullAccess,
			coredata.FrameworkEntityType:                fullAccess,
			coredata.ControlEntityType:                  fullAccess,
			coredata.MeasureEntityType:                  fullAccess,
			coredata.EvidenceEntityType:                 fullAccess,
			coredata.VendorEntityType:                   fullAccess,
			coredata.VendorContactEntityType:            fullAccess,
			coredata.VendorServiceEntityType:            fullAccess,
			coredata.PeopleEntityType:                   fullAccess,
			coredata.RiskEntityType:                     fullAccess,
			coredata.TaskEntityType:                     fullAccess,
			coredata.AssetEntityType:                    fullAccess,
			coredata.DatumEntityType:                    fullAccess,
			coredata.AuditEntityType:                    fullAccess,
			coredata.NonconformityEntityType:            fullAccess,
			coredata.ObligationEntityType:               fullAccess,
			coredata.ContinualImprovementEntityType:     fullAccess,
			coredata.ProcessingActivityEntityType:       fullAccess,
			coredata.SnapshotEntityType:                 fullAccess,
			coredata.TrustCenterEntityType:              fullAccess,
			coredata.TrustCenterFileEntityType:          fullAccess,
			coredata.TrustCenterAccessEntityType:        fullAccess,
			coredata.CustomDomainEntityType:             fullAccess,
			coredata.SAMLConfigurationEntityType:        fullAccess,
			coredata.ConnectorEntityType:                fullAccess,
			coredata.UserAPIKeyEntityType:               fullAccess,
			coredata.UserAPIKeyMembershipEntityType:     fullAccess,
		},

		RoleAdmin: {
			coredata.OrganizationEntityType:             fullExceptDelete,
			coredata.MembershipEntityType:               fullAccess,
			coredata.InvitationEntityType:               fullAccess,
			coredata.DocumentEntityType:                 fullAccess,
			coredata.DocumentVersionEntityType:          fullAccess,
			coredata.DocumentVersionSignatureEntityType: fullAccess,
			coredata.FrameworkEntityType:                fullAccess,
			coredata.ControlEntityType:                  fullAccess,
			coredata.MeasureEntityType:                  fullAccess,
			coredata.EvidenceEntityType:                 fullAccess,
			coredata.VendorEntityType:                   fullAccess,
			coredata.VendorContactEntityType:            fullAccess,
			coredata.VendorServiceEntityType:            fullAccess,
			coredata.PeopleEntityType:                   fullAccess,
			coredata.RiskEntityType:                     fullAccess,
			coredata.TaskEntityType:                     fullAccess,
			coredata.AssetEntityType:                    fullAccess,
			coredata.DatumEntityType:                    fullAccess,
			coredata.AuditEntityType:                    fullAccess,
			coredata.NonconformityEntityType:            fullAccess,
			coredata.ObligationEntityType:               fullAccess,
			coredata.ContinualImprovementEntityType:     fullAccess,
			coredata.ProcessingActivityEntityType:       fullAccess,
			coredata.SnapshotEntityType:                 fullAccess,
			coredata.TrustCenterEntityType:              fullAccess,
			coredata.TrustCenterFileEntityType:          fullAccess,
			coredata.TrustCenterAccessEntityType:        fullAccess,
			coredata.CustomDomainEntityType:             readOnly,
			coredata.SAMLConfigurationEntityType:        readOnly,
			coredata.ConnectorEntityType:                readOnly,
			coredata.UserAPIKeyEntityType:               noAccess,
			coredata.UserAPIKeyMembershipEntityType:     noAccess,
		},

		RoleViewer: {
			coredata.OrganizationEntityType:             readOnly,
			coredata.MembershipEntityType:               readOnly,
			coredata.InvitationEntityType:               readOnly,
			coredata.DocumentEntityType:                 readOnly,
			coredata.DocumentVersionEntityType:          readOnly,
			coredata.DocumentVersionSignatureEntityType: readOnly,
			coredata.FrameworkEntityType:                readOnly,
			coredata.ControlEntityType:                  readOnly,
			coredata.MeasureEntityType:                  readOnly,
			coredata.EvidenceEntityType:                 readOnly,
			coredata.VendorEntityType:                   readOnly,
			coredata.VendorContactEntityType:            readOnly,
			coredata.VendorServiceEntityType:            readOnly,
			coredata.PeopleEntityType:                   readOnly,
			coredata.RiskEntityType:                     readOnly,
			coredata.TaskEntityType:                     readOnly,
			coredata.AssetEntityType:                    readOnly,
			coredata.DatumEntityType:                    readOnly,
			coredata.AuditEntityType:                    readOnly,
			coredata.NonconformityEntityType:            readOnly,
			coredata.ObligationEntityType:               readOnly,
			coredata.ContinualImprovementEntityType:     readOnly,
			coredata.ProcessingActivityEntityType:       readOnly,
			coredata.SnapshotEntityType:                 readOnly,
			coredata.TrustCenterEntityType:              readOnly,
			coredata.TrustCenterFileEntityType:          readOnly,
			coredata.TrustCenterAccessEntityType:        readOnly,
			coredata.CustomDomainEntityType:             readOnly,
			coredata.SAMLConfigurationEntityType:        readOnly,
			coredata.ConnectorEntityType:                readOnly,
			coredata.UserAPIKeyEntityType:               noAccess,
			coredata.UserAPIKeyMembershipEntityType:     noAccess,
		},

		RoleFull: {
			coredata.OrganizationEntityType:             fullExceptDelete,
			coredata.MembershipEntityType:               fullAccess,
			coredata.InvitationEntityType:               fullAccess,
			coredata.DocumentEntityType:                 fullAccess,
			coredata.DocumentVersionEntityType:          fullAccess,
			coredata.DocumentVersionSignatureEntityType: fullAccess,
			coredata.FrameworkEntityType:                fullAccess,
			coredata.ControlEntityType:                  fullAccess,
			coredata.MeasureEntityType:                  fullAccess,
			coredata.EvidenceEntityType:                 fullAccess,
			coredata.VendorEntityType:                   fullAccess,
			coredata.VendorContactEntityType:            fullAccess,
			coredata.VendorServiceEntityType:            fullAccess,
			coredata.PeopleEntityType:                   fullAccess,
			coredata.RiskEntityType:                     fullAccess,
			coredata.TaskEntityType:                     fullAccess,
			coredata.AssetEntityType:                    fullAccess,
			coredata.DatumEntityType:                    fullAccess,
			coredata.AuditEntityType:                    fullAccess,
			coredata.NonconformityEntityType:            fullAccess,
			coredata.ObligationEntityType:               fullAccess,
			coredata.ContinualImprovementEntityType:     fullAccess,
			coredata.ProcessingActivityEntityType:       fullAccess,
			coredata.SnapshotEntityType:                 fullAccess,
			coredata.TrustCenterEntityType:              fullAccess,
			coredata.TrustCenterFileEntityType:          fullAccess,
			coredata.TrustCenterAccessEntityType:        fullAccess,
			coredata.CustomDomainEntityType:             fullAccess,
			coredata.SAMLConfigurationEntityType:        fullAccess,
			coredata.ConnectorEntityType:                fullAccess,
			coredata.UserAPIKeyEntityType:               noAccess,
			coredata.UserAPIKeyMembershipEntityType:     noAccess,
		},
	}
)

func authorizeForRole(role Role, entityType uint16, action Action) error {
	entityPerms, ok := RolePermissions[role]
	if !ok {
		return fmt.Errorf("unknown role: %s", role)
	}

	actionPerms, ok := entityPerms[entityType]
	if !ok {
		return fmt.Errorf("no permissions defined for entity type %d", entityType)
	}

	if !actionPerms[action] {
		return &PermissionDeniedError{Message: fmt.Sprintf("insufficient permissions for action %s on entity type %d", action, entityType)}
	}

	return nil
}

func canAssignRole(currentRole Role, targetRole Role) error {
	if currentRole == RoleOwner || currentRole == RoleFull {
		return nil
	}

	if currentRole == RoleAdmin {
		if targetRole == RoleOwner {
			return &PermissionDeniedError{Message: "admin users cannot assign owner role"}
		}
		return nil
	}

	return &PermissionDeniedError{Message: fmt.Sprintf("role %s cannot assign roles", currentRole)}
}
