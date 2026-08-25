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

// Package arn builds and reads AWS resource names the SDK treats as opaque
// strings. Parse and String come from the SDK; this package fills in the
// service-specific resource forms.
package arn

import (
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
)

const (
	// Partition is the commercial AWS partition. GovCloud and China are
	// different deployments with their own account namespaces.
	Partition = "aws"

	iamService = "iam"
)

// Format renders an ARN the way the AWS SDK does.
func Format(partition, service, region, accountID, resource string) string {
	return awsarn.ARN{
		Partition: partition,
		Service:   service,
		Region:    region,
		AccountID: accountID,
		Resource:  resource,
	}.String()
}

// IAM is an IAM ARN in the given partition. IAM is global, so the region
// is empty.
func IAM(partition, accountID, resource string) string {
	return Format(partition, iamService, "", accountID, resource)
}

// RoleARN builds the ARN of a role in one account of the given partition.
func RoleARN(partition, accountID, roleName string) string {
	return IAM(partition, accountID, "role/"+roleName)
}
