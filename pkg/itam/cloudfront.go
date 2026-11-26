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

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	aws "go.probo.inc/probo/pkg/cloud/aws"
)

var errAssetPageLimit = errors.New("cannot list assets: page limit reached")

func ListCloudFrontDistributions(
	ctx context.Context,
	session *aws.Session,
) ([]AssetObservation, error) {
	if session == nil {
		return nil, fmt.Errorf("cannot list CloudFront distributions: missing session")
	}

	client := cloudfront.NewFromConfig(session.Config(), func(options *cloudfront.Options) {
		options.Region = cloudFrontRegion
	})

	summaries, err := collectPages(
		ctx,
		maxAssetListPages,
		func(ctx context.Context, marker *string) ([]types.DistributionSummary, bool, *string, error) {
			out, err := client.ListDistributions(ctx, &cloudfront.ListDistributionsInput{Marker: marker})
			if err != nil {
				return nil, false, nil, fmt.Errorf("cannot list CloudFront distributions: %w", err)
			}

			if out == nil || out.DistributionList == nil {
				return nil, false, nil, nil
			}

			list := out.DistributionList
			truncated := list.IsTruncated != nil && *list.IsTruncated

			return list.Items, truncated, list.NextMarker, nil
		},
	)
	if err != nil {
		return nil, err
	}

	observed := make([]AssetObservation, 0, len(summaries))

	for _, summary := range summaries {
		item, err := mapCloudFrontDistribution(summary)
		if err != nil {
			return nil, err
		}

		observed = append(observed, item)
	}

	return observed, nil
}

func mapCloudFrontDistribution(item types.DistributionSummary) (AssetObservation, error) {
	if item.ARN == nil || *item.ARN == "" {
		return AssetObservation{}, fmt.Errorf("cannot map CloudFront distribution: missing ARN")
	}

	if item.DomainName == nil || *item.DomainName == "" {
		return AssetObservation{}, fmt.Errorf("cannot map CloudFront distribution: missing domain name")
	}

	configuration := CloudFrontDistributionConfiguration{
		DomainName: *item.DomainName,
		PriceClass: string(item.PriceClass),
	}

	if item.Enabled != nil {
		configuration.Enabled = *item.Enabled
	}

	if item.Status != nil {
		configuration.Status = *item.Status
	}

	if item.ViewerCertificate != nil && item.ViewerCertificate.CloudFrontDefaultCertificate != nil {
		configuration.DefaultCertificate = *item.ViewerCertificate.CloudFrontDefaultCertificate
	}

	return AssetObservation{
		ExternalID:    *item.ARN,
		Name:          *item.DomainName,
		Configuration: configuration,
		SchemaVersion: CloudFrontDistributionSchemaVersion,
	}, nil
}

// collectPages gathers every page. A truncated list that never ends returns
// errAssetPageLimit and no items, so the caller writes nothing.
func collectPages[T any](
	ctx context.Context,
	limit int,
	fetch func(context.Context, *string) (items []T, truncated bool, next *string, err error),
) ([]T, error) {
	if limit < 1 {
		return nil, fmt.Errorf("cannot list assets: page limit must be positive")
	}

	var (
		marker *string
		all    []T
	)

	for range limit {
		items, truncated, next, err := fetch(ctx, marker)
		if err != nil {
			return nil, err
		}

		all = append(all, items...)

		if !truncated {
			return all, nil
		}

		if next == nil || *next == "" {
			return nil, fmt.Errorf("cannot list assets: truncated page has no marker")
		}

		marker = next
	}

	return nil, errAssetPageLimit
}
