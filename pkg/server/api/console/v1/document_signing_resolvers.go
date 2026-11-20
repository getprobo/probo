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

package console_v1

import (
	"context"
	"errors"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/console/v1/schema"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlx"
	"go.probo.inc/probo/pkg/sign"
)

// parseStrCursorKey converts an optional string cursor key into a page.CursorKey.
func parseStrCursorKey(s *string) *page.CursorKey {
	if s == nil {
		return nil
	}
	ck, err := page.ParseCursorKey(*s)
	if err != nil {
		return nil
	}
	return &ck
}

// DocumentSigningConfig returns the DocumentSigningConfigResolver.
func (r *Resolver) DocumentSigningConfig() schema.DocumentSigningConfigResolver {
	return &documentSigningConfigResolver{r}
}

// ClickwrapAcceptance returns the ClickwrapAcceptanceResolver.
func (r *Resolver) ClickwrapAcceptance() schema.ClickwrapAcceptanceResolver {
	return &clickwrapAcceptanceResolver{r}
}

// ClickwrapAcceptanceConnection returns the ClickwrapAcceptanceConnectionResolver.
func (r *Resolver) ClickwrapAcceptanceConnection() schema.ClickwrapAcceptanceConnectionResolver {
	return &clickwrapAcceptanceConnectionResolver{r}
}

// ClickwrapSubject returns the ClickwrapSubjectResolver.
func (r *Resolver) ClickwrapSubject() schema.ClickwrapSubjectResolver {
	return &clickwrapSubjectResolver{r}
}

// ClickwrapSubjectConnection returns the ClickwrapSubjectConnectionResolver.
func (r *Resolver) ClickwrapSubjectConnection() schema.ClickwrapSubjectConnectionResolver {
	return &clickwrapSubjectConnectionResolver{r}
}

type (
	documentSigningConfigResolver         struct{ *Resolver }
	clickwrapAcceptanceResolver           struct{ *Resolver }
	clickwrapAcceptanceConnectionResolver struct{ *Resolver }
	clickwrapSubjectResolver              struct{ *Resolver }
	clickwrapSubjectConnectionResolver    struct{ *Resolver }
)

// --- DocumentSigningConfig field resolvers ---

// Document resolves the document field on DocumentSigningConfig.
func (r *documentSigningConfigResolver) Document(ctx context.Context, obj *types.DocumentSigningConfig) (*types.Document, error) {
	if _, err := r.authorize(ctx, obj.ID, probo.ActionDocumentGet); err != nil {
		return nil, err
	}

	doc, err := r.probo.Documents.Get(ctx, r.scopeFromGID(ctx, obj.ID), obj.Document.ID)
	if err != nil {
		return nil, gqlx.MapError(ctx, r.logger, "cannot get document", err)
	}

	return types.NewDocument(doc), nil
}

// Acceptances resolves the acceptances field on DocumentSigningConfig.
func (r *documentSigningConfigResolver) Acceptances(
	ctx context.Context,
	obj *types.DocumentSigningConfig,
	first *int, after *string, last *int, before *string,
) (*types.ClickwrapAcceptanceConnection, error) {
	scope, err := r.authorize(ctx, obj.ID, probo.ActionClickwrapAcceptanceList)
	if err != nil {
		return nil, err
	}

	cursor := types.NewCursor(
		first, parseStrCursorKey(after), last, parseStrCursorKey(before),
		page.OrderBy[coredata.ClickwrapAcceptanceOrderField]{
			Field:     coredata.ClickwrapAcceptanceOrderFieldAcceptedAt,
			Direction: page.OrderDirectionDesc,
		},
	)

	conn, err := r.sign.ListAcceptancesForConfig(ctx, scope, obj.ID, cursor)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot list acceptances", log.Error(err))
		return nil, gqlx.Internal(ctx)
	}

	return types.NewClickwrapAcceptanceConnection(conn), nil
}

// IntegrationHealth resolves the integrationHealth field on DocumentSigningConfig.
func (r *documentSigningConfigResolver) IntegrationHealth(
	ctx context.Context,
	obj *types.DocumentSigningConfig,
) (*types.DocumentSigningIntegrationHealth, error) {
	if _, err := r.authorize(ctx, obj.ID, probo.ActionDocumentSigningConfigGet); err != nil {
		return nil, err
	}

	health, err := r.sign.GetIntegrationHealth(ctx, obj.ID)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot get integration health", log.Error(err))
		return nil, gqlx.Internal(ctx)
	}

	entries := make([]*types.DataPlaneRequestLogEntry, len(health.LastRejectedRequests))
	for i, l := range health.LastRejectedRequests {
		entries[i] = types.NewDataPlaneRequestLogEntry(l)
	}

	return &types.DocumentSigningIntegrationHealth{
		LastRejectedRequests: entries,
		LastAcceptanceAt:     health.LastAcceptanceAt,
	}, nil
}

// --- ClickwrapAcceptance field resolvers ---

// SigningConfig resolves the signingConfig field on ClickwrapAcceptance.
func (r *clickwrapAcceptanceResolver) SigningConfig(
	ctx context.Context,
	obj *types.ClickwrapAcceptance,
) (*types.DocumentSigningConfig, error) {
	if _, err := r.authorize(ctx, obj.ID, probo.ActionDocumentSigningConfigGet); err != nil {
		return nil, err
	}

	cfg, err := r.sign.GetSigningConfig(ctx, r.scopeFromGID(ctx, obj.ID), obj.SigningConfig.ID)
	if err != nil {
		if errors.Is(err, sign.ErrConfigNotFound) {
			return nil, gqlx.NotFound(ctx, err)
		}
		r.logger.ErrorCtx(ctx, "cannot get signing config", log.Error(err))
		return nil, gqlx.Internal(ctx)
	}

	return types.NewDocumentSigningConfig(cfg), nil
}

// DocumentVersion resolves the documentVersion field on ClickwrapAcceptance.
func (r *clickwrapAcceptanceResolver) DocumentVersion(
	ctx context.Context,
	obj *types.ClickwrapAcceptance,
) (*types.DocumentVersion, error) {
	if _, err := r.authorize(ctx, obj.ID, probo.ActionDocumentVersionGet); err != nil {
		return nil, err
	}

	dv, err := r.probo.Documents.GetVersion(ctx, r.scopeFromGID(ctx, obj.ID), obj.DocumentVersion.ID)
	if err != nil {
		return nil, gqlx.MapError(ctx, r.logger, "cannot get document version", err)
	}

	return types.NewDocumentVersion(dv), nil
}

// Subject resolves the subject field on ClickwrapAcceptance.
func (r *clickwrapAcceptanceResolver) Subject(
	ctx context.Context,
	obj *types.ClickwrapAcceptance,
) (*types.ClickwrapSubject, error) {
	if _, err := r.authorize(ctx, obj.ID, probo.ActionClickwrapSubjectGet); err != nil {
		return nil, err
	}

	subj, err := r.sign.GetSubject(ctx, r.scopeFromGID(ctx, obj.ID), obj.Subject.ID)
	if err != nil {
		if errors.Is(err, sign.ErrSubjectNotFound) {
			return nil, gqlx.NotFound(ctx, err)
		}
		r.logger.ErrorCtx(ctx, "cannot get clickwrap subject", log.Error(err))
		return nil, gqlx.Internal(ctx)
	}

	return types.NewClickwrapSubject(subj), nil
}

// --- ClickwrapAcceptanceConnection field resolvers ---

// TotalCount resolves the totalCount field on ClickwrapAcceptanceConnection.
func (r *clickwrapAcceptanceConnectionResolver) TotalCount(
	ctx context.Context,
	obj *types.ClickwrapAcceptanceConnection,
) (int, error) {
	return obj.TotalCount, nil
}

// --- ClickwrapSubject field resolvers ---

// Organization resolves the organization field on ClickwrapSubject.
func (r *clickwrapSubjectResolver) Organization(
	ctx context.Context,
	obj *types.ClickwrapSubject,
) (*types.Organization, error) {
	if _, err := r.authorize(ctx, obj.ID, probo.ActionOrganizationGet); err != nil {
		return nil, err
	}

	org, err := r.probo.Organizations.Get(ctx, r.scopeFromGID(ctx, obj.ID), obj.Organization.ID)
	if err != nil {
		return nil, gqlx.MapError(ctx, r.logger, "cannot get organization", err)
	}

	return types.NewOrganization(org), nil
}

// Acceptances resolves the acceptances field on ClickwrapSubject.
func (r *clickwrapSubjectResolver) Acceptances(
	ctx context.Context,
	obj *types.ClickwrapSubject,
	first *int, after *string, last *int, before *string,
) (*types.ClickwrapAcceptanceConnection, error) {
	scope, err := r.authorize(ctx, obj.ID, probo.ActionClickwrapAcceptanceList)
	if err != nil {
		return nil, err
	}

	cursor := types.NewCursor(
		first, parseStrCursorKey(after), last, parseStrCursorKey(before),
		page.OrderBy[coredata.ClickwrapAcceptanceOrderField]{
			Field:     coredata.ClickwrapAcceptanceOrderFieldAcceptedAt,
			Direction: page.OrderDirectionDesc,
		},
	)

	conn, err := r.sign.ListAcceptancesForSubject(ctx, scope, obj.ID, cursor)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot list subject acceptances", log.Error(err))
		return nil, gqlx.Internal(ctx)
	}

	return types.NewClickwrapAcceptanceConnection(conn), nil
}

// --- ClickwrapSubjectConnection field resolvers ---

// TotalCount resolves the totalCount field on ClickwrapSubjectConnection.
func (r *clickwrapSubjectConnectionResolver) TotalCount(
	ctx context.Context,
	obj *types.ClickwrapSubjectConnection,
) (int, error) {
	return obj.TotalCount, nil
}

// --- Mutation resolvers ---

// CreateDocumentSigningConfig creates a new signing config.
func (r *mutationResolver) CreateDocumentSigningConfig(
	ctx context.Context,
	input types.CreateDocumentSigningConfigInput,
) (*types.CreateDocumentSigningConfigPayload, error) {
	scope, err := r.authorize(ctx, input.DocumentID, probo.ActionDocumentSigningConfigCreate)
	if err != nil {
		return nil, err
	}

	doc, err := r.probo.Documents.Get(ctx, scope, input.DocumentID)
	if err != nil {
		return nil, gqlx.MapError(ctx, r.logger, "cannot get document", err)
	}

	mode := coredata.DocumentSigningConfigModeLive
	if input.Mode != nil {
		mode = *input.Mode
	}

	var origins []string
	if input.AllowedOrigins != nil {
		origins = input.AllowedOrigins
	}

	cfg, err := r.sign.CreateSigningConfig(ctx, scope, sign.CreateSigningConfigRequest{
		OrganizationID:     doc.OrganizationID,
		DocumentID:         input.DocumentID,
		Kind:               input.Kind,
		ReacceptancePolicy: input.ReacceptancePolicy,
		Mode:               mode,
		AllowedOrigins:     origins,
	})
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot create signing config", log.Error(err))
		return nil, gqlx.Internal(ctx)
	}

	return &types.CreateDocumentSigningConfigPayload{
		DocumentSigningConfig: types.NewDocumentSigningConfig(cfg),
	}, nil
}

// UpdateDocumentSigningConfig updates an existing signing config.
func (r *mutationResolver) UpdateDocumentSigningConfig(
	ctx context.Context,
	input types.UpdateDocumentSigningConfigInput,
) (*types.UpdateDocumentSigningConfigPayload, error) {
	scope, err := r.authorize(ctx, input.ID, probo.ActionDocumentSigningConfigUpdate)
	if err != nil {
		return nil, err
	}

	var allowedOrigins *[]string
	if input.AllowedOrigins != nil {
		allowedOrigins = &input.AllowedOrigins
	}

	cfg, err := r.sign.UpdateSigningConfig(ctx, scope, sign.UpdateSigningConfigRequest{
		ID:                 input.ID,
		Kind:               input.Kind,
		ReacceptancePolicy: input.ReacceptancePolicy,
		AllowedOrigins:     allowedOrigins,
	})
	if err != nil {
		if errors.Is(err, sign.ErrConfigNotFound) {
			return nil, gqlx.NotFound(ctx, err)
		}
		r.logger.ErrorCtx(ctx, "cannot update signing config", log.Error(err))
		return nil, gqlx.Internal(ctx)
	}

	return &types.UpdateDocumentSigningConfigPayload{
		DocumentSigningConfig: types.NewDocumentSigningConfig(cfg),
	}, nil
}

// ActivateDocumentSigningConfig activates a signing config.
func (r *mutationResolver) ActivateDocumentSigningConfig(
	ctx context.Context,
	id gid.GID,
) (*types.ActivateDocumentSigningConfigPayload, error) {
	scope, err := r.authorize(ctx, id, probo.ActionDocumentSigningConfigActivate)
	if err != nil {
		return nil, err
	}

	cfg, err := r.sign.ActivateSigningConfig(ctx, scope, id)
	if err != nil {
		if errors.Is(err, sign.ErrConfigNotFound) {
			return nil, gqlx.NotFound(ctx, err)
		}
		r.logger.ErrorCtx(ctx, "cannot activate signing config", log.Error(err))
		return nil, gqlx.Internal(ctx)
	}

	return &types.ActivateDocumentSigningConfigPayload{
		DocumentSigningConfig: types.NewDocumentSigningConfig(cfg),
	}, nil
}

// ArchiveDocumentSigningConfig archives a signing config.
func (r *mutationResolver) ArchiveDocumentSigningConfig(
	ctx context.Context,
	id gid.GID,
) (*types.ArchiveDocumentSigningConfigPayload, error) {
	scope, err := r.authorize(ctx, id, probo.ActionDocumentSigningConfigArchive)
	if err != nil {
		return nil, err
	}

	cfg, err := r.sign.ArchiveSigningConfig(ctx, scope, id)
	if err != nil {
		if errors.Is(err, sign.ErrConfigNotFound) {
			return nil, gqlx.NotFound(ctx, err)
		}
		r.logger.ErrorCtx(ctx, "cannot archive signing config", log.Error(err))
		return nil, gqlx.Internal(ctx)
	}

	return &types.ArchiveDocumentSigningConfigPayload{
		DocumentSigningConfig: types.NewDocumentSigningConfig(cfg),
	}, nil
}

// RotateOrganizationSigningSecret rotates the signing secret for an org.
func (r *mutationResolver) RotateOrganizationSigningSecret(
	ctx context.Context,
	organizationID gid.GID,
	mode *coredata.DocumentSigningConfigMode,
) (*types.RotateOrganizationSigningSecretPayload, error) {
	scope, err := r.authorize(ctx, organizationID, probo.ActionOrganizationSigningSecretRotate)
	if err != nil {
		return nil, err
	}

	m := coredata.DocumentSigningConfigModeLive
	if mode != nil {
		m = *mode
	}

	plaintext, _, err := r.sign.RotateOrganizationSigningSecret(ctx, scope, organizationID, m)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot rotate signing secret", log.Error(err))
		return nil, gqlx.Internal(ctx)
	}

	return &types.RotateOrganizationSigningSecretPayload{
		SecretKeyPlaintext: plaintext,
	}, nil
}

// RevokeClickwrapAcceptance revokes a clickwrap acceptance.
func (r *mutationResolver) RevokeClickwrapAcceptance(
	ctx context.Context,
	id gid.GID,
	reason *string,
) (*types.RevokeClickwrapAcceptancePayload, error) {
	scope, err := r.authorize(ctx, id, probo.ActionClickwrapAcceptanceRevoke)
	if err != nil {
		return nil, err
	}

	acceptance, err := r.sign.RevokeAcceptance(ctx, scope, id, reason)
	if err != nil {
		if errors.Is(err, sign.ErrAcceptanceNotFound) {
			return nil, gqlx.NotFound(ctx, err)
		}
		if errors.Is(err, sign.ErrAlreadyRevoked) {
			return nil, gqlx.Conflict(ctx, err)
		}
		r.logger.ErrorCtx(ctx, "cannot revoke acceptance", log.Error(err))
		return nil, gqlx.Internal(ctx)
	}

	return &types.RevokeClickwrapAcceptancePayload{
		Acceptance: types.NewClickwrapAcceptance(acceptance),
	}, nil
}

// scopeFromGID creates a Scoper from a GID's tenant ID.
func (r *Resolver) scopeFromGID(ctx context.Context, id gid.GID) *coredata.Scope {
	_ = ctx
	return coredata.NewScope(id.TenantID(), gid.Nil)
}
