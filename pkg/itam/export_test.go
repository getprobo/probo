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

package itam

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	aws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func NewAssetSyncHandlerForTest(
	pgClient *pg.Client,
	connectors *connector.Service,
	list AssetLister,
) *assetSyncHandler {
	h := newAssetSyncHandler(pgClient, connectors, log.NewLogger(log.WithOutput(io.Discard)))
	for i := range h.kinds {
		h.kinds[i].list = list
	}

	return h
}

func NewAssetSyncHandlerWithListersForTest(
	pgClient *pg.Client,
	connectors *connector.Service,
	listers map[string]AssetLister,
) *assetSyncHandler {
	handler := NewAssetSyncHandlerForTest(
		pgClient,
		connectors,
		func(context.Context, *aws.Session) ([]AssetObservation, error) {
			return nil, nil
		},
	)
	SetAssetKindListersForTest(handler, listers)

	return handler
}

func SetAssetKindListersForTest(h *assetSyncHandler, listers map[string]AssetLister) {
	for i := range h.kinds {
		list, ok := listers[h.kinds[i].key]
		if !ok {
			continue
		}

		h.kinds[i].list = list
	}
}

func SetAssetSyncAccountForTest(h *assetSyncHandler, id gid.GID) {
	h.onlyAccount = &id
}

func SyncAccountForTest(
	h *assetSyncHandler,
	ctx context.Context,
	accountID gid.GID,
) (coredata.AssetSync, error) {
	h.onlyAccount = &accountID

	sync, err := h.Claim(ctx)
	if err != nil {
		return coredata.AssetSync{}, err
	}

	if err := h.Process(ctx, sync); err != nil {
		return sync, err
	}

	return sync, nil
}

func ClaimAssetSyncForTest(
	h *assetSyncHandler,
	ctx context.Context,
	accountID gid.GID,
) (coredata.AssetSync, error) {
	h.onlyAccount = &accountID

	return h.Claim(ctx)
}

func RecoverStaleAssetSyncsForTest(h *assetSyncHandler, ctx context.Context) error {
	return h.RecoverStale(ctx)
}

func AssetSyncFailureMessageForTest() string {
	return assetSyncFailureMessage
}

func AssetSyncStaleMessageForTest() string {
	return assetSyncStaleMessage
}

func BackdateAssetSyncForTest(
	ctx context.Context,
	pgClient *pg.Client,
	id gid.GID,
	startedAt time.Time,
) error {
	return pgClient.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(
			ctx,
			`UPDATE asset_syncs SET started_at = $1, completed_at = CASE WHEN completed_at IS NULL THEN completed_at ELSE $1 END WHERE id = $2`,
			startedAt,
			id,
		)
		if err != nil {
			return fmt.Errorf("cannot backdate asset sync: %w", err)
		}

		return nil
	})
}

func CollectPageLimitForTest(limit int) (int, error) {
	calls := 0
	_, err := collectPages(
		context.Background(),
		limit,
		func(context.Context, *string) ([]int, bool, *string, error) {
			calls++
			next := fmt.Sprintf("page-%d", calls)

			return []int{calls}, true, &next, nil
		},
	)

	return calls, err
}

func CollectSinglePageForTest() ([]int, error) {
	return collectPages(
		context.Background(),
		5,
		func(context.Context, *string) ([]int, bool, *string, error) {
			return []int{1, 2}, false, nil, nil
		},
	)
}

func MapCloudFrontForTest() (AssetObservation, error) {
	arn := "arn:aws:cloudfront::123456789012:distribution/EDFDVBD632BHDS5"
	domain := "d111111abcdef8.cloudfront.net"
	enabled := true
	status := "Deployed"
	certificate := true

	return mapCloudFrontDistribution(types.DistributionSummary{
		ARN:        &arn,
		DomainName: &domain,
		Enabled:    &enabled,
		Status:     &status,
		PriceClass: types.PriceClassPriceClass100,
		ViewerCertificate: &types.ViewerCertificate{
			CloudFrontDefaultCertificate: &certificate,
		},
	})
}

func MapCloudFrontMissingARNForTest() error {
	_, err := mapCloudFrontDistribution(types.DistributionSummary{})

	return err
}

func ListByRegionSecondErrorForTest() (int, []int, error) {
	calls := 0

	items, err := listByRegion(
		context.Background(),
		[]string{"us-east-1", "eu-west-1"},
		func(context.Context, string) ([]int, error) {
			calls++
			if calls == 2 {
				return []int{2}, errors.New("denied")
			}

			return []int{1}, nil
		},
	)
	if items == nil {
		return calls, nil, err
	}

	flat := make([]int, 0)
	for _, page := range items {
		flat = append(flat, page.items...)
	}

	return calls, flat, err
}

func MapEC2ForTest() (AssetObservation, error) {
	session := aws.NewSessionFromConfig("123456789012", "aws", awssdk.Config{})
	instanceID := "i-abc"
	name := "web"
	imageID := "ami-1"
	vpcID := "vpc-1"
	subnetID := "subnet-1"
	publicIP := "203.0.113.10"
	profileARN := "arn:aws:iam::123456789012:instance-profile/web"
	groupID := "sg-1"
	emptyGroup := ""
	tagKey := "Name"

	return mapEC2Instance(session, "us-east-1", ec2types.Instance{
		InstanceId:      &instanceID,
		InstanceType:    ec2types.InstanceType("t3.micro"),
		ImageId:         &imageID,
		VpcId:           &vpcID,
		SubnetId:        &subnetID,
		PublicIpAddress: &publicIP,
		State: &ec2types.InstanceState{
			Name: ec2types.InstanceStateName("running"),
		},
		IamInstanceProfile: &ec2types.IamInstanceProfile{Arn: &profileARN},
		SecurityGroups: []ec2types.GroupIdentifier{
			{GroupId: &groupID},
			{GroupId: &emptyGroup},
		},
		Tags: []ec2types.Tag{{Key: &tagKey, Value: &name}},
	})
}

func MapEC2MissingOptionalFieldsForTest() (AssetObservation, error) {
	session := aws.NewSessionFromConfig("123456789012", "aws", awssdk.Config{})
	instanceID := "i-bare"

	return mapEC2Instance(session, "eu-west-1", ec2types.Instance{InstanceId: &instanceID})
}

func MapEC2MissingInstanceIDForTest() error {
	session := aws.NewSessionFromConfig("123456789012", "aws", awssdk.Config{})
	_, err := mapEC2Instance(session, "us-east-1", ec2types.Instance{})

	return err
}

func MapRDSForTest() (AssetObservation, error) {
	arn := "arn:aws:rds:us-east-1:123456789012:db:app"
	identifier := "app"
	engine := "postgres"
	version := "16.3"
	class := "db.t3.micro"
	status := "available"
	host := "app.example.us-east-1.rds.amazonaws.com"
	yes := true

	return mapRDSDBInstance("us-east-1", rdstypes.DBInstance{
		DBInstanceArn:        &arn,
		DBInstanceIdentifier: &identifier,
		Engine:               &engine,
		EngineVersion:        &version,
		DBInstanceClass:      &class,
		DBInstanceStatus:     &status,
		PubliclyAccessible:   &yes,
		StorageEncrypted:     &yes,
		MultiAZ:              &yes,
		Endpoint:             &rdstypes.Endpoint{Address: &host},
	})
}

func MapRDSMissingOptionalFieldsForTest() (AssetObservation, error) {
	arn := "arn:aws:rds:eu-west-1:123456789012:db:bare"
	identifier := "bare"

	return mapRDSDBInstance("eu-west-1", rdstypes.DBInstance{
		DBInstanceArn:        &arn,
		DBInstanceIdentifier: &identifier,
	})
}

func MapRDSMissingARNForTest() error {
	identifier := "app"
	_, err := mapRDSDBInstance("us-east-1", rdstypes.DBInstance{DBInstanceIdentifier: &identifier})

	return err
}

func MapRDSMissingIdentifierForTest() error {
	arn := "arn:aws:rds:us-east-1:123456789012:db:app"
	_, err := mapRDSDBInstance("us-east-1", rdstypes.DBInstance{DBInstanceArn: &arn})

	return err
}
