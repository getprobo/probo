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

package gcp

import (
	"errors"
	"strings"
)

const (
	// CommercialUniverse is public Google Cloud. API hosts are
	// {service}.googleapis.com.
	CommercialUniverse = "googleapis.com"

	// S3NSUniverse is Cloud de Confiance by S3NS. API hosts are
	// {service}.s3nsapis.fr.
	S3NSUniverse = "s3nsapis.fr"

	// iamHost is the FRN host in JWT and STS audiences. S3NS keeps
	// iam.googleapis.com in resource names; only HTTP API hosts change.
	iamHost = "iam.googleapis.com"

	s3nsIAMHost = "iam.s3nsapis.fr"

	commercialServiceAccountSuffix = ".iam.gserviceaccount.com"
	s3nsServiceAccountSuffix       = ".s3ns.iam.gserviceaccount.com"
)

var (
	// errUnsupportedUniverse is returned when the provider resource names a
	// host other than commercial IAM or S3NS IAM.
	errUnsupportedUniverse = errors.New("workloadIdentityProvider is not a supported GCP universe")

	// errUniverseMismatch is returned when the provider host is S3NS but the
	// service account email is commercial. The reverse is allowed: S3NS FRNs
	// still use iam.googleapis.com.
	errUniverseMismatch = errors.New("workloadIdentityProvider and serviceAccountEmail name different GCP universes")
)

func supportedIAMHost(host string) bool {
	switch host {
	case iamHost, s3nsIAMHost:
		return true
	default:
		return false
	}
}

func universeFromServiceAccountEmail(email string) (string, bool) {
	if strings.HasSuffix(email, s3nsServiceAccountSuffix) {
		return S3NSUniverse, true
	}

	if strings.HasSuffix(email, commercialServiceAccountSuffix) {
		return CommercialUniverse, true
	}

	return "", false
}

// inferUniverse picks the Google Cloud universe from the service-account
// email. A pasted iam.googleapis.com host is not a commercial signal: S3NS
// full resource names keep that host. A pasted iam.s3nsapis.fr host with a
// commercial email is a mismatch.
func inferUniverse(p providerResource, email string) (string, error) {
	fromEmail, ok := universeFromServiceAccountEmail(email)
	if !ok {
		return "", errInvalidServiceAccountEmail
	}

	if p.iamHost == s3nsIAMHost && fromEmail != S3NSUniverse {
		return "", errUniverseMismatch
	}

	return fromEmail, nil
}
