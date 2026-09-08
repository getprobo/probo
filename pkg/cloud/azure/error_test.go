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

package azure_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/stretchr/testify/assert"
	"go.gearno.de/kit/log"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
)

func TestSafeLogFields_OmitsResponseBody(t *testing.T) {
	t.Parallel()

	const (
		upn            = "alice@contoso.com"
		subscriptionID = "11111111-1111-1111-1111-111111111111"
	)

	body := `{"error":{"message":"Principal ` + upn + ` cannot read subscription ` + subscriptionID + `"}}`

	tests := []struct {
		name string
		err  error
		want []log.Attr
	}{
		{
			name: "forbidden with error code",
			err: &azcore.ResponseError{
				StatusCode: http.StatusForbidden,
				ErrorCode:  "AuthorizationFailed",
				RawResponse: &http.Response{
					Body: io.NopCloser(strings.NewReader(body)),
				},
			},
			want: []log.Attr{
				log.Int("status", http.StatusForbidden),
				log.String("error_code", "AuthorizationFailed"),
			},
		},
		{
			name: "forbidden without error code",
			err: &azcore.ResponseError{
				StatusCode: http.StatusForbidden,
				RawResponse: &http.Response{
					Body: io.NopCloser(strings.NewReader(body)),
				},
			},
			want: []log.Attr{log.Int("status", http.StatusForbidden)},
		},
		{
			name: "wrapped azure api error",
			err: fmt.Errorf(
				"cannot list azure role assignments: %w",
				&azcore.ResponseError{
					StatusCode: http.StatusForbidden,
					ErrorCode:  "AuthorizationFailed",
					RawResponse: &http.Response{
						Body: io.NopCloser(strings.NewReader(body)),
					},
				},
			),
			want: []log.Attr{
				log.Int("status", http.StatusForbidden),
				log.String("error_code", "AuthorizationFailed"),
			},
		},
		{
			name: "non azure api error",
			err:  fmt.Errorf("cannot list azure role assignments"),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got := cloudazure.SafeLogFields(tt.err)
				assert.Equal(t, tt.want, got)

				for _, field := range got {
					assert.NotContains(t, field.Key, upn)
					assert.NotContains(t, field.Key, subscriptionID)
					assert.NotContains(t, field.Value.String(), upn)
					assert.NotContains(t, field.Value.String(), subscriptionID)
					assert.NotContains(t, field.Value.String(), body)
				}
			},
		)
	}
}

func TestAs_PermissionDenied(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "forbidden",
			err:  &azcore.ResponseError{StatusCode: http.StatusForbidden, ErrorCode: "AuthorizationFailed"},
			want: true,
		},
		{
			name: "wrapped forbidden",
			err:  fmt.Errorf("cannot list: %w", &azcore.ResponseError{StatusCode: http.StatusForbidden}),
			want: true,
		},
		{
			name: "not found",
			err:  &azcore.ResponseError{StatusCode: http.StatusNotFound},
		},
		{
			name: "canceled",
			err:  context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, cloudazure.As[cloudazure.ErrPermissionDenied](tt.err))
			},
		)
	}
}

func TestAs_NotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "not found",
			err:  &azcore.ResponseError{StatusCode: http.StatusNotFound, ErrorCode: "ResourceNotFound"},
			want: true,
		},
		{
			name: "wrapped not found",
			err:  fmt.Errorf("cannot get: %w", &azcore.ResponseError{StatusCode: http.StatusNotFound}),
			want: true,
		},
		{
			name: "forbidden",
			err:  &azcore.ResponseError{StatusCode: http.StatusForbidden},
		},
		{
			name: "canceled",
			err:  context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, cloudazure.As[cloudazure.ErrNotFound](tt.err))
			},
		)
	}
}
