// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

// This file deliberately holds NO package-level Probe(ctx, Provider)
// function -- each typed *AWSProvider / *GCPProvider / *AzureProvider
// implements Probe(ctx) directly (satisfying the Probeable interface).
// Shared helpers (chiefly MapSDKError, which converts raw cloud-SDK
// errors into the typed package sentinels every caller switches on)
// live here.

package cloudaccount

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	smithy "github.com/aws/smithy-go"
	"google.golang.org/api/googleapi"
)

// MapSDKError converts a raw AWS / GCP / Azure SDK error chain into
// one of the typed cloudaccount package sentinels
// (ErrCredentialsInvalid, ErrInsufficientPermissions,
// ErrScopeUnreachable). Errors that don't match any known shape pass
// through with their original wrapping intact -- callers should treat
// those as "unclassified probe failure".
//
// The mapping rules are intentionally narrow and recover only the
// ambiguities that matter for the customer-facing status surface
// (last_probe_error). Everything else stays raw so we don't silently
// swallow new SDK error shapes.
func MapSDKError(err error) error {
	if err == nil {
		return nil
	}

	// AWS smithy-go API error: the SDK wraps every modelled error
	// in this interface and exposes the canonical service-side
	// error code.
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "AccessDenied", "AccessDeniedException", "UnauthorizedOperation":
			return fmt.Errorf("%w: %s", ErrInsufficientPermissions, apiErr.ErrorMessage())
		case "InvalidClientTokenId",
			"SignatureDoesNotMatch",
			"ExpiredToken",
			"ExpiredTokenException",
			"InvalidAccessKeyId",
			"InvalidUserID.NotFound":
			return fmt.Errorf("%w: %s", ErrCredentialsInvalid, apiErr.ErrorMessage())
		case "NoSuchEntity",
			"NoSuchEntityException",
			"ResourceNotFoundException",
			"NotFound":
			return fmt.Errorf("%w: %s", ErrScopeUnreachable, apiErr.ErrorMessage())
		}
	}

	// GCP googleapi error: HTTP status maps cleanly to our buckets.
	var gErr *googleapi.Error
	if errors.As(err, &gErr) {
		switch gErr.Code {
		case http.StatusUnauthorized:
			return fmt.Errorf("%w: %s", ErrCredentialsInvalid, gErr.Message)
		case http.StatusForbidden:
			return fmt.Errorf("%w: %s", ErrInsufficientPermissions, gErr.Message)
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s", ErrScopeUnreachable, gErr.Message)
		}
	}

	// Azure ResponseError: HTTP status maps the same way.
	var azErr *azcore.ResponseError
	if errors.As(err, &azErr) {
		switch azErr.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("%w: %s", ErrCredentialsInvalid, azErr.ErrorCode)
		case http.StatusForbidden:
			return fmt.Errorf("%w: %s", ErrInsufficientPermissions, azErr.ErrorCode)
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s", ErrScopeUnreachable, azErr.ErrorCode)
		}
	}

	// Azure azidentity / GCP google.CredentialsFromJSON failures
	// don't surface a typed error, but their messages are stable
	// enough to fingerprint without false positives.
	msg := err.Error()
	switch {
	case strings.Contains(msg, "AADSTS70011"),
		strings.Contains(msg, "AADSTS7000215"),
		strings.Contains(msg, "AADSTS700016"),
		strings.Contains(msg, "invalid_client"),
		strings.Contains(msg, "invalid_grant"):
		return fmt.Errorf("%w: %s", ErrCredentialsInvalid, msg)
	}

	return err
}
