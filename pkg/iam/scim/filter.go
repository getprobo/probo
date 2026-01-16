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

package scim

import (
	"fmt"
	"strings"

	scimerrors "github.com/elimity-com/scim/errors"
	scimfilter "github.com/scim2/filter-parser/v2"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/mail"
)

func ParseUserFilter(expr scimfilter.Expression) (*coredata.MembershipFilter, error) {
	filter := coredata.NewMembershipFilter()

	if expr == nil {
		return filter, nil
	}

	stack := []scimfilter.Expression{expr}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		switch e := current.(type) {
		case *scimfilter.AttributeExpression:
			if e.Operator != scimfilter.EQ {
				return nil, scimerrors.ScimErrorBadRequest(
					fmt.Sprintf("operator '%s' is not supported, only 'eq' is supported", e.Operator))
			}

			value, ok := e.CompareValue.(string)
			if !ok {
				return nil, scimerrors.ScimErrorBadRequest("filter value must be a string")
			}

			attrName := strings.ToLower(e.AttributePath.AttributeName)
			switch attrName {
			case "username":
				email, err := mail.ParseAddr(value)
				if err != nil {
					return nil, scimerrors.ScimErrorBadRequest(
						fmt.Sprintf("invalid email format for userName: %s", value))
				}
				filter.WithEmail(&email)
			default:
				return nil, scimerrors.ScimErrorBadRequest(
					fmt.Sprintf("attribute '%s' is not supported for filtering, only 'userName' is supported", e.AttributePath.AttributeName))
			}

		case *scimfilter.LogicalExpression:
			if e.Operator != scimfilter.AND {
				return nil, scimerrors.ScimErrorBadRequest(
					fmt.Sprintf("logical operator '%s' is not supported, only 'and' is supported", e.Operator))
			}
			stack = append(stack, e.Left, e.Right)

		case *scimfilter.NotExpression:
			return nil, scimerrors.ScimErrorBadRequest("NOT expressions are not supported")

		case *scimfilter.ValuePath:
			return nil, scimerrors.ScimErrorBadRequest("value path expressions are not supported")

		default:
			return nil, scimerrors.ScimErrorBadRequest("unknown filter expression type")
		}
	}

	return filter, nil
}
