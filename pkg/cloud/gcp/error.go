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

	"go.gearno.de/kit/log"
	"google.golang.org/api/googleapi"
)

// SafeLogFields returns status and reason from a Google API error. It
// never includes Message, Body, or Errors[].Message: those strings
// often name the resource, which for IAM is a service-account email.
func SafeLogFields(err error) []log.Attr {
	apiErr, ok := errors.AsType[*googleapi.Error](err)
	if !ok {
		return nil
	}

	fields := []log.Attr{log.Int("status", apiErr.Code)}
	if len(apiErr.Errors) > 0 && apiErr.Errors[0].Reason != "" {
		fields = append(fields, log.String("reason", apiErr.Errors[0].Reason))
	}

	return fields
}
