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

package agent

import (
	"encoding/json"
	"fmt"

	"go.probo.inc/probo/pkg/llm"
)

// OutputType describes a structured output schema that the agent should
// produce. Build one with NewOutputType[T]().
type OutputType struct {
	Name   string
	Schema json.RawMessage
}

func NewOutputType[T any](name string) (*OutputType, error) {
	schema, err := jsonSchemaFor[T]()
	if err != nil {
		return nil, fmt.Errorf("cannot create output type %q: %w", name, err)
	}

	return &OutputType{
		Name:   name,
		Schema: schema,
	}, nil
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
