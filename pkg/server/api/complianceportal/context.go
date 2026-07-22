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

package complianceportal

import (
	"context"

	"go.probo.inc/probo/pkg/coredata"
)

type ctxKey struct{ name string }

var (
	compliancePortalKey        = &ctxKey{name: "compliance_portal"}
	compliancePortalBaseURLKey = &ctxKey{name: "compliance_portal_base_url"}
)

func CompliancePortalFromContext(ctx context.Context) *coredata.CompliancePortal {
	page, _ := ctx.Value(compliancePortalKey).(*coredata.CompliancePortal)
	return page
}

func ContextWithCompliancePortal(ctx context.Context, page *coredata.CompliancePortal) context.Context {
	return context.WithValue(ctx, compliancePortalKey, page)
}

func CompliancePortalBaseURLFromContext(ctx context.Context) *string {
	page, _ := ctx.Value(compliancePortalBaseURLKey).(*string)
	return page
}

// ContextWithCompliancePortalBaseURL stores the portal origin (scheme://host).
// Callers must not include a path — consumers append their own routes.
func ContextWithCompliancePortalBaseURL(ctx context.Context, baseURL string) context.Context {
	return context.WithValue(ctx, compliancePortalBaseURLKey, &baseURL)
}
