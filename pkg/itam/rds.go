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

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	aws "go.probo.inc/probo/pkg/cloud/aws"
)

func ListRDSDBInstances(
	ctx context.Context,
	session *aws.Session,
) ([]AssetObservation, error) {
	if session == nil {
		return nil, fmt.Errorf("cannot list RDS instances: missing session")
	}

	regions, err := enabledRegions(ctx, session)
	if err != nil {
		return nil, err
	}

	listed, err := listByRegion(
		ctx,
		regions,
		func(ctx context.Context, region string) ([]types.DBInstance, error) {
			instances, err := listRDSDBInstancesInRegion(ctx, session, region)
			if err != nil {
				return nil, fmt.Errorf("cannot list RDS instances in %s: %w", region, err)
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
			item, err := mapRDSDBInstance(page.region, instance)
			if err != nil {
				return nil, err
			}

			observed = append(observed, item)
		}
	}

	return observed, nil
}

func listRDSDBInstancesInRegion(
	ctx context.Context,
	session *aws.Session,
	region string,
) ([]types.DBInstance, error) {
	client := rds.NewFromConfig(session.Config(), func(options *rds.Options) {
		options.Region = region
	})

	return collectPages(
		ctx,
		maxAssetListPages,
		func(ctx context.Context, marker *string) ([]types.DBInstance, bool, *string, error) {
			out, err := client.DescribeDBInstances(
				ctx,
				&rds.DescribeDBInstancesInput{Marker: marker},
			)
			if err != nil {
				return nil, false, nil, fmt.Errorf("cannot list RDS instances: %w", err)
			}

			if out == nil {
				return nil, false, nil, nil
			}

			truncated, next := pageToken(out.Marker)

			return out.DBInstances, truncated, next, nil
		},
	)
}

func mapRDSDBInstance(region string, item types.DBInstance) (AssetObservation, error) {
	if item.DBInstanceArn == nil || *item.DBInstanceArn == "" {
		return AssetObservation{}, fmt.Errorf("cannot map RDS instance: missing ARN")
	}

	if item.DBInstanceIdentifier == nil || *item.DBInstanceIdentifier == "" {
		return AssetObservation{}, fmt.Errorf("cannot map RDS instance: missing identifier")
	}

	configuration := RDSDBInstanceConfiguration{
		Engine:             stringOrEmpty(item.Engine),
		EngineVersion:      stringOrEmpty(item.EngineVersion),
		Class:              stringOrEmpty(item.DBInstanceClass),
		Status:             stringOrEmpty(item.DBInstanceStatus),
		PubliclyAccessible: boolOrFalse(item.PubliclyAccessible),
		StorageEncrypted:   boolOrFalse(item.StorageEncrypted),
		MultiAZ:            boolOrFalse(item.MultiAZ),
	}

	if item.Endpoint != nil {
		configuration.EndpointHost = stringOrEmpty(item.Endpoint.Address)
	}

	return AssetObservation{
		ExternalID:    *item.DBInstanceArn,
		Name:          *item.DBInstanceIdentifier,
		Region:        new(region),
		Configuration: configuration,
		SchemaVersion: RDSDBInstanceSchemaVersion,
	}, nil
}

func boolOrFalse(value *bool) bool {
	return value != nil && *value
}
