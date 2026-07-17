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

type RightsRequestType string

const (
	RightsRequestTypeAccess        RightsRequestType = "ACCESS"
	RightsRequestTypeDeletion      RightsRequestType = "DELETION"
	RightsRequestTypeRectification RightsRequestType = "RECTIFICATION"
	RightsRequestTypePortability   RightsRequestType = "PORTABILITY"
	RightsRequestTypeObjection     RightsRequestType = "OBJECTION"
	RightsRequestTypeComplaint     RightsRequestType = "COMPLAINT"
)

var (
	_ fmt.Stringer             = RightsRequestType("")
	_ encoding.TextMarshaler   = RightsRequestType("")
	_ encoding.TextUnmarshaler = (*RightsRequestType)(nil)
)

func RightsRequestTypes() []RightsRequestType {
	return []RightsRequestType{
		RightsRequestTypeAccess,
		RightsRequestTypeDeletion,
		RightsRequestTypeRectification,
		RightsRequestTypePortability,
		RightsRequestTypeObjection,
		RightsRequestTypeComplaint,
	}
}

func (v RightsRequestType) IsValid() bool {
	switch v {
	case
		RightsRequestTypeAccess,
		RightsRequestTypeDeletion,
		RightsRequestTypeRectification,
		RightsRequestTypePortability,
		RightsRequestTypeObjection,
		RightsRequestTypeComplaint:
		return true
	}

	return false
}

func (v RightsRequestType) String() string {
	return string(v)
}

func (v RightsRequestType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RightsRequestType) UnmarshalText(text []byte) error {
	val := RightsRequestType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RightsRequestType value: %q", string(text))
	}

	*v = val

	return nil
}
