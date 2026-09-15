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

package coredata

import (
	"encoding"
	"fmt"

	"go.probo.inc/probo/pkg/page"
)

type CommonGVLVendorOrderField string

const (
	CommonGVLVendorOrderFieldName        CommonGVLVendorOrderField = "NAME"
	CommonGVLVendorOrderFieldIABVendorID CommonGVLVendorOrderField = "IAB_VENDOR_ID"
)

var (
	_ page.OrderField          = CommonGVLVendorOrderField("")
	_ fmt.Stringer             = CommonGVLVendorOrderField("")
	_ encoding.TextMarshaler   = CommonGVLVendorOrderField("")
	_ encoding.TextUnmarshaler = (*CommonGVLVendorOrderField)(nil)
)

func CommonGVLVendorOrderFields() []CommonGVLVendorOrderField {
	return []CommonGVLVendorOrderField{
		CommonGVLVendorOrderFieldName,
		CommonGVLVendorOrderFieldIABVendorID,
	}
}

func (v CommonGVLVendorOrderField) IsValid() bool {
	return isValidOrderField(v, CommonGVLVendorOrderFields())
}

func (v CommonGVLVendorOrderField) String() string {
	return string(v)
}

func (v CommonGVLVendorOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CommonGVLVendorOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, CommonGVLVendorOrderFields())
}

func (v CommonGVLVendorOrderField) Column() string {
	switch v {
	case CommonGVLVendorOrderFieldName:
		return "name"
	case CommonGVLVendorOrderFieldIABVendorID:
		return "iab_vendor_id"
	}

	panic(fmt.Sprintf("unsupported order by: %s", v))
}
