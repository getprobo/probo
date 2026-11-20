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

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	aws "go.probo.inc/probo/pkg/cloud/aws"
)

type (
	regionPage[T any] struct {
		region string
		items  []T
	}
)

// enabledRegions lists regions the account has enabled. DescribeRegions
// omits opted-out regions unless AllRegions is set. An empty list is a
// successful kind with nothing to store.
func enabledRegions(ctx context.Context, session *aws.Session) ([]string, error) {
	client := ec2.NewFromConfig(session.Config())

	out, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("cannot list AWS regions: %w", err)
	}

	if out == nil {
		return []string{}, nil
	}

	regions := make([]string, 0, len(out.Regions))
	for _, region := range out.Regions {
		if region.RegionName == nil || *region.RegionName == "" {
			continue
		}

		regions = append(regions, *region.RegionName)
	}

	return regions, nil
}

// listByRegion calls list once per region. A region error discards every
// page already collected, so the caller writes nothing for the kind.
func listByRegion[T any](
	ctx context.Context,
	regions []string,
	list func(context.Context, string) ([]T, error),
) ([]regionPage[T], error) {
	listed := make([]regionPage[T], 0, len(regions))

	for _, region := range regions {
		items, err := list(ctx, region)
		if err != nil {
			return nil, err
		}

		listed = append(listed, regionPage[T]{region: region, items: items})
	}

	return listed, nil
}

func pageToken(token *string) (bool, *string) {
	if token == nil || *token == "" {
		return false, nil
	}

	return true, token
}
