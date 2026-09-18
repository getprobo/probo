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

package aws

import (
	"fmt"

	"go.probo.inc/probo/pkg/awsx/arn"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

// NewMemberSession opens a session on one member account of the organization
// the management role belongs to, by assuming that account's copy of the
// audit role.
//
// The partition comes from the management ARN and never from the literal
// "aws": a GovCloud or China organization's member ARNs are not in the
// commercial partition, and assuming one fails with an STS error naming
// neither the partition nor the account.
//
// No role chaining: this is a second AssumeRoleWithWebIdentity against the
// member's own trust policy, exactly like the standalone path, so a member
// Probo cannot reach fails on its own rather than poisoning the organization
// credential.
//
// memberRoleName is the role the customer's StackSet created in every member
// account. Empty means the name the published template uses.
func NewMemberSession(
	issuer *identityfederation.Issuer,
	organizationID gid.GID,
	managementRoleARN string,
	memberRoleName string,
	accountID string,
) (*Session, error) {
	management, err := arn.ParseRole(managementRoleARN)
	if err != nil {
		return nil, fmt.Errorf("cannot open aws member session: management role ARN is not an IAM role ARN")
	}

	if memberRoleName == "" {
		memberRoleName = coredata.DefaultAWSRoleName
	}

	session, err := NewSession(
		issuer,
		organizationID,
		arn.RoleARN(management.Partition, accountID, memberRoleName),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot open aws member session: %w", err)
	}

	return session, nil
}
