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

type DevicePostureReportOrderField string

const (
	DevicePostureReportOrderFieldCreatedAt DevicePostureReportOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = DevicePostureReportOrderField("")
	_ fmt.Stringer             = DevicePostureReportOrderField("")
	_ encoding.TextMarshaler   = DevicePostureReportOrderField("")
	_ encoding.TextUnmarshaler = (*DevicePostureReportOrderField)(nil)
)

func DevicePostureReportOrderFields() []DevicePostureReportOrderField {
	return []DevicePostureReportOrderField{
		DevicePostureReportOrderFieldCreatedAt,
	}
}

func (v DevicePostureReportOrderField) IsValid() bool {
	return isValidOrderField(v, DevicePostureReportOrderFields())
}

func (v DevicePostureReportOrderField) String() string {
	return string(v)
}

func (v DevicePostureReportOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DevicePostureReportOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, DevicePostureReportOrderFields())
}

func (f DevicePostureReportOrderField) Column() string {
	switch f {
	case DevicePostureReportOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", f))
}
