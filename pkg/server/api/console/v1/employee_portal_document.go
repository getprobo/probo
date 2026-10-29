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

	"github.com/vikstrous/dataloadgen"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/console/v1/dataloader"
	"go.probo.inc/probo/pkg/server/gqlx"
)

func (r *employeePortalDocumentResolver) publishedVersion(
	ctx context.Context,
	documentID gid.GID,
) (*coredata.DocumentVersion, error) {
	version, err := dataloader.FromContext(ctx).CurrentPublishedDocumentVersion.Load(ctx, documentID)
	if err != nil {
		if errors.Is(err, dataloadgen.ErrNotFound) {
			return nil, gqlx.NotFoundf(ctx, "published document %q not found", documentID)
		}

		r.logger.ErrorCtx(ctx, "cannot load published document version", log.Error(err))

		return nil, gqlx.Internal(ctx)
	}

	return version, nil
}
