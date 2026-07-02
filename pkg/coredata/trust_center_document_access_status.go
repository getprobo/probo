// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package coredata

import (
	"encoding"
	"fmt"
)

type TrustCenterDocumentAccessStatus string

const (
	TrustCenterDocumentAccessStatusRequested TrustCenterDocumentAccessStatus = "REQUESTED"
	TrustCenterDocumentAccessStatusGranted   TrustCenterDocumentAccessStatus = "GRANTED"
	TrustCenterDocumentAccessStatusRejected  TrustCenterDocumentAccessStatus = "REJECTED"
	TrustCenterDocumentAccessStatusRevoked   TrustCenterDocumentAccessStatus = "REVOKED"
)

var (
	_ fmt.Stringer             = TrustCenterDocumentAccessStatus("")
	_ encoding.TextMarshaler   = TrustCenterDocumentAccessStatus("")
	_ encoding.TextUnmarshaler = (*TrustCenterDocumentAccessStatus)(nil)
)

func TrustCenterDocumentAccessStatuses() []TrustCenterDocumentAccessStatus {
	return []TrustCenterDocumentAccessStatus{
		TrustCenterDocumentAccessStatusRequested,
		TrustCenterDocumentAccessStatusGranted,
		TrustCenterDocumentAccessStatusRejected,
		TrustCenterDocumentAccessStatusRevoked,
	}
}

func (v TrustCenterDocumentAccessStatus) IsValid() bool {
	switch v {
	case
		TrustCenterDocumentAccessStatusRequested,
		TrustCenterDocumentAccessStatusGranted,
		TrustCenterDocumentAccessStatusRejected,
		TrustCenterDocumentAccessStatusRevoked:
		return true
	}

	return false
}

func (v TrustCenterDocumentAccessStatus) String() string {
	return string(v)
}

func (v TrustCenterDocumentAccessStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *TrustCenterDocumentAccessStatus) UnmarshalText(text []byte) error {
	val := TrustCenterDocumentAccessStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid TrustCenterDocumentAccessStatus value: %q", string(text))
	}

	*v = val

	return nil
}
