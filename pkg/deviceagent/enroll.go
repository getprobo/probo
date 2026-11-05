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

package deviceagent

import "errors"

var (
	// ErrEnrollmentTokenUsed is returned for a replayed or unknown enrollment token.
	ErrEnrollmentTokenUsed = errors.New("this link was already used")

	// ErrEnrollmentTokenExpired is returned when the enrollment token is past its expiry.
	ErrEnrollmentTokenExpired = errors.New("this enrollment link has expired")

	// ErrEnrollmentTokenInvalid is returned when the enrollment token cannot be exchanged.
	ErrEnrollmentTokenInvalid = errors.New("this enrollment link is invalid")

	// ErrAlreadyEnrolledInOrganization is returned when this machine is already enrolled in that organization.
	ErrAlreadyEnrolledInOrganization = errors.New("this machine is already enrolled in that organization")
)
