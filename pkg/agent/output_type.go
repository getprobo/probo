// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package agent

import (
	"encoding/json"

	"go.probo.inc/probo/pkg/llm"
)

// OutputType describes a structured output schema that the agent should
// produce. Build one with NewOutputType[T]().
type OutputType struct {
	Name   string
	Schema json.RawMessage
}

func NewOutputType[T any](name string) *OutputType {
	return &OutputType{
		Name:   name,
		Schema: jsonSchemaFor[T](),
	}
}

func (o *OutputType) responseFormat() *llm.ResponseFormat {
	return &llm.ResponseFormat{
		Type: llm.ResponseFormatJSONSchema,
		JSONSchema: &llm.JSONSchema{
			Name:   o.Name,
			Schema: o.Schema,
			Strict: true,
		},
	}
}
