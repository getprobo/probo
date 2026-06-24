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

package scim

import (
	"fmt"
	"strings"

	scimerrors "github.com/elimity-com/scim/errors"
	scimfilter "github.com/scim2/filter-parser/v2"
	"go.probo.inc/probo/pkg/coredata"
)

func ParseUserFilter(expr scimfilter.Expression) (*coredata.MembershipProfileFilter, error) {
	filter := coredata.NewMembershipProfileFilter(nil).WithMembership()

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
				filter.WithUserName(value)
			case "externalid":
				filter.WithExternalID(value)
			default:
				return nil, scimerrors.ScimErrorBadRequest(
					fmt.Sprintf("attribute '%s' is not supported for filtering, only 'userName' and 'externalId' are supported", e.AttributePath.AttributeName))
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
