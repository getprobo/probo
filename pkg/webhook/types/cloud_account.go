// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package types

import (
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// CloudAccount is the redacted webhook payload shape for cloud
// account lifecycle events. It deliberately omits credential
// envelope, scope identifier, external_id, and last_probe_error so
// webhook subscribers never receive secrets or reconnaissance fields.
type CloudAccount struct {
	ID                  gid.GID                             `json:"id"`
	OrganizationID      gid.GID                             `json:"organizationId"`
	Label               string                              `json:"label"`
	Provider            coredata.CloudAccountProvider       `json:"provider"`
	CredentialKind      coredata.CloudAccountCredentialKind `json:"credentialKind"`
	ScopeKind           coredata.CloudAccountScopeKind      `json:"scopeKind"`
	EnabledAuditModules []coredata.CloudAccountAuditModule  `json:"enabledAuditModules"`
	Status              coredata.CloudAccountStatus         `json:"status"`
	LastVerifiedAt      *time.Time                          `json:"lastVerifiedAt"`
	CreatedAt           time.Time                           `json:"createdAt"`
	UpdatedAt           time.Time                           `json:"updatedAt"`
}

// NewCloudAccount returns the webhook payload shape for the supplied
// coredata cloud account. Callers must NOT pass any field whose
// content is credential-adjacent; the entity-to-payload mapping
// here is the single source of truth for what subscribers see.
func NewCloudAccount(account *coredata.CloudAccount) CloudAccount {
	return CloudAccount{
		ID:                  account.ID,
		OrganizationID:      account.OrganizationID,
		Label:               account.Label,
		Provider:            account.Provider,
		CredentialKind:      account.CredentialKind,
		ScopeKind:           account.ScopeKind,
		EnabledAuditModules: account.EnabledAuditModules,
		Status:              account.Status,
		LastVerifiedAt:      account.LastVerifiedAt,
		CreatedAt:           account.CreatedAt,
		UpdatedAt:           account.UpdatedAt,
	}
}
