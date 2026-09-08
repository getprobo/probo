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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// azureActivityTimeout caps enrichment so a slow Graph walk cannot
	// spend the campaign fetch budget and then drop the identity list.
	azureActivityTimeout = 20 * time.Second

	azureSignInActivitySelect      = "id,signInActivity"
	azureSignInActivityPageSize    = 500
	azureSignInActivityFilterChunk = 15
	azureSignInActivityBatchChunk  = 20
)

type (
	azureSignInActivity struct {
		LastSuccessfulSignInDateTime string `json:"lastSuccessfulSignInDateTime"`
		LastSignInDateTime           string `json:"lastSignInDateTime"`
	}

	azureUserSignIn struct {
		ID             string              `json:"id"`
		SignInActivity azureSignInActivity `json:"signInActivity"`
	}

	azureUsersPage struct {
		Value    []azureUserSignIn `json:"value"`
		NextLink string            `json:"@odata.nextLink"`
	}

	azureBatchRequest struct {
		Requests []azureBatchRequestItem `json:"requests"`
	}

	azureBatchRequestItem struct {
		ID     string `json:"id"`
		Method string `json:"method"`
		URL    string `json:"url"`
	}

	azureBatchResponse struct {
		Responses []azureBatchResponseItem `json:"responses"`
	}

	azureBatchResponseItem struct {
		ID     string          `json:"id"`
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	}
)

func enrichAzureIdentities(
	ctx context.Context,
	session *cloudazure.Session,
	records []AccountRecord,
) error {
	if len(records) == 0 {
		return nil
	}

	enrichCtx, cancel := context.WithTimeout(ctx, azureActivityTimeout)
	defer cancel()

	logins, activityErr := fetchAzureActivity(enrichCtx, session, records)
	if errors.Is(activityErr, context.Canceled) && ctx.Err() != nil {
		return activityErr
	}

	mfa, mfaErr := fetchAzureMFA(enrichCtx, session, records)
	if errors.Is(mfaErr, context.Canceled) && ctx.Err() != nil {
		return mfaErr
	}

	applyAzureActivity(records, logins, mfa)

	if activityErr != nil {
		return activityErr
	}

	return mfaErr
}

func fetchAzureActivity(
	ctx context.Context,
	session *cloudazure.Session,
	records []AccountRecord,
) (map[string]time.Time, error) {
	ids := azureUserIDs(records)
	if len(ids) == 0 {
		return nil, nil
	}

	logins, err := listAzureSignInActivity(ctx, session, ids)
	if err == nil {
		return logins, nil
	}

	if ctx.Err() != nil {
		return nil, err
	}

	if azureGraphEnrichmentUnavailable(err) {
		return nil, fmt.Errorf("cannot read azure sign-in activity: %w", err)
	}

	if !azureGraphBadRequest(err) {
		return nil, fmt.Errorf("cannot list azure sign-in activity: %w", err)
	}

	logins, err = batchAzureSignInActivity(ctx, session, ids)
	if err != nil {
		return nil, fmt.Errorf("cannot batch azure sign-in activity: %w", err)
	}

	return logins, nil
}

func listAzureSignInActivity(
	ctx context.Context,
	session *cloudazure.Session,
	ids []string,
) (map[string]time.Time, error) {
	found := make(map[string]time.Time, len(ids))

	for start := 0; start < len(ids); start += azureSignInActivityFilterChunk {
		end := min(start+azureSignInActivityFilterChunk, len(ids))

		pageURL, err := azureUsersSignInActivityURL(session.GraphBaseURL(), ids[start:end])
		if err != nil {
			return nil, err
		}

		if err := collectAzureSignInActivityPages(ctx, session, pageURL, found); err != nil {
			return nil, err
		}
	}

	return found, nil
}

func collectAzureSignInActivityPages(
	ctx context.Context,
	session *cloudazure.Session,
	pageURL string,
	found map[string]time.Time,
) error {
	for range maxPaginationPages {
		var page azureUsersPage
		if err := fetchAzureGraphJSON(ctx, session, pageURL, &page); err != nil {
			return err
		}

		for _, user := range page.Value {
			if at, ok := azureSignInTime(user.SignInActivity); ok {
				found[user.ID] = at
			}
		}

		if page.NextLink == "" {
			return nil
		}

		next, err := sameHostNextPageURL("azure", session.GraphBaseURL(), page.NextLink)
		if err != nil {
			return err
		}

		pageURL = next
	}

	return fmt.Errorf("cannot list all azure sign-in activity: %w", ErrPaginationLimitReached)
}

func batchAzureSignInActivity(
	ctx context.Context,
	session *cloudazure.Session,
	ids []string,
) (map[string]time.Time, error) {
	batchURL, err := azureGraphBatchURL(session.GraphBaseURL())
	if err != nil {
		return nil, err
	}

	found := make(map[string]time.Time, len(ids))

	for start := 0; start < len(ids); start += azureSignInActivityBatchChunk {
		end := min(start+azureSignInActivityBatchChunk, len(ids))
		if err := postAzureSignInActivityBatch(ctx, session, batchURL, ids[start:end], found); err != nil {
			return nil, err
		}
	}

	return found, nil
}

func postAzureSignInActivityBatch(
	ctx context.Context,
	session *cloudazure.Session,
	batchURL string,
	ids []string,
	found map[string]time.Time,
) error {
	requests := make([]azureBatchRequestItem, 0, len(ids))
	for i, id := range ids {
		itemURL, err := azureBatchUserSignInPath(id)
		if err != nil {
			return err
		}

		requests = append(
			requests,
			azureBatchRequestItem{
				ID:     strconv.Itoa(i + 1),
				Method: http.MethodGet,
				URL:    itemURL,
			},
		)
	}

	payload, err := json.Marshal(azureBatchRequest{Requests: requests})
	if err != nil {
		return fmt.Errorf("cannot encode azure sign-in batch request: %w", err)
	}

	var page azureBatchResponse
	if err := postAzureGraphJSON(ctx, session, batchURL, payload, &page); err != nil {
		return err
	}

	for _, item := range page.Responses {
		if item.Status == http.StatusNotFound {
			continue
		}

		if item.Status < 200 || item.Status >= 300 {
			return azureGraphError(item.Status, item.Body)
		}

		var user azureUserSignIn
		if err := json.Unmarshal(item.Body, &user); err != nil {
			return fmt.Errorf("cannot decode azure sign-in batch item: %w", err)
		}

		if at, ok := azureSignInTime(user.SignInActivity); ok {
			found[user.ID] = at
		}
	}

	return nil
}

func applyAzureActivity(
	records []AccountRecord,
	logins map[string]time.Time,
	mfa map[string]coredata.MFAStatus,
) {
	for i := range records {
		if !azureRecordIsUser(records[i]) {
			continue
		}

		if at, ok := logins[records[i].ExternalID]; ok {
			records[i].LastLogin = new(at)
		}

		if status, ok := mfa[records[i].ExternalID]; ok {
			records[i].MFAStatus = status
		}
	}
}

func azureRecordIsUser(record AccountRecord) bool {
	return record.AuthMethod == coredata.AccessReviewEntryAuthMethodSSO
}

func azureUserIDs(records []AccountRecord) []string {
	ids := make([]string, 0, len(records))
	seen := make(map[string]struct{}, len(records))

	for _, record := range records {
		if !azureRecordIsUser(record) || record.ExternalID == "" {
			continue
		}

		if _, ok := seen[record.ExternalID]; ok {
			continue
		}

		seen[record.ExternalID] = struct{}{}
		ids = append(ids, record.ExternalID)
	}

	return ids
}

func azureUserIDSet(records []AccountRecord) map[string]struct{} {
	ids := azureUserIDs(records)

	wanted := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		wanted[id] = struct{}{}
	}

	return wanted
}

func azureSignInTime(activity azureSignInActivity) (time.Time, bool) {
	// lastSuccessfulSignInDateTime is the honest "this account was used"
	// signal, but it was not backfilled. Fall back so a dormant-since-2023
	// account does not look unused.
	if at, ok := azureParseGraphTime(activity.LastSuccessfulSignInDateTime); ok {
		return at, true
	}

	return azureParseGraphTime(activity.LastSignInDateTime)
}

func azureParseGraphTime(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}

	at, err := time.Parse(time.RFC3339, value)
	if err != nil {
		at, err = time.Parse(time.RFC3339Nano, value)
	}

	if err != nil || at.IsZero() {
		return time.Time{}, false
	}

	return at.UTC(), true
}

func azureUsersSignInActivityURL(graphBaseURL string, ids []string) (string, error) {
	endpoint, err := url.JoinPath(graphBaseURL, azureGraphAPIVersion, azureGraphUsersPath)
	if err != nil {
		return "", fmt.Errorf("cannot build azure sign-in activity URL: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("cannot parse azure sign-in activity URL: %w", err)
	}

	q := parsed.Query()
	q.Set("$select", azureSignInActivitySelect)
	q.Set("$top", strconv.Itoa(azureSignInActivityPageSize))
	q.Set("$filter", azureUsersIDFilter(ids))
	parsed.RawQuery = q.Encode()

	return parsed.String(), nil
}

func azureUsersIDFilter(ids []string) string {
	quoted := make([]string, 0, len(ids))
	for _, id := range ids {
		quoted = append(quoted, "'"+strings.ReplaceAll(id, "'", "''")+"'")
	}

	return "id in (" + strings.Join(quoted, ",") + ")"
}

func azureGraphBatchURL(graphBaseURL string) (string, error) {
	endpoint, err := url.JoinPath(graphBaseURL, azureGraphAPIVersion, azureGraphBatchPath)
	if err != nil {
		return "", fmt.Errorf("cannot build azure graph batch URL: %w", err)
	}

	return endpoint, nil
}

func azureBatchUserSignInPath(id string) (string, error) {
	endpoint, err := url.JoinPath("/"+azureGraphUsersPath, url.PathEscape(id))
	if err != nil {
		return "", fmt.Errorf("cannot build azure sign-in batch path: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("cannot parse azure sign-in batch path: %w", err)
	}

	q := parsed.Query()
	q.Set("$select", azureSignInActivitySelect)
	parsed.RawQuery = q.Encode()

	return parsed.String(), nil
}
