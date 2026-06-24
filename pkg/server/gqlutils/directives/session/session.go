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

package session

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

// Directive enforces that the request carries a real session. It is meant to
// be combined with @authentication(required: PRESENT): @authentication handles
// the "must be authenticated" check, while @sessionOnly additionally rejects
// API key authentication, which also carries an identity but no session.
func Directive(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	if authn.SessionFromContext(ctx) == nil {
		return nil, gqlutils.Forbiddenf(
			ctx,
			"this resource can only be accessed with session authentication, not API key authentication",
		)
	}

	return next(ctx)
}
