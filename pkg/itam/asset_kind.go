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

	aws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	CloudFrontDistributionKind          = "aws.cloudfront.distribution"
	CloudFrontDistributionSchemaVersion = 1

	// CloudFront's control plane answers in us-east-1.
	cloudFrontRegion = "us-east-1"

	// maxAssetListPages is the same cap as
	// accessreview/drivers.maxPaginationPages. That constant is unexported.
	maxAssetListPages = 500
)

type (
	// CloudFrontDistributionConfiguration is schema version 1. encoding/json
	// ignores unknown fields and zero-fills missing ones.
	CloudFrontDistributionConfiguration struct {
		DomainName         string `json:"domain_name"`
		Enabled            bool   `json:"enabled"`
		Status             string `json:"status"`
		PriceClass         string `json:"price_class"`
		DefaultCertificate bool   `json:"default_certificate"`
	}

	AssetObservation struct {
		ExternalID    string
		Name          string
		Region        *string
		Configuration any
		SchemaVersion int
	}

	AssetLister func(context.Context, *aws.Session) ([]AssetObservation, error)

	assetKind struct {
		key      string
		provider coredata.ConnectorProvider
		list     AssetLister
	}
)

func hasDiscoverer(provider coredata.ConnectorProvider) bool {
	for _, kind := range defaultAssetKinds {
		if kind.provider == provider {
			return true
		}
	}

	return false
}
