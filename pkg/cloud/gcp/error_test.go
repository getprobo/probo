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

package gcp_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.gearno.de/kit/log"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"google.golang.org/api/googleapi"
)

func TestSafeLogFields_OmitsServiceAccountEmail(t *testing.T) {
	t.Parallel()

	const email = "ci@my-project.iam.gserviceaccount.com"

	message := "Permission 'iam.serviceAccountKeys.list' denied on resource " +
		"'projects/-/serviceAccounts/" + email + "' (or it may not exist)."

	tests := []struct {
		name string
		err  error
		want []log.Attr
	}{
		{
			name: "permission denied with reason",
			err: &googleapi.Error{
				Code:    http.StatusForbidden,
				Message: message,
				Errors: []googleapi.ErrorItem{
					{Reason: "forbidden", Message: message},
				},
			},
			want: []log.Attr{
				log.Int("status", http.StatusForbidden),
				log.String("reason", "forbidden"),
			},
		},
		{
			name: "permission denied without reason",
			err: &googleapi.Error{
				Code:    http.StatusForbidden,
				Message: message,
			},
			want: []log.Attr{log.Int("status", http.StatusForbidden)},
		},
		{
			name: "wrapped google api error",
			err: fmt.Errorf(
				"cannot list keys of a gcp service account: %w",
				&googleapi.Error{
					Code:    http.StatusForbidden,
					Message: message,
					Errors: []googleapi.ErrorItem{
						{Reason: "forbidden", Message: message},
					},
				},
			),
			want: []log.Attr{
				log.Int("status", http.StatusForbidden),
				log.String("reason", "forbidden"),
			},
		},
		{
			name: "non google api error",
			err:  fmt.Errorf("cannot list keys of a gcp service account"),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got := cloudgcp.SafeLogFields(tt.err)
				assert.Equal(t, tt.want, got)

				for _, field := range got {
					assert.NotContains(t, field.Key, email)
					assert.NotContains(t, field.Value.String(), email)
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
			err:  &googleapi.Error{Code: http.StatusForbidden, Message: "denied"},
			want: true,
		},
		{
			name: "wrapped forbidden",
			err:  fmt.Errorf("cannot list: %w", &googleapi.Error{Code: http.StatusForbidden}),
			want: true,
		},
		{
			name: "not found",
			err:  &googleapi.Error{Code: http.StatusNotFound},
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

				assert.Equal(t, tt.want, cloudgcp.As[cloudgcp.ErrPermissionDenied](tt.err))
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
			err:  &googleapi.Error{Code: http.StatusNotFound, Message: "missing"},
			want: true,
		},
		{
			name: "wrapped not found",
			err:  fmt.Errorf("cannot get: %w", &googleapi.Error{Code: http.StatusNotFound}),
			want: true,
		},
		{
			name: "forbidden",
			err:  &googleapi.Error{Code: http.StatusForbidden},
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

				assert.Equal(t, tt.want, cloudgcp.As[cloudgcp.ErrNotFound](tt.err))
			},
		)
	}
}
