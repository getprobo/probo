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

package page

import (
	"fmt"
)

type OrderDirection string

const (
	OrderDirectionAsc  OrderDirection = "ASC"
	OrderDirectionDesc OrderDirection = "DESC"
)

func (od OrderDirection) MarshalText() ([]byte, error) {
	return []byte(od.String()), nil
}

func (od *OrderDirection) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case OrderDirectionAsc.String():
		*od = OrderDirectionAsc
	case OrderDirectionDesc.String():
		*od = OrderDirectionDesc
	default:
		return fmt.Errorf("invalid OrderDirection value: %q", val)
	}

	return nil
}

func (od OrderDirection) String() string {
	return string(od)
}
