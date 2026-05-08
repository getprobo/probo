// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package drivers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// AWSIAMReader is the narrow seam CloudAWSDriver depends on for
	// IAM reads. The real *iam.Client (returned by
	// *cloudaccount.AWSProvider.IAM()) satisfies it implicitly;
	// tests inject a stub. Exported so the closure factory in
	// pkg/probo can wire concrete seam adapters at construction
	// time without forcing pkg/accessreview to import
	// pkg/cloudaccount.
	AWSIAMReader interface {
		ListUsers(ctx context.Context, in *iam.ListUsersInput, opts ...func(*iam.Options)) (*iam.ListUsersOutput, error)
		ListMFADevices(ctx context.Context, in *iam.ListMFADevicesInput, opts ...func(*iam.Options)) (*iam.ListMFADevicesOutput, error)
		GenerateCredentialReport(ctx context.Context, in *iam.GenerateCredentialReportInput, opts ...func(*iam.Options)) (*iam.GenerateCredentialReportOutput, error)
		GetCredentialReport(ctx context.Context, in *iam.GetCredentialReportInput, opts ...func(*iam.Options)) (*iam.GetCredentialReportOutput, error)
	}

	// CloudAWSDriver fetches IAM users from a customer's AWS
	// account via the typed seam interface above. The driver
	// stores only the narrow seam, NOT the concrete provider --
	// pkg/accessreview must remain SDK-agnostic, with the AWS SDK
	// pulled in only via the iam input/output types it needs.
	CloudAWSDriver struct {
		iam AWSIAMReader
	}
)

// NewCloudAWSDriver wires the driver to the supplied IAM seam.
// Constructor is called from pkg/probo's CloudAccountDriverFactory
// closure; the engine never imports pkg/cloudaccount directly.
func NewCloudAWSDriver(iamReader AWSIAMReader) *CloudAWSDriver {
	return &CloudAWSDriver{iam: iamReader}
}

// ListAccounts emits one AccountRecord per IAM user. v1 lists users
// only and decorates with MFA presence; the credential-report path
// is reserved for a follow-up enrichment pass when the customer
// opts into the longer-running probe.
func (d *CloudAWSDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var (
		records []AccountRecord
		marker  *string
	)

	for range maxPaginationPages {
		out, err := d.iam.ListUsers(ctx, &iam.ListUsersInput{Marker: marker})
		if err != nil {
			return nil, fmt.Errorf("cannot list aws iam users: %w", err)
		}

		for _, user := range out.Users {
			record := AccountRecord{
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
				CreatedAt:   user.CreateDate,
			}

			if user.UserId != nil {
				record.ExternalID = *user.UserId
			}
			if user.UserName != nil {
				record.FullName = *user.UserName
				record.Email = *user.UserName
			}

			if user.UserName != nil {
				mfaOut, err := d.iam.ListMFADevices(ctx, &iam.ListMFADevicesInput{UserName: user.UserName})
				if err != nil {
					return nil, fmt.Errorf("cannot list aws iam mfa devices: %w", err)
				}
				if len(mfaOut.MFADevices) > 0 {
					record.MFAStatus = coredata.MFAStatusEnabled
				} else {
					record.MFAStatus = coredata.MFAStatusDisabled
				}
			}

			records = append(records, record)
		}

		if !out.IsTruncated || out.Marker == nil {
			return records, nil
		}
		marker = out.Marker
	}

	return nil, ErrPaginationLimitReached
}
