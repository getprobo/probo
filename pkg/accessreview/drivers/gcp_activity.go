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
	"net/url"
	"strings"
	"time"

	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/coredata"
	cloudlogging "google.golang.org/api/logging/v2"
	"google.golang.org/api/option"
	policyanalyzer "google.golang.org/api/policyanalyzer/v1"
)

const (
	// gcpActivityLookback is Cloud Audit Logs Admin Activity retention
	// that this driver will walk. Older project actions are invisible,
	// not absent.
	gcpActivityLookback = 90 * 24 * time.Hour

	// gcpActivityTimeout caps enrichment so a long Logging walk cannot
	// spend the campaign fetch budget and then drop the identity list.
	gcpActivityTimeout = 20 * time.Second

	gcpActivityLogID               = "cloudaudit.googleapis.com%2Factivity"
	gcpActivityTypeSALastAuth      = "serviceAccountLastAuthentication"
	gcpActivityTypeSAKeyLastAuth   = "serviceAccountKeyLastAuthentication"
	gcpLoggingPageSize             = 1000
	gcpPolicyAnalyzerPageSize      = 1000
	gcpLoggingFilterMaxLen         = 18000
	gcpLoggingOrderByNewest        = "timestamp desc"
	gcpServiceAccountsResourceMark = "/serviceAccounts/"
)

type gcpPolicyActivityPayload struct {
	LastAuthenticatedTime string `json:"lastAuthenticatedTime"`
}

type gcpAuditLogPayload struct {
	AuthenticationInfo struct {
		PrincipalEmail string `json:"principalEmail"`
	} `json:"authenticationInfo"`
}

func enrichGCPIdentities(
	ctx context.Context,
	session *cloudgcp.Session,
	records []AccountRecord,
) error {
	if len(records) == 0 {
		return nil
	}

	enrichCtx, cancel := context.WithTimeout(ctx, gcpActivityTimeout)
	defer cancel()

	logins, usedKey, activityErr := fetchGCPActivity(enrichCtx, session, records)
	if errors.Is(activityErr, context.Canceled) && ctx.Err() != nil {
		return activityErr
	}

	mfa, mfaErr := fetchGCPMFA(enrichCtx, session, records)
	if errors.Is(mfaErr, context.Canceled) && ctx.Err() != nil {
		return mfaErr
	}

	applyGCPActivity(records, logins, usedKey, mfa)

	if activityErr != nil {
		return activityErr
	}

	return mfaErr
}

func fetchGCPActivity(
	ctx context.Context,
	session *cloudgcp.Session,
	records []AccountRecord,
) (map[string]time.Time, map[string]bool, error) {
	saLogins, usedKey, paErr := queryGCPPolicyAnalyzer(ctx, session)
	if paErr != nil && ctx.Err() != nil {
		return nil, nil, paErr
	}

	calEmails := gcpActivityEmails(records, gcpPrincipalUser)
	if paErr != nil {
		calEmails = append(calEmails, gcpActivityEmails(records, gcpPrincipalServiceAccount)...)
	}

	calLogins, calErr := queryGCPAdminActivity(ctx, session, calEmails)
	if calErr != nil && ctx.Err() != nil {
		return saLogins, usedKey, calErr
	}

	logins := make(map[string]time.Time, len(saLogins)+len(calLogins))
	mapsCopyTimes(logins, saLogins)
	mapsCopyTimes(logins, calLogins)

	if paErr != nil {
		if calErr != nil {
			return logins, usedKey, paErr
		}

		return logins, usedKey, nil
	}

	return logins, usedKey, calErr
}

func queryGCPPolicyAnalyzer(
	ctx context.Context,
	session *cloudgcp.Session,
) (map[string]time.Time, map[string]bool, error) {
	svc, err := policyanalyzer.NewService(
		ctx,
		option.WithHTTPClient(session.HTTPClient()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create gcp policy analyzer client: %w", err)
	}

	lastAuthParent, err := gcpActivityTypeParent(session.AccountID(), gcpActivityTypeSALastAuth)
	if err != nil {
		return nil, nil, err
	}

	logins, err := queryGCPPolicyActivities(ctx, svc, lastAuthParent)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot query gcp service account last authentication: %w", err)
	}

	keyParent, err := gcpActivityTypeParent(session.AccountID(), gcpActivityTypeSAKeyLastAuth)
	if err != nil {
		return logins, nil, err
	}

	keyLogins, keyErr := queryGCPPolicyActivities(ctx, svc, keyParent)
	if keyErr != nil {
		if ctx.Err() != nil {
			return logins, nil, keyErr
		}

		if !cloudgcp.As[cloudgcp.ErrPermissionDenied](keyErr) {
			return logins, nil, fmt.Errorf("cannot query gcp service account key last authentication: %w", keyErr)
		}

		return logins, nil, nil
	}

	usedKey := make(map[string]bool, len(keyLogins))
	for id := range keyLogins {
		usedKey[id] = true
		usedKey[strings.ToLower(id)] = true
	}

	return logins, usedKey, nil
}

func queryGCPPolicyActivities(
	ctx context.Context,
	svc *policyanalyzer.Service,
	parent string,
) (map[string]time.Time, error) {
	found := make(map[string]time.Time)
	pageToken := ""

	for range maxPaginationPages {
		call := svc.Projects.Locations.ActivityTypes.Activities.Query(parent).
			PageSize(gcpPolicyAnalyzerPageSize).
			Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return found, err
		}

		for _, activity := range resp.Activities {
			id, at, ok := parseGCPPolicyActivity(activity)
			if !ok {
				continue
			}

			foldGCPActivityTimestamp(found, id, at)
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			return found, nil
		}
	}

	return found, fmt.Errorf("cannot query gcp policy analyzer activities: %w", ErrPaginationLimitReached)
}

func parseGCPPolicyActivity(activity *policyanalyzer.GoogleCloudPolicyanalyzerV1Activity) (string, time.Time, bool) {
	if activity == nil {
		return "", time.Time{}, false
	}

	id, ok := gcpServiceAccountIDFromResource(activity.FullResourceName)
	if !ok {
		return "", time.Time{}, false
	}

	var payload gcpPolicyActivityPayload

	if len(activity.Activity) == 0 {
		return "", time.Time{}, false
	}

	if err := json.Unmarshal(activity.Activity, &payload); err != nil {
		return "", time.Time{}, false
	}

	at, ok := parseGCPTimestamp(payload.LastAuthenticatedTime)
	if !ok {
		return "", time.Time{}, false
	}

	return id, at, true
}

func gcpServiceAccountIDFromResource(name string) (string, bool) {
	_, rest, ok := strings.Cut(name, gcpServiceAccountsResourceMark)
	if !ok {
		return "", false
	}

	id, _, _ := strings.Cut(rest, "/")

	return id, id != ""
}

func queryGCPAdminActivity(
	ctx context.Context,
	session *cloudgcp.Session,
	emails []string,
) (map[string]time.Time, error) {
	if len(emails) == 0 {
		return nil, nil
	}

	svc, err := cloudlogging.NewService(
		ctx,
		option.WithHTTPClient(session.HTTPClient()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create gcp logging client: %w", err)
	}

	resource, err := url.JoinPath("projects", url.PathEscape(session.AccountID()))
	if err != nil {
		return nil, fmt.Errorf("cannot build gcp logging resource name: %w", err)
	}

	since := time.Now().UTC().Add(-gcpActivityLookback)

	wanted := make(map[string]struct{}, len(emails))
	for _, email := range emails {
		wanted[strings.ToLower(email)] = struct{}{}
	}

	found := make(map[string]time.Time, len(emails))

	for _, filter := range gcpEmailFilterBatches(session.AccountID(), since, emails, gcpLoggingFilterMaxLen) {
		if err := listGCPAdminActivity(
			ctx,
			svc,
			resource,
			filter,
			wanted,
			found,
		); err != nil {
			return found, err
		}

		if len(found) == len(wanted) {
			return found, nil
		}
	}

	return found, nil
}

func listGCPAdminActivity(
	ctx context.Context,
	svc *cloudlogging.Service,
	resource string,
	filter string,
	wanted map[string]struct{},
	found map[string]time.Time,
) error {
	pageToken := ""

	for range maxPaginationPages {
		req := &cloudlogging.ListLogEntriesRequest{
			ResourceNames: []string{resource},
			Filter:        filter,
			OrderBy:       gcpLoggingOrderByNewest,
			PageSize:      gcpLoggingPageSize,
			PageToken:     pageToken,
		}

		resp, err := svc.Entries.List(req).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("cannot list gcp admin activity: %w", err)
		}

		for _, entry := range resp.Entries {
			email, at, ok := parseGCPAdminActivityEntry(entry)
			if !ok {
				continue
			}

			if _, want := wanted[email]; !want {
				continue
			}

			foldGCPActivityTimestamp(found, email, at)

			if len(found) == len(wanted) {
				return nil
			}
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			return nil
		}
	}

	return fmt.Errorf("cannot list all gcp admin activity: %w", ErrPaginationLimitReached)
}

func parseGCPAdminActivityEntry(entry *cloudlogging.LogEntry) (string, time.Time, bool) {
	if entry == nil {
		return "", time.Time{}, false
	}

	at, ok := parseGCPTimestamp(entry.Timestamp)
	if !ok {
		return "", time.Time{}, false
	}

	var payload gcpAuditLogPayload

	if len(entry.ProtoPayload) == 0 {
		return "", time.Time{}, false
	}

	if err := json.Unmarshal(entry.ProtoPayload, &payload); err != nil {
		return "", time.Time{}, false
	}

	email := strings.ToLower(strings.TrimSpace(payload.AuthenticationInfo.PrincipalEmail))
	if email == "" {
		return "", time.Time{}, false
	}

	return email, at, true
}

func applyGCPActivity(
	records []AccountRecord,
	logins map[string]time.Time,
	usedKey map[string]bool,
	mfa map[string]coredata.MFAStatus,
) {
	for i := range records {
		kind := gcpRecordKind(records[i])
		if kind == gcpPrincipalGroup {
			continue
		}

		if at, ok := gcpActivityLookupTime(logins, records[i]); ok {
			records[i].LastLogin = new(at)
		}

		if kind == gcpPrincipalServiceAccount &&
			records[i].AuthMethod == coredata.AccessReviewEntryAuthMethodUnknown &&
			gcpActivityLookupBool(usedKey, records[i]) {
			records[i].AuthMethod = coredata.AccessReviewEntryAuthMethodAPIKey
		}

		if kind != gcpPrincipalUser {
			continue
		}

		if status, ok := mfa[strings.ToLower(records[i].Email)]; ok {
			records[i].MFAStatus = status
		}
	}
}

func gcpRecordKind(rec AccountRecord) gcpPrincipalKind {
	if rec.AccountType == coredata.AccessReviewEntryAccountTypeServiceAccount {
		return gcpPrincipalServiceAccount
	}

	if strings.HasPrefix(rec.ExternalID, gcpPrincipalGroupPrefix) {
		return gcpPrincipalGroup
	}

	return gcpPrincipalUser
}

func gcpActivityEmails(records []AccountRecord, kind gcpPrincipalKind) []string {
	emails := make([]string, 0, len(records))
	seen := make(map[string]struct{}, len(records))

	for _, rec := range records {
		if gcpRecordKind(rec) != kind || rec.Email == "" {
			continue
		}

		key := strings.ToLower(rec.Email)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}

		emails = append(emails, rec.Email)
	}

	return emails
}

func gcpActivityLookupTime(logins map[string]time.Time, rec AccountRecord) (time.Time, bool) {
	if at, ok := logins[strings.ToLower(rec.Email)]; ok {
		return at, true
	}

	if rec.ExternalID != "" {
		if at, ok := logins[rec.ExternalID]; ok {
			return at, true
		}

		if at, ok := logins[strings.ToLower(rec.ExternalID)]; ok {
			return at, true
		}
	}

	return time.Time{}, false
}

func gcpActivityLookupBool(flags map[string]bool, rec AccountRecord) bool {
	if flags[strings.ToLower(rec.Email)] {
		return true
	}

	if rec.ExternalID != "" && (flags[rec.ExternalID] || flags[strings.ToLower(rec.ExternalID)]) {
		return true
	}

	return false
}

func foldGCPActivityTimestamp(found map[string]time.Time, id string, at time.Time) {
	key := strings.ToLower(id)
	if prev, ok := found[key]; ok && !at.After(prev) {
		return
	}

	found[key] = at
}

func gcpEmailFilterBatches(project string, since time.Time, emails []string, maxLen int) []string {
	if len(emails) == 0 {
		return nil
	}

	var (
		batches []string
		batch   []string
	)

	for _, email := range emails {
		trial := append(append([]string{}, batch...), email)

		filter := gcpAdminActivityFilter(project, since, trial)
		if len(batch) > 0 && len(filter) > maxLen {
			batches = append(batches, gcpAdminActivityFilter(project, since, batch))
			batch = []string{email}

			continue
		}

		batch = trial
	}

	if len(batch) > 0 {
		batches = append(batches, gcpAdminActivityFilter(project, since, batch))
	}

	return batches
}

func gcpAdminActivityFilter(project string, since time.Time, emails []string) string {
	var b strings.Builder

	b.WriteString("logName=")
	b.WriteString(gcpLoggingQuote(gcpActivityLogName(project)))
	b.WriteString(" AND timestamp>=")
	b.WriteString(gcpLoggingQuote(since.UTC().Format(time.RFC3339)))

	if len(emails) == 0 {
		return b.String()
	}

	b.WriteString(" AND (")

	for i, email := range emails {
		if i > 0 {
			b.WriteString(" OR ")
		}

		b.WriteString("protoPayload.authenticationInfo.principalEmail=")
		b.WriteString(gcpLoggingQuote(email))
	}

	b.WriteByte(')')

	return b.String()
}

func gcpActivityLogName(project string) string {
	return "projects/" + url.PathEscape(project) + "/logs/" + gcpActivityLogID
}

func gcpActivityTypeParent(project string, activityType string) (string, error) {
	parent, err := url.JoinPath(
		"projects",
		url.PathEscape(project),
		"locations",
		"global",
		"activityTypes",
		activityType,
	)
	if err != nil {
		return "", fmt.Errorf("cannot build gcp policy analyzer parent: %w", err)
	}

	return parent, nil
}

func gcpLoggingQuote(value string) string {
	var b strings.Builder

	b.WriteByte('"')

	for _, r := range value {
		if r == '\\' || r == '"' {
			b.WriteByte('\\')
		}

		b.WriteRune(r)
	}

	b.WriteByte('"')

	return b.String()
}

func parseGCPTimestamp(raw string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}

	if at, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return at.UTC(), true
	}

	if at, err := time.Parse(time.RFC3339, raw); err == nil {
		return at.UTC(), true
	}

	return time.Time{}, false
}

func mapsCopyTimes(dst map[string]time.Time, src map[string]time.Time) {
	for key, at := range src {
		foldGCPActivityTimestamp(dst, key, at)
	}
}
