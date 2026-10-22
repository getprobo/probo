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

package bigint

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
)

type BigIntScalar = int64

func MarshalBigIntScalar(i int64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = w.Write([]byte(strconv.FormatInt(i, 10)))
	})
}

func UnmarshalBigIntScalar(v any) (int64, error) {
	switch val := v.(type) {
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid BigInt value: %v", err)
		}

		return i, nil
	case int:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case float32:
		if val != float32(int64(val)) {
			return 0, fmt.Errorf("BigInt cannot represent non-integer value: %v", val)
		}

		return int64(val), nil
	case float64:
		if val != float64(int64(val)) {
			return 0, fmt.Errorf("BigInt cannot represent non-integer value: %v", val)
		}

		return int64(val), nil
	default:
		return 0, fmt.Errorf("cannot unmarshal %T into BigInt", v)
	}
}
