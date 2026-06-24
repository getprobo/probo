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
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

func UserSchema() schema.Schema {
	return schema.Schema{
		ID:          schema.UserSchema,
		Name:        optional.NewString("User"),
		Description: optional.NewString("User Account"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(
					schema.StringParams{
						Name:       "userName",
						Required:   true,
						Uniqueness: schema.AttributeUniquenessServer(),
					},
				),
			),
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(
					schema.StringParams{
						Name: "displayName",
					},
				),
			),
			schema.ComplexCoreAttribute(
				schema.ComplexParams{
					Name: "name",
					SubAttributes: []schema.SimpleParams{
						schema.SimpleStringParams(schema.StringParams{Name: "formatted"}),
						schema.SimpleStringParams(schema.StringParams{Name: "familyName"}),
						schema.SimpleStringParams(schema.StringParams{Name: "givenName"}),
						schema.SimpleStringParams(schema.StringParams{Name: "middleName"}),
						schema.SimpleStringParams(schema.StringParams{Name: "honorificPrefix"}),
						schema.SimpleStringParams(schema.StringParams{Name: "honorificSuffix"}),
					},
				},
			),
			schema.SimpleCoreAttribute(
				schema.SimpleBooleanParams(
					schema.BooleanParams{
						Name: "active",
					},
				),
			),
			schema.ComplexCoreAttribute(
				schema.ComplexParams{
					Name:        "emails",
					MultiValued: true,
					SubAttributes: []schema.SimpleParams{
						schema.SimpleStringParams(schema.StringParams{Name: "value"}),
						schema.SimpleStringParams(schema.StringParams{Name: "type"}),
						schema.SimpleBooleanParams(schema.BooleanParams{Name: "primary"}),
						schema.SimpleStringParams(schema.StringParams{Name: "display"}),
					},
				},
			),
			schema.ComplexCoreAttribute(
				schema.ComplexParams{
					Name:        "phoneNumbers",
					MultiValued: true,
					SubAttributes: []schema.SimpleParams{
						schema.SimpleStringParams(schema.StringParams{Name: "value"}),
						schema.SimpleStringParams(schema.StringParams{Name: "type"}),
						schema.SimpleBooleanParams(schema.BooleanParams{Name: "primary"}),
						schema.SimpleStringParams(schema.StringParams{Name: "display"}),
					},
				},
			),
			schema.ComplexCoreAttribute(
				schema.ComplexParams{
					Name:        "addresses",
					MultiValued: true,
					SubAttributes: []schema.SimpleParams{
						schema.SimpleStringParams(schema.StringParams{Name: "formatted"}),
						schema.SimpleStringParams(schema.StringParams{Name: "streetAddress"}),
						schema.SimpleStringParams(schema.StringParams{Name: "locality"}),
						schema.SimpleStringParams(schema.StringParams{Name: "region"}),
						schema.SimpleStringParams(schema.StringParams{Name: "postalCode"}),
						schema.SimpleStringParams(schema.StringParams{Name: "country"}),
						schema.SimpleStringParams(schema.StringParams{Name: "type"}),
						schema.SimpleBooleanParams(schema.BooleanParams{Name: "primary"}),
					},
				},
			),
			schema.ComplexCoreAttribute(
				schema.ComplexParams{
					Name:        "roles",
					MultiValued: true,
					SubAttributes: []schema.SimpleParams{
						schema.SimpleStringParams(schema.StringParams{Name: "value"}),
						schema.SimpleStringParams(schema.StringParams{Name: "display"}),
						schema.SimpleStringParams(schema.StringParams{Name: "type"}),
						schema.SimpleBooleanParams(schema.BooleanParams{Name: "primary"}),
					},
				},
			),
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(schema.StringParams{Name: "title"}),
			),
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(schema.StringParams{Name: "userType"}),
			),
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(schema.StringParams{Name: "nickName"}),
			),
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(schema.StringParams{Name: "preferredLanguage"}),
			),
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(schema.StringParams{Name: "locale"}),
			),
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(schema.StringParams{Name: "timezone"}),
			),
			schema.SimpleCoreAttribute(
				schema.SimpleReferenceParams(schema.ReferenceParams{Name: "profileUrl"}),
			),
		},
	}
}

func EnterpriseUserSchema() schema.Schema {
	return schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		Name:        optional.NewString("EnterpriseUser"),
		Description: optional.NewString("Enterprise User Extension"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(schema.StringParams{Name: "employeeNumber"}),
			),
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(schema.StringParams{Name: "costCenter"}),
			),
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(schema.StringParams{Name: "organization"}),
			),
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(schema.StringParams{Name: "division"}),
			),
			schema.SimpleCoreAttribute(
				schema.SimpleStringParams(schema.StringParams{Name: "department"}),
			),
			schema.ComplexCoreAttribute(
				schema.ComplexParams{
					Name: "manager",
					SubAttributes: []schema.SimpleParams{
						schema.SimpleStringParams(schema.StringParams{Name: "value"}),
						schema.SimpleReferenceParams(schema.ReferenceParams{Name: "$ref"}),
						schema.SimpleStringParams(schema.StringParams{Name: "displayName"}),
					},
				},
			),
		},
	}
}
