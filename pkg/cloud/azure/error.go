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

package azure

import (
	"errors"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"go.gearno.de/kit/log"
)

// SafeLogFields returns status and error code from an Azure API error. It
// never includes RawResponse or the body: a Graph error names a UPN and an
// ARM error names a subscription.
func SafeLogFields(err error) []log.Attr {
	apiErr, ok := errors.AsType[*azcore.ResponseError](err)
	if !ok {
		return nil
	}

	fields := []log.Attr{log.Int("status", apiErr.StatusCode)}
	if apiErr.ErrorCode != "" {
		fields = append(fields, log.String("error_code", apiErr.ErrorCode))
	}

	return fields
}

type (
	// ErrPermissionDenied matches an Azure API 403. Pass it to As.
	ErrPermissionDenied struct{}

	// ErrNotFound matches an Azure API 404. Pass it to As.
	ErrNotFound struct{}

	statusError interface {
		ErrPermissionDenied | ErrNotFound
		code() int
	}
)

func (ErrPermissionDenied) code() int { return http.StatusForbidden }

func (ErrNotFound) code() int { return http.StatusNotFound }

// As reports whether err is or wraps an Azure API error whose status is T.
func As[T statusError](err error) bool {
	apiErr, ok := errors.AsType[*azcore.ResponseError](err)
	if !ok {
		return false
	}

	var status T

	return apiErr.StatusCode == status.code()
}
