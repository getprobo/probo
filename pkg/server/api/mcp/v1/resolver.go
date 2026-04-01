// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

//go:generate go run go.probo.inc/mcpgen generate

package mcp_v1

import (
	"context"
	"encoding/json"
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/server/api/authn"
)

type Resolver struct {
	proboSvc *probo.Service
	iamSvc   *iam.Service
	logger   *log.Logger
}

func markdownToProseMirrorJSON(markdown string) (string, error) {
	node, err := prosemirror.ParseMarkdown(markdown)
	if err != nil {
		return "", fmt.Errorf("cannot parse markdown: %w", err)
	}

	out, err := json.Marshal(node)
	if err != nil {
		return "", fmt.Errorf("cannot marshal prosemirror node: %w", err)
	}

	return string(out), nil
}

func (r *Resolver) MustAuthorize(ctx context.Context, entityID gid.GID, action iam.Action) {
	identity := authn.IdentityFromContext(ctx)

	err := r.iamSvc.Authorizer.Authorize(
		ctx,
		iam.AuthorizeParams{
			Principal: identity.ID,
			Resource:  entityID,
			Action:    action,
		},
	)
	if err != nil {
		panic(err)
	}
}
