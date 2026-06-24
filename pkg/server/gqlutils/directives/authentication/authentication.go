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

package authentication

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

type AuthenticationRequirement string

const (
	AuthenticationRequirementPresent  AuthenticationRequirement = "PRESENT"
	AuthenticationRequirementNone     AuthenticationRequirement = "NONE"
	AuthenticationRequirementOptional AuthenticationRequirement = "OPTIONAL"
)

var AllAuthenticationRequirement = []AuthenticationRequirement{
	AuthenticationRequirementPresent,
	AuthenticationRequirementNone,
	AuthenticationRequirementOptional,
}

func (e AuthenticationRequirement) IsValid() bool {
	switch e {
	case AuthenticationRequirementPresent, AuthenticationRequirementNone, AuthenticationRequirementOptional:
		return true
	}

	return false
}

func (e AuthenticationRequirement) String() string {
	return string(e)
}

func (e *AuthenticationRequirement) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AuthenticationRequirement(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AuthenticationRequirement", str)
	}

	return nil
}

func (e AuthenticationRequirement) MarshalGQL(w io.Writer) {
	_, _ = fmt.Fprint(w, strconv.Quote(e.String()))
}

func (e *AuthenticationRequirement) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	return e.UnmarshalGQL(s)
}

func (e AuthenticationRequirement) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	e.MarshalGQL(&buf)

	return buf.Bytes(), nil
}

func Directive(ctx context.Context, obj any, next graphql.Resolver, required AuthenticationRequirement) (any, error) {
	identity := authn.IdentityFromContext(ctx)

	switch required {
	case AuthenticationRequirementOptional:
	case AuthenticationRequirementPresent:
		if identity == nil {
			return nil, gqlutils.Unauthenticatedf(
				ctx,
				"authentication is required to access this resource",
			)
		}
	case AuthenticationRequirementNone:
		if identity != nil {
			return nil, gqlutils.AlreadyAuthenticatedf(
				ctx,
				"authentication not allowed for this resource/action",
			)
		}
	}

	return next(ctx)
}
