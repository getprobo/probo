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

package oauth2scope

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"go.probo.inc/probo/pkg/coredata"
)

type OAuth2ScopeScalar = coredata.OAuth2Scope

func MarshalOAuth2ScopeScalar(s OAuth2ScopeScalar) graphql.Marshaler {
	return graphql.WriterFunc(
		func(w io.Writer) {
			_, _ = w.Write([]byte(strconv.Quote(s.String())))
		},
	)
}

func UnmarshalOAuth2ScopeScalar(v any) (OAuth2ScopeScalar, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("oauth2 scope must be a string")
	}

	return OAuth2ScopeScalar(s), nil
}
