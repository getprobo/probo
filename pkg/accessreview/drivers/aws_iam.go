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
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/awsx/arn"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
)

const (
	// credentialReportWait bounds how long a caller waits for IAM to build a
	// credential report. The report is a per-account snapshot IAM regenerates
	// on request, and a caller runs inside an access-review fetch budget, so
	// waiting is capped and the credential signals degrade to unknown rather
	// than spending the whole budget here.
	credentialReportWait = 5 * time.Second

	// credentialReportPoll is how often the report is re-requested while it
	// builds. IAM typically has one within a few seconds.
	credentialReportPoll = 500 * time.Millisecond

	// rootUser is the credential report's row for the account root. It is not
	// an IAM user and has no ARN-addressable identity, so it is reported under
	// its own name rather than dropped: root having an active access key is
	// exactly what a reviewer needs to see.
	rootUser = "<root_account>"
)

type (
	// iamUser is one IAM identity of a single AWS account, joined from the
	// account's authorization details and its credential report.
	iamUser struct {
		ARN              string
		Name             string
		Groups           []string
		AttachedPolicies []string
		InlinePolicies   []string
		CreatedAt        *time.Time

		// MFAEnabled, ConsoleAccess and AccessKeyActive are three-valued: nil
		// means the credential report was unavailable, so the driver reports no
		// signal rather than a fabricated negative.
		MFAEnabled      *bool
		ConsoleAccess   *bool
		AccessKeyActive *bool
		LastActiveAt    *time.Time
	}

	// credentialReportRow is the subset of the credential report CSV this
	// driver reads. The report carries far more columns; the ones left out are
	// certificate and key-rotation details no reviewer field maps to.
	credentialReportRow struct {
		mfaActive        bool
		passwordEnabled  bool
		passwordLastUsed *time.Time
		accessKeyActive  bool
		accessKeyLastUse *time.Time
	}
)

// listIAMUsers returns every IAM identity of this session's account.
//
// Two calls per account, not five per user: iam:GetAccountAuthorizationDetails
// returns every user with its groups and policies in one paginated walk, and
// iam:GetCredentialReport returns MFA and credential activity for all of them
// in one CSV. Asking per user instead would multiply the call count by the user
// count and outgrow the access-review fetch budget on any real organization.
func listIAMUsers(ctx context.Context, session *cloudaws.Session, logger *log.Logger) ([]iamUser, error) {
	client := iam.NewFromConfig(session.Config())

	users, err := listAuthorizationDetails(ctx, client)
	if err != nil {
		return nil, err
	}

	report, err := credentialReport(ctx, client)
	if err != nil {
		// A canceled fetch must not commit a partial list. The report's own
		// 5s wait is a child timeout, so a report still building degrades
		// rather than failing the account.
		if ctx.Err() != nil {
			return nil, err
		}

		logger.WarnCtx(
			ctx,
			"cannot read aws credential report, reporting credential signals unknown",
			log.Error(err),
		)
	}

	for i := range users {
		row, ok := report[users[i].Name]
		if !ok {
			continue
		}

		users[i].MFAEnabled = aws.Bool(row.mfaActive)
		users[i].ConsoleAccess = aws.Bool(row.passwordEnabled)
		users[i].AccessKeyActive = aws.Bool(row.accessKeyActive)
		users[i].LastActiveAt = latest(row.passwordLastUsed, row.accessKeyLastUse)
	}

	// Root is not an IAM user. Listing it from the session keeps the most
	// privileged identity visible when the report is missing.
	root, ok := rootUserFromReport(session.Partition(), session.AccountID(), report)
	if !ok {
		root = rootIdentity(session.Partition(), session.AccountID())
	}

	return append(users, root), nil
}

// listAuthorizationDetails walks iam:GetAccountAuthorizationDetails, filtered
// to users.
//
// Group grants are resolved from the same response: a user's group membership
// carries no policies, so the policies attached to those groups are collected
// from the response's group list rather than by asking per group.
func listAuthorizationDetails(ctx context.Context, client *iam.Client) ([]iamUser, error) {
	paginator := iam.NewGetAccountAuthorizationDetailsPaginator(
		client,
		&iam.GetAccountAuthorizationDetailsInput{
			Filter: []iamtypes.EntityType{
				iamtypes.EntityTypeUser,
				iamtypes.EntityTypeGroup,
			},
		},
	)

	var (
		details []iamtypes.UserDetail
		groups  = make(map[string][]string)
	)

	for range maxPaginationPages {
		if !paginator.HasMorePages() {
			return joinGroupPolicies(details, groups), nil
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot read the authorization details of an aws account: %w", err)
		}

		details = append(details, page.UserDetailList...)

		for _, group := range page.GroupDetailList {
			name := aws.ToString(group.GroupName)
			if name == "" {
				continue
			}

			for _, policy := range group.AttachedManagedPolicies {
				groups[name] = append(groups[name], aws.ToString(policy.PolicyName))
			}

			for _, policy := range group.GroupPolicyList {
				groups[name] = append(groups[name], aws.ToString(policy.PolicyName))
			}
		}
	}

	return nil, fmt.Errorf("cannot read all the authorization details of an aws account: %w", ErrPaginationLimitReached)
}

// joinGroupPolicies turns the SDK's user details into iamUsers, folding each
// user's group policies in beside its own.
func joinGroupPolicies(
	details []iamtypes.UserDetail,
	groupPolicies map[string][]string,
) []iamUser {
	users := make([]iamUser, 0, len(details))

	for _, detail := range details {
		userARN := aws.ToString(detail.Arn)
		if userARN == "" {
			continue
		}

		user := iamUser{
			ARN:       userARN,
			Name:      aws.ToString(detail.UserName),
			Groups:    detail.GroupList,
			CreatedAt: detail.CreateDate,
		}

		for _, policy := range detail.AttachedManagedPolicies {
			user.AttachedPolicies = append(user.AttachedPolicies, aws.ToString(policy.PolicyName))
		}

		for _, policy := range detail.UserPolicyList {
			user.InlinePolicies = append(user.InlinePolicies, aws.ToString(policy.PolicyName))
		}

		// A policy reached through a group is as real a grant as one attached
		// directly, and a reviewer cannot see it from the group name alone.
		for _, group := range detail.GroupList {
			user.AttachedPolicies = append(user.AttachedPolicies, groupPolicies[group]...)
		}

		user.AttachedPolicies = slices.Compact(slices.Sorted(slices.Values(user.AttachedPolicies)))

		users = append(users, user)
	}

	return users
}

// credentialReport requests the account's credential report and parses it.
//
// IAM builds the report asynchronously and answers
// CredentialReportNotReadyException until it is done, so the request is retried
// within a bounded wait. Exhausting that wait is not an error worth failing the
// account over — the caller treats a missing report as "no credential signal".
func credentialReport(ctx context.Context, client *iam.Client) (map[string]credentialReportRow, error) {
	ctx, cancel := context.WithTimeout(ctx, credentialReportWait)
	defer cancel()

	if _, err := client.GenerateCredentialReport(ctx, &iam.GenerateCredentialReportInput{}); err != nil {
		if !credentialReportAlreadyGenerated(err) {
			return nil, fmt.Errorf("cannot generate the credential report of an aws account: %w", err)
		}
	}

	for {
		report, err := client.GetCredentialReport(ctx, &iam.GetCredentialReportInput{})
		if err == nil {
			return parseCredentialReport(report.Content)
		}

		if !credentialReportBuilding(err) {
			return nil, fmt.Errorf("cannot read the credential report of an aws account: %w", err)
		}

		timer := time.NewTimer(credentialReportPoll)

		select {
		case <-ctx.Done():
			timer.Stop()

			return nil, fmt.Errorf("cannot read the credential report of an aws account: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

// credentialReportAlreadyGenerated reports whether GenerateCredentialReport
// refused because IAM already holds a recent report, or is already generating
// one. Either way GetCredentialReport is the next call.
func credentialReportAlreadyGenerated(err error) bool {
	if _, ok := errors.AsType[*iamtypes.LimitExceededException](err); ok {
		return true
	}

	_, ok := errors.AsType[*iamtypes.ReportGenerationLimitExceededException](err)

	return ok
}

// credentialReportBuilding reports whether IAM is still producing the report.
//
// Both answers are transient immediately after GenerateCredentialReport: "in
// progress" is the documented one, and "not present" is what a report that has
// not materialised yet looks like. Waiting on either is what the bounded poll is
// for; anything else is a real failure.
func credentialReportBuilding(err error) bool {
	if _, ok := errors.AsType[*iamtypes.CredentialReportNotReadyException](err); ok {
		return true
	}

	_, ok := errors.AsType[*iamtypes.CredentialReportNotPresentException](err)

	return ok
}

// parseCredentialReport reads the report CSV, keyed by user name.
//
// Columns are located by header name rather than by position: AWS has appended
// columns to this report before, and a positional read would silently shift.
func parseCredentialReport(content []byte) (map[string]credentialReportRow, error) {
	reader := csv.NewReader(strings.NewReader(string(content)))

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("cannot read the credential report header: %w", err)
	}

	index := make(map[string]int, len(header))
	for i, name := range header {
		index[strings.TrimSpace(name)] = i
	}

	rows := make(map[string]credentialReportRow)

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return rows, nil
		}

		if err != nil {
			return nil, fmt.Errorf("cannot read a credential report row: %w", err)
		}

		column := func(name string) string {
			i, ok := index[name]
			if !ok || i >= len(record) {
				return ""
			}

			return strings.TrimSpace(record[i])
		}

		user := column("user")
		if user == "" {
			continue
		}

		accessKeyActive := column("access_key_1_active") == "true" ||
			column("access_key_2_active") == "true"

		rows[user] = credentialReportRow{
			mfaActive:        column("mfa_active") == "true",
			passwordEnabled:  column("password_enabled") == "true",
			passwordLastUsed: parseCredentialReportTime(column("password_last_used")),
			accessKeyActive:  accessKeyActive,
			accessKeyLastUse: latest(
				parseCredentialReportTime(column("access_key_1_last_used_date")),
				parseCredentialReportTime(column("access_key_2_last_used_date")),
			),
		}
	}
}

// rootIdentity is the account root without credential signals. The root is
// not an IAM user, so it is listed from the session even when the report
// is missing.
func rootIdentity(partition string, accountID string) iamUser {
	return iamUser{
		ARN:  arn.IAM(partition, accountID, "root"),
		Name: rootUser,
	}
}

// rootUserFromReport builds the account root's entry from the credential
// report.
//
// The root is not an IAM user, so GetAccountAuthorizationDetails never lists it
// and it would otherwise be invisible to a review. It is exactly the identity a
// reviewer most wants to see, because a root account with an active access key
// or without MFA is the finding that matters most. Its ARN is the account's
// canonical root in the session's partition, so it needs no synthetic
// identifier.
func rootUserFromReport(
	partition string,
	accountID string,
	report map[string]credentialReportRow,
) (iamUser, bool) {
	row, ok := report[rootUser]
	if !ok {
		return iamUser{}, false
	}

	root := rootIdentity(partition, accountID)
	root.MFAEnabled = aws.Bool(row.mfaActive)
	root.ConsoleAccess = aws.Bool(row.passwordEnabled)
	root.AccessKeyActive = aws.Bool(row.accessKeyActive)
	root.LastActiveAt = latest(row.passwordLastUsed, row.accessKeyLastUse)

	return root, true
}

// parseCredentialReportTime reads a report timestamp. The report writes
// "N/A" and "not_supported" for a credential that was never used, both of
// which mean "no timestamp" rather than a parse failure.
func parseCredentialReportTime(value string) *time.Time {
	if value == "" || value == "N/A" || value == "not_supported" {
		return nil
	}

	at, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}

	return &at
}

func latest(a, b *time.Time) *time.Time {
	if a == nil {
		return b
	}

	if b == nil {
		return a
	}

	if b.After(*a) {
		return b
	}

	return a
}
