// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package session

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

type SessionRequirement string

const (
	SessionRequirementPresent  SessionRequirement = "PRESENT"
	SessionRequirementNone     SessionRequirement = "NONE"
	SessionRequirementOptional SessionRequirement = "OPTIONAL"
)

var AllSessionRequirement = []SessionRequirement{
	SessionRequirementPresent,
	SessionRequirementNone,
	SessionRequirementOptional,
}

func (e SessionRequirement) IsValid() bool {
	switch e {
	case SessionRequirementPresent, SessionRequirementNone, SessionRequirementOptional:
		return true
	}
	return false
}

func (e SessionRequirement) String() string {
	return string(e)
}

func (e *SessionRequirement) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SessionRequirement(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SessionRequirement", str)
	}
	return nil
}

func (e SessionRequirement) MarshalGQL(w io.Writer) {
	_, _ = fmt.Fprint(w, strconv.Quote(e.String()))
}

func (e *SessionRequirement) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	return e.UnmarshalGQL(s)
}

func (e SessionRequirement) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	e.MarshalGQL(&buf)
	return buf.Bytes(), nil
}

func Directive(ctx context.Context, obj any, next graphql.Resolver, required SessionRequirement) (any, error) {
	identity := authn.IdentityFromContext(ctx)

	switch required {
	case SessionRequirementOptional:
	case SessionRequirementPresent:
		if identity == nil {
			return nil, gqlutils.Unauthenticatedf(
				ctx,
				"authentication is required to access this resouce",
			)
		}
	case SessionRequirementNone:
		if identity != nil {
			return nil, gqlutils.AlreadyAuthenticatedf(
				ctx,
				"authentication not allowed for this resource/action",
			)
		}
	}

	return next(ctx)
}
