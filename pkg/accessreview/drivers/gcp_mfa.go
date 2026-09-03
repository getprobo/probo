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

package drivers

import (
	"context"
	"fmt"
	"strings"

	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/coredata"
	admin "google.golang.org/api/admin/directory/v1"
)

func fetchGCPMFA(
	ctx context.Context,
	session *cloudgcp.Session,
	records []AccountRecord,
) (map[string]coredata.MFAStatus, error) {
	emails := gcpActivityEmails(records, gcpPrincipalUser)
	if len(emails) == 0 {
		return nil, nil
	}

	svc, err := admin.NewService(ctx, session.ServiceOptions()...)
	if err != nil {
		return nil, fmt.Errorf("cannot create gcp admin directory client: %w", err)
	}

	found := make(map[string]coredata.MFAStatus, len(emails))

	for _, email := range emails {
		user, err := svc.Users.Get(email).Projection("full").Context(ctx).Do()
		if err != nil {
			if ctx.Err() != nil {
				return found, err
			}

			if cloudgcp.As[cloudgcp.ErrPermissionDenied](err) {
				return found, fmt.Errorf("cannot read gcp directory mfa enrollment: %w", err)
			}

			if cloudgcp.As[cloudgcp.ErrNotFound](err) {
				continue
			}

			return found, fmt.Errorf("cannot get gcp directory user: %w", err)
		}

		status := coredata.MFAStatusDisabled
		if user != nil && user.IsEnrolledIn2Sv {
			status = coredata.MFAStatusEnabled
		}

		found[strings.ToLower(email)] = status
	}

	return found, nil
}
