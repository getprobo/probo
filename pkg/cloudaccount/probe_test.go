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

package cloudaccount_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	smithy "github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/googleapi"

	"go.probo.inc/probo/pkg/cloudaccount"
)

func TestMapSDKError_AWSSmithy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		code     string
		expected error
	}{
		{"access denied -> insufficient permissions", "AccessDenied", cloudaccount.ErrInsufficientPermissions},
		{"unauthorized operation -> insufficient permissions", "UnauthorizedOperation", cloudaccount.ErrInsufficientPermissions},
		{"expired token -> credentials invalid", "ExpiredTokenException", cloudaccount.ErrCredentialsInvalid},
		{"invalid client token -> credentials invalid", "InvalidClientTokenId", cloudaccount.ErrCredentialsInvalid},
		{"no such entity -> scope unreachable", "NoSuchEntity", cloudaccount.ErrScopeUnreachable},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := &smithy.GenericAPIError{Code: tc.code, Message: "stub message"}
			mapped := cloudaccount.MapSDKError(err)
			assert.ErrorIs(t, mapped, tc.expected)
		})
	}
}

func TestMapSDKError_GCPGoogleapi(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		code     int
		expected error
	}{
		{"401 -> credentials invalid", http.StatusUnauthorized, cloudaccount.ErrCredentialsInvalid},
		{"403 -> insufficient permissions", http.StatusForbidden, cloudaccount.ErrInsufficientPermissions},
		{"404 -> scope unreachable", http.StatusNotFound, cloudaccount.ErrScopeUnreachable},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := &googleapi.Error{Code: tc.code, Message: "stub message"}
			mapped := cloudaccount.MapSDKError(err)
			assert.ErrorIs(t, mapped, tc.expected)
		})
	}
}

func TestMapSDKError_AzureResponseError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		code     int
		expected error
	}{
		{"401 -> credentials invalid", http.StatusUnauthorized, cloudaccount.ErrCredentialsInvalid},
		{"403 -> insufficient permissions", http.StatusForbidden, cloudaccount.ErrInsufficientPermissions},
		{"404 -> scope unreachable", http.StatusNotFound, cloudaccount.ErrScopeUnreachable},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := &azcore.ResponseError{StatusCode: tc.code, ErrorCode: "stub"}
			mapped := cloudaccount.MapSDKError(err)
			assert.ErrorIs(t, mapped, tc.expected)
		})
	}
}

func TestMapSDKError_AzureAADStringFingerprint(t *testing.T) {
	t.Parallel()

	err := errors.New("AADSTS7000215: Invalid client secret provided")
	mapped := cloudaccount.MapSDKError(err)
	assert.ErrorIs(t, mapped, cloudaccount.ErrCredentialsInvalid)
}

func TestMapSDKError_NilPassesThrough(t *testing.T) {
	t.Parallel()

	assert.NoError(t, cloudaccount.MapSDKError(nil))
}

func TestMapSDKError_UnknownPassesThrough(t *testing.T) {
	t.Parallel()

	original := errors.New("nothing typed about this")
	mapped := cloudaccount.MapSDKError(original)
	assert.ErrorIs(t, mapped, original)
	assert.NotErrorIs(t, mapped, cloudaccount.ErrCredentialsInvalid)
	assert.NotErrorIs(t, mapped, cloudaccount.ErrInsufficientPermissions)
	assert.NotErrorIs(t, mapped, cloudaccount.ErrScopeUnreachable)
}
