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
)

// UserFilter represents filter criteria for listing SCIM users
type UserFilter struct {
	// UserName filters by userName (email) with exact match
	UserName *string
}

// ParseUserFilter converts a SCIM filter AST expression to a UserFilter.
// Returns (nil, nil) if no filter is provided.
// Returns an error if the filter uses unsupported operators or attributes.
func ParseUserFilter(expr scimfilter.Expression) (*UserFilter, error) {
	if expr == nil {
		return nil, nil
	}

	filter := &UserFilter{}

	switch e := expr.(type) {
	case *scimfilter.AttributeExpression:
		if err := parseAttributeExpression(e, filter); err != nil {
			return nil, err
		}
	case *scimfilter.LogicalExpression:
		if e.Operator != scimfilter.AND {
			return nil, scimerrors.ScimErrorBadRequest(fmt.Sprintf("logical operator '%s' is not supported, only 'and' is supported", e.Operator))
		}
		if err := parseLogicalExpression(e, filter); err != nil {
			return nil, err
		}
	case *scimfilter.NotExpression:
		return nil, scimerrors.ScimErrorBadRequest("NOT expressions are not supported")
	case *scimfilter.ValuePath:
		return nil, scimerrors.ScimErrorBadRequest("value path expressions are not supported")
	default:
		return nil, scimerrors.ScimErrorBadRequest("unknown filter expression type")
	}

	return filter, nil
}

func parseAttributeExpression(e *scimfilter.AttributeExpression, filter *UserFilter) error {
	// Only support "eq" operator
	if e.Operator != scimfilter.EQ {
		return scimerrors.ScimErrorBadRequest(fmt.Sprintf("operator '%s' is not supported, only 'eq' is supported", e.Operator))
	}

	// Get the attribute name (lowercase for comparison)
	attrName := strings.ToLower(e.AttributePath.AttributeName)

	// Extract the string value
	value, ok := e.CompareValue.(string)
	if !ok {
		return scimerrors.ScimErrorBadRequest("filter value must be a string")
	}

	switch attrName {
	case "username":
		filter.UserName = &value
	default:
		return scimerrors.ScimErrorBadRequest(fmt.Sprintf("attribute '%s' is not supported for filtering, only 'userName' is supported", e.AttributePath.AttributeName))
	}

	return nil
}

func parseLogicalExpression(e *scimfilter.LogicalExpression, filter *UserFilter) error {
	// Process left expression
	if left, ok := e.Left.(*scimfilter.AttributeExpression); ok {
		if err := parseAttributeExpression(left, filter); err != nil {
			return err
		}
	} else {
		return scimerrors.ScimErrorBadRequest("nested logical expressions are not supported")
	}

	// Process right expression
	if right, ok := e.Right.(*scimfilter.AttributeExpression); ok {
		if err := parseAttributeExpression(right, filter); err != nil {
			return err
		}
	} else {
		return scimerrors.ScimErrorBadRequest("nested logical expressions are not supported")
	}

	return nil
}

// User represents parsed SCIM user attributes with extracted values.
type User struct {
	Email    string
	FullName string
	Active   *bool
}
