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
			return nil, &ErrUnsupportedFilter{Reason: fmt.Sprintf("logical operator '%s' is not supported, only 'and' is supported", e.Operator)}
		}
		if err := parseLogicalExpression(e, filter); err != nil {
			return nil, err
		}
	case *scimfilter.NotExpression:
		return nil, &ErrUnsupportedFilter{Reason: "NOT expressions are not supported"}
	case *scimfilter.ValuePath:
		return nil, &ErrUnsupportedFilter{Reason: "value path expressions are not supported"}
	default:
		return nil, &ErrUnsupportedFilter{Reason: "unknown filter expression type"}
	}

	return filter, nil
}

func parseAttributeExpression(e *scimfilter.AttributeExpression, filter *UserFilter) error {
	// Only support "eq" operator
	if e.Operator != scimfilter.EQ {
		return &ErrUnsupportedFilter{Reason: fmt.Sprintf("operator '%s' is not supported, only 'eq' is supported", e.Operator)}
	}

	// Get the attribute name (lowercase for comparison)
	attrName := strings.ToLower(e.AttributePath.AttributeName)

	// Extract the string value
	value, ok := e.CompareValue.(string)
	if !ok {
		return &ErrUnsupportedFilter{Reason: "filter value must be a string"}
	}

	switch attrName {
	case "username":
		filter.UserName = &value
	default:
		return &ErrUnsupportedFilter{Reason: fmt.Sprintf("attribute '%s' is not supported for filtering, only 'userName' is supported", e.AttributePath.AttributeName)}
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
		return &ErrUnsupportedFilter{Reason: "nested logical expressions are not supported"}
	}

	// Process right expression
	if right, ok := e.Right.(*scimfilter.AttributeExpression); ok {
		if err := parseAttributeExpression(right, filter); err != nil {
			return err
		}
	} else {
		return &ErrUnsupportedFilter{Reason: "nested logical expressions are not supported"}
	}

	return nil
}

// SCIM 2.0 User Resource
// https://datatracker.ietf.org/doc/html/rfc7643#section-4.1
type User struct {
	Schemas     []string `json:"schemas"`
	ID          string   `json:"id,omitempty"`
	ExternalID  string   `json:"externalId,omitempty"`
	UserName    string   `json:"userName"`
	Name        *Name    `json:"name,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
	Emails      []Email  `json:"emails,omitempty"`
	Active      *bool    `json:"active,omitempty"`
	Meta        *Meta    `json:"meta,omitempty"`
}

type Name struct {
	Formatted       string `json:"formatted,omitempty"`
	FamilyName      string `json:"familyName,omitempty"`
	GivenName       string `json:"givenName,omitempty"`
	MiddleName      string `json:"middleName,omitempty"`
	HonorificPrefix string `json:"honorificPrefix,omitempty"`
	HonorificSuffix string `json:"honorificSuffix,omitempty"`
}

type Email struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

type Meta struct {
	ResourceType string `json:"resourceType,omitempty"`
	Created      string `json:"created,omitempty"`
	LastModified string `json:"lastModified,omitempty"`
	Location     string `json:"location,omitempty"`
	Version      string `json:"version,omitempty"`
}

// SCIM 2.0 List Response
// https://datatracker.ietf.org/doc/html/rfc7644#section-3.4.2
type ListResponse struct {
	Schemas      []string `json:"schemas"`
	TotalResults int      `json:"totalResults"`
	StartIndex   int      `json:"startIndex,omitempty"`
	ItemsPerPage int      `json:"itemsPerPage,omitempty"`
	Resources    []User   `json:"Resources"`
}

// SCIM 2.0 Error Response
// https://datatracker.ietf.org/doc/html/rfc7644#section-3.12
type ErrorResponse struct {
	Schemas  []string `json:"schemas"`
	Detail   string   `json:"detail,omitempty"`
	Status   string   `json:"status"`
	ScimType string   `json:"scimType,omitempty"`
}

// SCIM 2.0 Patch Operation
// https://datatracker.ietf.org/doc/html/rfc7644#section-3.5.2
type PatchOp struct {
	Schemas    []string    `json:"schemas"`
	Operations []Operation `json:"Operations"`
}

type Operation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// SCIM 2.0 Service Provider Config
// https://datatracker.ietf.org/doc/html/rfc7643#section-5
type ServiceProviderConfig struct {
	Schemas               []string        `json:"schemas"`
	DocumentationUri      string          `json:"documentationUri,omitempty"`
	Patch                 Supported       `json:"patch"`
	Bulk                  BulkSupported   `json:"bulk"`
	Filter                FilterSupported `json:"filter"`
	ChangePassword        Supported       `json:"changePassword"`
	Sort                  Supported       `json:"sort"`
	Etag                  Supported       `json:"etag"`
	AuthenticationSchemes []AuthScheme    `json:"authenticationSchemes"`
	Meta                  *Meta           `json:"meta,omitempty"`
}

type Supported struct {
	Supported bool `json:"supported"`
}

type BulkSupported struct {
	Supported      bool `json:"supported"`
	MaxOperations  int  `json:"maxOperations"`
	MaxPayloadSize int  `json:"maxPayloadSize"`
}

type FilterSupported struct {
	Supported  bool `json:"supported"`
	MaxResults int  `json:"maxResults"`
}

type AuthScheme struct {
	Type             string `json:"type"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	SpecUri          string `json:"specUri,omitempty"`
	DocumentationUri string `json:"documentationUri,omitempty"`
	Primary          bool   `json:"primary,omitempty"`
}

// SCIM 2.0 Schemas response
type SchemasResponse struct {
	Schemas      []string `json:"schemas"`
	TotalResults int      `json:"totalResults"`
	Resources    []Schema `json:"Resources"`
}

type Schema struct {
	ID          string            `json:"id"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Attributes  []SchemaAttribute `json:"attributes,omitempty"`
	Meta        *Meta             `json:"meta,omitempty"`
}

type SchemaAttribute struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	MultiValued   bool              `json:"multiValued"`
	Description   string            `json:"description,omitempty"`
	Required      bool              `json:"required"`
	CaseExact     bool              `json:"caseExact,omitempty"`
	Mutability    string            `json:"mutability,omitempty"`
	Returned      string            `json:"returned,omitempty"`
	Uniqueness    string            `json:"uniqueness,omitempty"`
	SubAttributes []SchemaAttribute `json:"subAttributes,omitempty"`
}

// SCIM Schema URIs
const (
	SchemaURIUser                  = "urn:ietf:params:scim:schemas:core:2.0:User"
	SchemaURIListResponse          = "urn:ietf:params:scim:api:messages:2.0:ListResponse"
	SchemaURIError                 = "urn:ietf:params:scim:api:messages:2.0:Error"
	SchemaURIPatchOp               = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
	SchemaURIServiceProviderConfig = "urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"
	SchemaURISchema                = "urn:ietf:params:scim:schemas:core:2.0:Schema"
)

func NewUser() *User {
	return &User{
		Schemas: []string{SchemaURIUser},
	}
}

func NewListResponse(users []User, totalResults int) *ListResponse {
	return &ListResponse{
		Schemas:      []string{SchemaURIListResponse},
		TotalResults: totalResults,
		StartIndex:   1,
		ItemsPerPage: len(users),
		Resources:    users,
	}
}

func NewErrorResponse(status int, detail string, scimType string) *ErrorResponse {
	return &ErrorResponse{
		Schemas:  []string{SchemaURIError},
		Detail:   detail,
		Status:   fmt.Sprintf("%d", status),
		ScimType: scimType,
	}
}

func (u *User) GetPrimaryEmail() string {
	for _, email := range u.Emails {
		if email.Primary {
			return email.Value
		}
	}
	if len(u.Emails) > 0 {
		return u.Emails[0].Value
	}
	return u.UserName
}

func (u *User) GetFullName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.Name != nil {
		if u.Name.Formatted != "" {
			return u.Name.Formatted
		}
		parts := []string{}
		if u.Name.GivenName != "" {
			parts = append(parts, u.Name.GivenName)
		}
		if u.Name.FamilyName != "" {
			parts = append(parts, u.Name.FamilyName)
		}
		if len(parts) > 0 {
			return join(parts, " ")
		}
	}
	return u.UserName
}

func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
