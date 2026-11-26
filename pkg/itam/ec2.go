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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	aws "go.probo.inc/probo/pkg/cloud/aws"
)

func ListEC2Instances(
	ctx context.Context,
	session *aws.Session,
) ([]AssetObservation, error) {
	if session == nil {
		return nil, fmt.Errorf("cannot list EC2 instances: missing session")
	}

	regions, err := enabledRegions(ctx, session)
	if err != nil {
		return nil, err
	}

	listed, err := listByRegion(
		ctx,
		regions,
		func(ctx context.Context, region string) ([]types.Instance, error) {
			instances, err := listEC2InstancesInRegion(ctx, session, region)
			if err != nil {
				return nil, fmt.Errorf("cannot list EC2 instances in %s: %w", region, err)
			}

			return instances, nil
		},
	)
	if err != nil {
		return nil, err
	}

	observed := make([]AssetObservation, 0)

	for _, page := range listed {
		for _, instance := range page.items {
			if ec2InstanceTerminated(instance) {
				continue
			}

			item, err := mapEC2Instance(session, page.region, instance)
			if err != nil {
				return nil, err
			}

			observed = append(observed, item)
		}
	}

	return observed, nil
}

func listEC2InstancesInRegion(
	ctx context.Context,
	session *aws.Session,
	region string,
) ([]types.Instance, error) {
	client := ec2.NewFromConfig(session.Config(), func(options *ec2.Options) {
		options.Region = region
	})

	return collectPages(
		ctx,
		maxAssetListPages,
		func(ctx context.Context, token *string) ([]types.Instance, bool, *string, error) {
			out, err := client.DescribeInstances(
				ctx,
				&ec2.DescribeInstancesInput{NextToken: token},
			)
			if err != nil {
				return nil, false, nil, fmt.Errorf("cannot list EC2 instances: %w", err)
			}

			if out == nil {
				return nil, false, nil, nil
			}

			instances := make([]types.Instance, 0)
			for _, reservation := range out.Reservations {
				instances = append(instances, reservation.Instances...)
			}

			truncated, next := pageToken(out.NextToken)

			return instances, truncated, next, nil
		},
	)
}

func ec2InstanceTerminated(item types.Instance) bool {
	return item.State != nil && item.State.Name == types.InstanceStateNameTerminated
}

func mapEC2Instance(
	session *aws.Session,
	region string,
	item types.Instance,
) (AssetObservation, error) {
	if item.InstanceId == nil || *item.InstanceId == "" {
		return AssetObservation{}, fmt.Errorf("cannot map EC2 instance: missing instance id")
	}

	instanceID := *item.InstanceId
	name := instanceID

	for _, tag := range item.Tags {
		if tag.Key == nil || *tag.Key != "Name" || tag.Value == nil || *tag.Value == "" {
			continue
		}

		name = *tag.Value

		break
	}

	groupIDs := make([]string, 0, len(item.SecurityGroups))
	for _, group := range item.SecurityGroups {
		if group.GroupId == nil || *group.GroupId == "" {
			continue
		}

		groupIDs = append(groupIDs, *group.GroupId)
	}

	configuration := EC2InstanceConfiguration{
		InstanceType:     string(item.InstanceType),
		ImageID:          stringOrEmpty(item.ImageId),
		VpcID:            stringOrEmpty(item.VpcId),
		SubnetID:         stringOrEmpty(item.SubnetId),
		PublicIP:         stringOrEmpty(item.PublicIpAddress),
		SecurityGroupIDs: groupIDs,
	}

	if item.State != nil {
		configuration.State = string(item.State.Name)
	}

	if item.IamInstanceProfile != nil {
		configuration.IAMInstanceProfileARN = stringOrEmpty(item.IamInstanceProfile.Arn)
	}

	return AssetObservation{
		ExternalID: arn.ARN{
			Partition: session.Partition(),
			Service:   "ec2",
			Region:    region,
			AccountID: session.AccountID(),
			Resource:  "instance/" + instanceID,
		}.String(),
		Name:          name,
		Region:        new(region),
		Configuration: configuration,
		SchemaVersion: EC2InstanceSchemaVersion,
	}, nil
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
