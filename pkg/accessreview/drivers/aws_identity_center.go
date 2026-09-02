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
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	istypes "github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/aws/smithy-go"
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
	// identityCenterUser is one Identity Center principal from the identity
	// store. Grants are the permission sets and groups assigned to this
	// session's account; they are empty when the user has no assignment here.
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

		// MFAEnabled and LastLogin are filled after listing, from registered
		// MFA devices and CloudTrail UserAuthentication. Nil is no signal.
		MFAEnabled *bool
		LastLogin  *time.Time
	}

	identityCenterInstance struct {
		arn     string
		storeID string
		region  string
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

// identityCenterCommercialRegions is every commercial Region where an
// organization can host IAM Identity Center, default-enabled homes first.
// https://docs.aws.amazon.com/singlesignon/latest/userguide/regions.html
var (
	identityCenterCommercialRegions = []string{
		"us-east-1",
		"us-east-2",
		"us-west-2",
		"eu-west-1",
		"eu-central-1",
		"us-west-1",
		"eu-west-2",
		"eu-west-3",
		"eu-north-1",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-northeast-3",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-south-1",
		"sa-east-1",
		"ca-central-1",
		"af-south-1",
		"ap-east-1",
		"ap-east-2",
		"ap-south-2",
		"ap-southeast-3",
		"ap-southeast-4",
		"ap-southeast-5",
		"ap-southeast-6",
		"ap-southeast-7",
		"ca-west-1",
		"eu-south-1",
		"eu-south-2",
		"eu-central-2",
		"il-central-1",
		"me-south-1",
		"me-central-1",
		"mx-central-1",
	}

	identityCenterGovRegions = []string{
		"us-gov-west-1",
		"us-gov-east-1",
	}
)

// identityCenterRegions is the ListInstances search order: the session
// region first, then the rest of the partition's Identity Center regions.
func identityCenterRegions(partition, preferred string) []string {
	var rest []string

	switch partition {
	case cloudaws.GovPartition:
		rest = identityCenterGovRegions
	case cloudaws.ChinaPartition:
		if preferred == "" {
			return nil
		}

		return []string{preferred}
	default:
		rest = identityCenterCommercialRegions
	}

	regions := make([]string, 0, len(rest)+1)
	if preferred != "" {
		regions = append(regions, preferred)
	}

	for _, region := range rest {
		if region == preferred {
			continue
		}

		regions = append(regions, region)
	}

	return regions
}

func identityCenterAccessDenied(err error) bool {
	apiErr, ok := errors.AsType[smithy.APIError](err)
	if !ok || apiErr == nil {
		return false
	}

	switch apiErr.ErrorCode() {
	case "AccessDenied", "AccessDeniedException", "UnauthorizedException":
		return true
	default:
		return false
	}
}

// listIdentityCenterUsers returns Identity Center users in the identity
// store, with grants for assignments on this session's account.
//
// Discovery walks Identity Center regions until ListInstances returns an
// instance. AccessDenied on ListInstances, or any non-cancel error talking
// to SSO Admin after discovery, degrades to IAM-only (a member account, a
// custom role without the needed sso:*, or no instance at all). Identity
// store failures fail the fetch.
func listIdentityCenterUsers(
	ctx context.Context,
	session *cloudaws.Session,
	logger *log.Logger,
) ([]identityCenterUser, error) {
	return listIdentityCenterUsersInRegions(
		ctx,
		session,
		logger,
		identityCenterRegions(session.Partition(), session.Config().Region),
	)
}

func listIdentityCenterUsersInRegions(
	ctx context.Context,
	session *cloudaws.Session,
	logger *log.Logger,
	regions []string,
) ([]identityCenterUser, error) {
	instance, err := discoverIdentityCenter(ctx, session, logger, regions)
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

	cfg := session.Config()
	cfg.Region = instance.region
	sso := ssoadmin.NewFromConfig(cfg)

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

	store := identitystore.NewFromConfig(cfg)

	if err := expandIdentityCenterGroupAssignments(ctx, store, instance.storeID, groups, grants); err != nil {
		return nil, err
	}

	users, err := listIdentityStoreUsers(ctx, store, instance.storeID, grants)
	if err != nil {
		return nil, err
	}

	for i := range users {
		users[i].ARN = identityCenterUserARN(session.Partition(), session.AccountID(), users[i].ID)
	}

	if len(users) == 0 {
		return users, nil
	}

	if err := enrichIdentityCenterActivity(ctx, session, instance, users); err != nil {
		if ctx.Err() != nil {
			return nil, err
		}

		logger.WarnCtx(
			ctx,
			"cannot enrich identity center activity, reporting mfa and last login unknown",
			log.Error(err),
		)
	}

	return users, nil
}

func discoverIdentityCenter(
	ctx context.Context,
	session *cloudaws.Session,
	logger *log.Logger,
	regions []string,
) (identityCenterInstance, error) {
	sessionRegion := session.Config().Region

	for _, region := range regions {
		cfg := session.Config()
		cfg.Region = region
		client := ssoadmin.NewFromConfig(cfg)

		instance, err := listIdentityCenterInstances(ctx, client)
		if err != nil {
			if ctx.Err() != nil {
				return identityCenterInstance{}, err
			}

			if identityCenterAccessDenied(err) {
				return identityCenterInstance{}, err
			}

			continue
		}

		if instance.arn == "" {
			continue
		}

		instance.region = region
		if region != sessionRegion {
			logger.InfoCtx(
				ctx,
				"discovered iam identity center in a non-session region",
				log.String("region", region),
			)
		}

		return instance, nil
	}

	return identityCenterInstance{}, nil
}

func listIdentityCenterInstances(
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

func listIdentityStoreUsers(
	ctx context.Context,
	client *identitystore.Client,
	storeID string,
	grants map[string]*identityCenterGrants,
) ([]identityCenterUser, error) {
	paginator := identitystore.NewListUsersPaginator(
		client,
		&identitystore.ListUsersInput{IdentityStoreId: aws.String(storeID)},
	)

	var users []identityCenterUser

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
			if userID == "" {
				continue
			}

			listed := identityCenterUser{
				ID:          userID,
				UserName:    aws.ToString(user.UserName),
				DisplayName: aws.ToString(user.DisplayName),
				Email:       identityCenterEmail(user.Emails, aws.ToString(user.UserName)),
				Title:       aws.ToString(user.Title),
				Status:      user.UserStatus,
				CreatedAt:   user.CreatedAt,
			}

			if grant, ok := grants[userID]; ok {
				listed.Grants = slices.Compact(slices.Sorted(slices.Values(grant.names)))
				listed.Admin = grant.admin
			}

			users = append(users, listed)
		}
	}

	return nil, fmt.Errorf("cannot list all iam identity store users: %w", ErrPaginationLimitReached)
}

func identityCenterAssignedUsers(users []identityCenterUser) []identityCenterUser {
	assigned := make([]identityCenterUser, 0, len(users))

	for _, user := range users {
		if len(user.Grants) > 0 {
			assigned = append(assigned, user)
		}
	}

	return assigned
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

// identityCenterUserRecord maps one Identity Center user. MFA and last
// login come from registered devices and CloudTrail UserAuthentication
// when those calls succeed. Active follows UserStatus when the store
// reports it.
func identityCenterUserRecord(user identityCenterUser) AccountRecord {
	return AccountRecord{
		Email:       user.Email,
		FullName:    identityCenterFullName(user),
		JobTitle:    user.Title,
		Roles:       user.Grants,
		IsAdmin:     identityCenterIsAdmin(user),
		Active:      identityCenterActive(user.Status),
		MFAStatus:   awsMFAStatus(user.MFAEnabled),
		AuthMethod:  coredata.AccessReviewEntryAuthMethodSSO,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		LastLogin:   user.LastLogin,
		CreatedAt:   user.CreatedAt,
		ExternalID:  user.ARN,
	}
}
