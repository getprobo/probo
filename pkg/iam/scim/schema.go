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
						schema.SimpleStringParams(
							schema.StringParams{
								Name: "formatted",
							},
						),
						schema.SimpleStringParams(
							schema.StringParams{
								Name: "familyName",
							},
						),
						schema.SimpleStringParams(
							schema.StringParams{
								Name: "givenName",
							},
						),
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
						schema.SimpleStringParams(
							schema.StringParams{
								Name: "value",
							},
						),
						schema.SimpleStringParams(
							schema.StringParams{
								Name: "type",
							},
						),
						schema.SimpleBooleanParams(
							schema.BooleanParams{
								Name: "primary",
							},
						),
					},
				},
			),
		},
	}
}
