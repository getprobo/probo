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
	"slices"
	"strings"

	"go.gearno.de/kit/log"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// awsAdministratorAccess is the AWS-managed policy that grants unrestricted
	// access. Its presence is an explicit admin signal; its absence proves
	// nothing, since a custom policy can grant just as much.
	awsAdministratorAccess = "AdministratorAccess"
)

// AWSDriver lists the identities of the one AWS account the connector
// names: IAM users, plus Identity Center users assigned to that account
// when an instance is visible in the session region. One connector
// produces one source, the same as GitHub, so roles need no account
// qualification. Member-account connectors and instances hosted in
// another region degrade to IAM only.
type AWSDriver struct {
	session *cloudaws.Session
	logger  *log.Logger
}

var _ Driver = (*AWSDriver)(nil)

// NewAWSDriver builds the driver over a session already assumed on the
// connected account.
func NewAWSDriver(session *cloudaws.Session, logger *log.Logger) *AWSDriver {
	return &AWSDriver{session: session, logger: logger}
}

func (d *AWSDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	users, err := listIAMUsers(ctx, d.session, d.logger)
	if err != nil {
		return nil, fmt.Errorf("cannot list iam identities of the aws account: %w", err)
	}

	records := make([]AccountRecord, 0, len(users))
	for _, user := range users {
		records = append(records, iamUserRecord(user))
	}

	icUsers, err := listIdentityCenterUsers(ctx, d.session, d.logger)
	if err != nil {
		return nil, fmt.Errorf("cannot list identity center identities of the aws account: %w", err)
	}

	for _, user := range icUsers {
		records = append(records, identityCenterUserRecord(user))
	}

	return records, nil
}

// iamUserRecord maps one IAM user.
//
// Email stays empty: an IAM user has no email attribute, and inventing one from
// the user name would create an identity that matches the wrong human during
// reconciliation. Active follows the credential report (console password or an
// access key), not an IAM enabled/disabled flag — IAM has none.
func iamUserRecord(user iamUser) AccountRecord {
	grants := slices.Concat(user.Groups, user.AttachedPolicies, user.InlinePolicies)

	return AccountRecord{
		FullName:    user.Name,
		Roles:       grants,
		IsAdmin:     awsIsAdmin(user, grants),
		Active:      awsActive(user),
		MFAStatus:   awsMFAStatus(user.MFAEnabled),
		AuthMethod:  iamAuthMethod(user),
		AccountType: iamAccountType(user),
		LastLogin:   user.LastActiveAt,
		CreatedAt:   user.CreatedAt,
		ExternalID:  user.ARN,
	}
}

// awsActive reports whether the identity can still sign in. IAM has no
// enabled or disabled flag; a console password or an active access key is the
// usable-account signal. Both fields nil means the credential report was
// missing, so there is no signal.
func awsActive(user iamUser) *bool {
	if user.ConsoleAccess == nil && user.AccessKeyActive == nil {
		return nil
	}

	console := user.ConsoleAccess != nil && *user.ConsoleAccess
	keys := user.AccessKeyActive != nil && *user.AccessKeyActive

	return new(console || keys)
}

// awsIsAdmin reports admin when a grant says so outright, or when the identity
// is the account root. Root is unconditionally privileged and is not an IAM
// principal, so it never carries AdministratorAccess.
//
// It returns nil rather than false when nothing matches, because a custom policy
// can grant everything AdministratorAccess does under any name. Reporting false
// would tell a reviewer this identity is not privileged, which the evidence does
// not support.
func awsIsAdmin(user iamUser, grants []string) *bool {
	if awsIsRoot(user) || slices.Contains(grants, awsAdministratorAccess) {
		return new(true)
	}

	return nil
}

// awsIsRoot reports the account root. Its name is the credential-report
// sentinel, and its ARN resource is "root" (not "user/root").
func awsIsRoot(user iamUser) bool {
	return user.Name == rootUser || strings.HasSuffix(user.ARN, ":root")
}

func awsMFAStatus(enabled *bool) coredata.MFAStatus {
	if enabled == nil {
		return coredata.MFAStatusUnknown
	}

	if *enabled {
		return coredata.MFAStatusEnabled
	}

	return coredata.MFAStatusDisabled
}

// iamAuthMethod reports how the user actually signs in. A console password is
// the human path and takes precedence; an access key alone is a programmatic
// identity.
func iamAuthMethod(user iamUser) coredata.AccessReviewEntryAuthMethod {
	switch {
	case user.ConsoleAccess != nil && *user.ConsoleAccess:
		return coredata.AccessReviewEntryAuthMethodPassword
	case user.AccessKeyActive != nil && *user.AccessKeyActive:
		return coredata.AccessReviewEntryAuthMethodAPIKey
	default:
		return coredata.AccessReviewEntryAuthMethodUnknown
	}
}

// iamAccountType classifies a user with keys but no console password as a
// service account. That combination is what an application credential looks
// like, and it is the only signal IAM offers — the enum has no unknown member,
// so everything else is reported as a user.
func iamAccountType(user iamUser) coredata.AccessReviewEntryAccountType {
	keysOnly := user.AccessKeyActive != nil && *user.AccessKeyActive &&
		user.ConsoleAccess != nil && !*user.ConsoleAccess

	if keysOnly {
		return coredata.AccessReviewEntryAccountTypeServiceAccount
	}

	return coredata.AccessReviewEntryAccountTypeUser
}
