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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	istypes "github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/awsx/arn"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// awsSSOAdministratorAccess is the predefined Identity Center permission
	// set that grants the same unrestricted access as AdministratorAccess.
	awsSSOAdministratorAccess = "AWSAdministratorAccess"

	identityStoreService = "identitystore"
)

type (
	// identityCenterUser is one Identity Center principal assigned to the
	// connected AWS account, joined from permission-set assignments and the
	// identity store.
	identityCenterUser struct {
		ID          string
		ARN         string
		UserName    string
		DisplayName string
		Email       string
		Title       string
		Grants      []string
		Admin       bool
		Status      istypes.UserStatus
		CreatedAt   *time.Time
	}

	identityCenterInstance struct {
		arn     string
		storeID string
	}

	identityCenterPermissionSet struct {
		arn   string
		name  string
		admin bool
	}

	identityCenterGrants struct {
		names []string
		admin bool
	}
)

// listIdentityCenterUsers returns Identity Center users assigned to this
// session's account.
//
// SSO Admin degrades: an empty ListInstances, or any non-cancel error talking
// to SSO Admin (discovery or a later read), means this account does not
// expose a usable instance in the session region (a member account, a custom
// role without the needed sso:*, or an instance hosted elsewhere). The IAM
// walk still stands. Identity-store failures fail the fetch.
func listIdentityCenterUsers(
	ctx context.Context,
	session *cloudaws.Session,
	logger *log.Logger,
) ([]identityCenterUser, error) {
	sso := ssoadmin.NewFromConfig(session.Config())

	instance, err := discoverIdentityCenter(ctx, sso)
	if err != nil {
		if ctx.Err() != nil {
			return nil, err
		}

		logger.WarnCtx(
			ctx,
			"cannot discover iam identity center, listing iam identities only",
			log.Error(err),
		)

		return nil, nil
	}

	if instance.arn == "" {
		return nil, nil
	}

	sets, err := listIdentityCenterPermissionSets(ctx, sso, instance.arn)
	if err != nil {
		if ctx.Err() != nil {
			return nil, err
		}

		logger.WarnCtx(
			ctx,
			"cannot read iam identity center after discovery, listing iam identities only",
			log.Error(err),
		)

		return nil, nil
	}

	grants := make(map[string]*identityCenterGrants)
	groups := make(map[string][]identityCenterPermissionSet)

	for _, set := range sets {
		assignments, err := listIdentityCenterAccountAssignments(
			ctx,
			sso,
			instance.arn,
			session.AccountID(),
			set.arn,
		)
		if err != nil {
			if ctx.Err() != nil {
				return nil, err
			}

			logger.WarnCtx(
				ctx,
				"cannot read iam identity center after discovery, listing iam identities only",
				log.Error(err),
			)

			return nil, nil
		}

		for _, assignment := range assignments {
			principalID := aws.ToString(assignment.PrincipalId)
			if principalID == "" {
				continue
			}

			switch assignment.PrincipalType {
			case ssotypes.PrincipalTypeUser:
				addIdentityCenterGrant(grants, principalID, set.name, set.admin)
			case ssotypes.PrincipalTypeGroup:
				groups[principalID] = append(groups[principalID], set)
			}
		}
	}

	store := identitystore.NewFromConfig(session.Config())

	if err := expandIdentityCenterGroupAssignments(ctx, store, instance.storeID, groups, grants); err != nil {
		return nil, err
	}

	if len(grants) == 0 {
		return nil, nil
	}

	users, err := listAssignedIdentityStoreUsers(ctx, store, instance.storeID, grants)
	if err != nil {
		return nil, err
	}

	for i := range users {
		users[i].ARN = identityCenterUserARN(session.Partition(), session.AccountID(), users[i].ID)
	}

	return users, nil
}

func discoverIdentityCenter(
	ctx context.Context,
	client *ssoadmin.Client,
) (identityCenterInstance, error) {
	paginator := ssoadmin.NewListInstancesPaginator(client, &ssoadmin.ListInstancesInput{})

	for range maxPaginationPages {
		if !paginator.HasMorePages() {
			return identityCenterInstance{}, nil
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return identityCenterInstance{}, fmt.Errorf("cannot list iam identity center instances: %w", err)
		}

		for _, instance := range page.Instances {
			instanceARN := aws.ToString(instance.InstanceArn)

			storeID := aws.ToString(instance.IdentityStoreId)
			if instanceARN == "" || storeID == "" {
				continue
			}

			return identityCenterInstance{arn: instanceARN, storeID: storeID}, nil
		}
	}

	return identityCenterInstance{}, fmt.Errorf(
		"cannot list all iam identity center instances: %w",
		ErrPaginationLimitReached,
	)
}

func listIdentityCenterPermissionSets(
	ctx context.Context,
	client *ssoadmin.Client,
	instanceARN string,
) ([]identityCenterPermissionSet, error) {
	paginator := ssoadmin.NewListPermissionSetsPaginator(
		client,
		&ssoadmin.ListPermissionSetsInput{InstanceArn: aws.String(instanceARN)},
	)

	var arns []string

	for range maxPaginationPages {
		if !paginator.HasMorePages() {
			sets := make([]identityCenterPermissionSet, 0, len(arns))
			for _, setARN := range arns {
				set, err := describeIdentityCenterPermissionSet(ctx, client, instanceARN, setARN)
				if err != nil {
					return nil, err
				}

				sets = append(sets, set)
			}

			return sets, nil
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot list iam identity center permission sets: %w", err)
		}

		arns = append(arns, page.PermissionSets...)
	}

	return nil, fmt.Errorf("cannot list all iam identity center permission sets: %w", ErrPaginationLimitReached)
}

func describeIdentityCenterPermissionSet(
	ctx context.Context,
	client *ssoadmin.Client,
	instanceARN string,
	permissionSetARN string,
) (identityCenterPermissionSet, error) {
	described, err := client.DescribePermissionSet(
		ctx,
		&ssoadmin.DescribePermissionSetInput{
			InstanceArn:      aws.String(instanceARN),
			PermissionSetArn: aws.String(permissionSetARN),
		},
	)
	if err != nil {
		return identityCenterPermissionSet{}, fmt.Errorf(
			"cannot describe an iam identity center permission set: %w",
			err,
		)
	}

	name := permissionSetARN

	if described.PermissionSet != nil {
		if describedName := aws.ToString(described.PermissionSet.Name); describedName != "" {
			name = describedName
		}
	}

	admin, err := identityCenterPermissionSetIsAdmin(ctx, client, instanceARN, permissionSetARN, name)
	if err != nil {
		return identityCenterPermissionSet{}, err
	}

	return identityCenterPermissionSet{arn: permissionSetARN, name: name, admin: admin}, nil
}

func identityCenterPermissionSetIsAdmin(
	ctx context.Context,
	client *ssoadmin.Client,
	instanceARN string,
	permissionSetARN string,
	name string,
) (bool, error) {
	if identityCenterAdminName(name) {
		return true, nil
	}

	paginator := ssoadmin.NewListManagedPoliciesInPermissionSetPaginator(
		client,
		&ssoadmin.ListManagedPoliciesInPermissionSetInput{
			InstanceArn:      aws.String(instanceARN),
			PermissionSetArn: aws.String(permissionSetARN),
		},
	)

	for range maxPaginationPages {
		if !paginator.HasMorePages() {
			return false, nil
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return false, fmt.Errorf(
				"cannot list managed policies of an iam identity center permission set: %w",
				err,
			)
		}

		for _, policy := range page.AttachedManagedPolicies {
			if identityCenterAdminName(aws.ToString(policy.Name)) {
				return true, nil
			}
		}
	}

	return false, fmt.Errorf(
		"cannot list all managed policies of an iam identity center permission set: %w",
		ErrPaginationLimitReached,
	)
}

func listIdentityCenterAccountAssignments(
	ctx context.Context,
	client *ssoadmin.Client,
	instanceARN string,
	accountID string,
	permissionSetARN string,
) ([]ssotypes.AccountAssignment, error) {
	paginator := ssoadmin.NewListAccountAssignmentsPaginator(
		client,
		&ssoadmin.ListAccountAssignmentsInput{
			AccountId:        aws.String(accountID),
			InstanceArn:      aws.String(instanceARN),
			PermissionSetArn: aws.String(permissionSetARN),
		},
	)

	var assignments []ssotypes.AccountAssignment

	for range maxPaginationPages {
		if !paginator.HasMorePages() {
			return assignments, nil
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf(
				"cannot list iam identity center account assignments: %w",
				err,
			)
		}

		assignments = append(assignments, page.AccountAssignments...)
	}

	return nil, fmt.Errorf(
		"cannot list all iam identity center account assignments: %w",
		ErrPaginationLimitReached,
	)
}

func expandIdentityCenterGroupAssignments(
	ctx context.Context,
	client *identitystore.Client,
	storeID string,
	groups map[string][]identityCenterPermissionSet,
	grants map[string]*identityCenterGrants,
) error {
	for groupID, sets := range groups {
		name, err := describeIdentityCenterGroup(ctx, client, storeID, groupID)
		if err != nil {
			return err
		}

		members, err := listIdentityCenterGroupMembers(ctx, client, storeID, groupID)
		if err != nil {
			return err
		}

		for _, userID := range members {
			for _, set := range sets {
				addIdentityCenterGrant(grants, userID, set.name, set.admin)
				addIdentityCenterGrant(grants, userID, name, false)
			}
		}
	}

	return nil
}

func describeIdentityCenterGroup(
	ctx context.Context,
	client *identitystore.Client,
	storeID string,
	groupID string,
) (string, error) {
	group, err := client.DescribeGroup(
		ctx,
		&identitystore.DescribeGroupInput{
			IdentityStoreId: aws.String(storeID),
			GroupId:         aws.String(groupID),
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot describe an iam identity center group: %w", err)
	}

	name := aws.ToString(group.DisplayName)
	if name == "" {
		return groupID, nil
	}

	return name, nil
}

func listIdentityCenterGroupMembers(
	ctx context.Context,
	client *identitystore.Client,
	storeID string,
	groupID string,
) ([]string, error) {
	paginator := identitystore.NewListGroupMembershipsPaginator(
		client,
		&identitystore.ListGroupMembershipsInput{
			IdentityStoreId: aws.String(storeID),
			GroupId:         aws.String(groupID),
		},
	)

	var members []string

	for range maxPaginationPages {
		if !paginator.HasMorePages() {
			return members, nil
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot list iam identity center group memberships: %w", err)
		}

		for _, membership := range page.GroupMemberships {
			userID, ok := identityStoreMemberUserID(membership.MemberId)
			if !ok {
				continue
			}

			members = append(members, userID)
		}
	}

	return nil, fmt.Errorf("cannot list all iam identity center group memberships: %w", ErrPaginationLimitReached)
}

func listAssignedIdentityStoreUsers(
	ctx context.Context,
	client *identitystore.Client,
	storeID string,
	grants map[string]*identityCenterGrants,
) ([]identityCenterUser, error) {
	paginator := identitystore.NewListUsersPaginator(
		client,
		&identitystore.ListUsersInput{IdentityStoreId: aws.String(storeID)},
	)

	users := make([]identityCenterUser, 0, len(grants))

	for range maxPaginationPages {
		if !paginator.HasMorePages() {
			return users, nil
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot list iam identity store users: %w", err)
		}

		for _, user := range page.Users {
			userID := aws.ToString(user.UserId)

			grant, ok := grants[userID]
			if !ok {
				continue
			}

			users = append(
				users,
				identityCenterUser{
					ID:          userID,
					UserName:    aws.ToString(user.UserName),
					DisplayName: aws.ToString(user.DisplayName),
					Email:       identityCenterEmail(user.Emails, aws.ToString(user.UserName)),
					Title:       aws.ToString(user.Title),
					Grants:      slices.Compact(slices.Sorted(slices.Values(grant.names))),
					Admin:       grant.admin,
					Status:      user.UserStatus,
					CreatedAt:   user.CreatedAt,
				},
			)
		}
	}

	return nil, fmt.Errorf("cannot list all iam identity store users: %w", ErrPaginationLimitReached)
}

func addIdentityCenterGrant(
	grants map[string]*identityCenterGrants,
	userID string,
	name string,
	admin bool,
) {
	grant, ok := grants[userID]
	if !ok {
		grant = &identityCenterGrants{}
		grants[userID] = grant
	}

	if name != "" {
		grant.names = append(grant.names, name)
	}

	if admin {
		grant.admin = true
	}
}

func identityStoreMemberUserID(member istypes.MemberId) (string, bool) {
	user, ok := member.(*istypes.MemberIdMemberUserId)
	if !ok || user.Value == "" {
		return "", false
	}

	return user.Value, true
}

func identityCenterAdminName(name string) bool {
	return name == awsAdministratorAccess || name == awsSSOAdministratorAccess
}

func identityCenterEmail(emails []istypes.Email, userName string) string {
	var first string

	for _, email := range emails {
		value := aws.ToString(email.Value)
		if value == "" {
			continue
		}

		if email.Primary {
			return value
		}

		if first == "" {
			first = value
		}
	}

	if first != "" {
		return first
	}

	if strings.Contains(userName, "@") {
		return userName
	}

	return ""
}

func identityCenterFullName(user identityCenterUser) string {
	if user.DisplayName != "" {
		return user.DisplayName
	}

	return user.UserName
}

func identityCenterUserARN(partition, accountID, userID string) string {
	return arn.Format(partition, identityStoreService, "", accountID, "user/"+userID)
}

func identityCenterIsAdmin(user identityCenterUser) *bool {
	if user.Admin || slices.ContainsFunc(user.Grants, identityCenterAdminName) {
		return new(true)
	}

	return nil
}

func identityCenterActive(status istypes.UserStatus) *bool {
	switch status {
	case istypes.UserStatusEnabled:
		return new(true)
	case istypes.UserStatusDisabled:
		return new(false)
	default:
		return nil
	}
}

// identityCenterUserRecord maps one Identity Center user assigned to the
// connected account. MFA and last login stay unknown: Identity Store has no
// public API for either. Active follows UserStatus when the store reports it.
func identityCenterUserRecord(user identityCenterUser) AccountRecord {
	return AccountRecord{
		Email:       user.Email,
		FullName:    identityCenterFullName(user),
		JobTitle:    user.Title,
		Roles:       user.Grants,
		IsAdmin:     identityCenterIsAdmin(user),
		Active:      identityCenterActive(user.Status),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodSSO,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		CreatedAt:   user.CreatedAt,
		ExternalID:  user.ARN,
	}
}
