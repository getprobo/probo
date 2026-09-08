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
	"net/url"

	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	azureUserRegistrationDetails struct {
		ID           string `json:"id"`
		IsMFACapable *bool  `json:"isMfaCapable"`
	}

	azureUserRegistrationDetailsPage struct {
		Value    []azureUserRegistrationDetails `json:"value"`
		NextLink string                         `json:"@odata.nextLink"`
	}
)

func fetchAzureMFA(
	ctx context.Context,
	session *cloudazure.Session,
	records []AccountRecord,
) (map[string]coredata.MFAStatus, error) {
	wanted := azureUserIDSet(records)
	if len(wanted) == 0 {
		return nil, nil
	}

	pageURL, err := azureUserRegistrationDetailsURL(session.GraphBaseURL())
	if err != nil {
		return nil, err
	}

	found := make(map[string]coredata.MFAStatus, len(wanted))

	for range maxPaginationPages {
		var page azureUserRegistrationDetailsPage
		if err := fetchAzureGraphJSON(ctx, session, pageURL, &page); err != nil {
			if ctx.Err() != nil {
				return found, err
			}

			if azureGraphEnrichmentUnavailable(err) {
				return nil, fmt.Errorf("cannot read azure mfa registration: %w", err)
			}

			return nil, fmt.Errorf("cannot list azure mfa registration: %w", err)
		}

		for _, details := range page.Value {
			if _, ok := wanted[details.ID]; !ok {
				continue
			}

			status, ok := azureRegistrationMFAStatus(details)
			if !ok {
				continue
			}

			found[details.ID] = status
		}

		if page.NextLink == "" {
			return found, nil
		}

		pageURL, err = sameHostNextPageURL("azure", session.GraphBaseURL(), page.NextLink)
		if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("cannot list all azure mfa registration: %w", ErrPaginationLimitReached)
}

func azureRegistrationMFAStatus(
	details azureUserRegistrationDetails,
) (coredata.MFAStatus, bool) {
	if details.IsMFACapable == nil {
		return coredata.MFAStatusUnknown, false
	}

	if *details.IsMFACapable {
		return coredata.MFAStatusEnabled, true
	}

	return coredata.MFAStatusDisabled, true
}

func azureUserRegistrationDetailsURL(graphBaseURL string) (string, error) {
	endpoint, err := url.JoinPath(
		graphBaseURL,
		azureGraphAPIVersion,
		azureGraphReportsPath,
		azureGraphAuthenticationMethodsPath,
		azureGraphUserRegistrationDetailsPath,
	)
	if err != nil {
		return "", fmt.Errorf("cannot build azure mfa registration URL: %w", err)
	}

	return endpoint, nil
}
