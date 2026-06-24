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
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"go.probo.inc/probo/pkg/llm"
)

type TypedResult[T any] struct {
	Result
	Output T
}

func RunTyped[T any](
	ctx context.Context,
	a *Agent,
	messages []llm.Message,
) (*TypedResult[T], error) {
	typed := a.clone()

	schema, err := jsonSchemaFor[T]()
	if err != nil {
		return nil, fmt.Errorf("cannot create typed runner: %w", err)
	}

	typed.responseFormat = &llm.ResponseFormat{
		Type: llm.ResponseFormatJSONSchema,
		JSONSchema: &llm.JSONSchema{
			Name:   typeName[T](),
			Schema: schema,
			Strict: true,
		},
	}

	result, err := typed.Run(ctx, messages)
	if err != nil {
		return nil, err
	}

	text := result.FinalMessage().Text()

	var output T
	if err := json.Unmarshal([]byte(text), &output); err != nil {
		return nil, fmt.Errorf("cannot parse typed output: %w", err)
	}

	return &TypedResult[T]{
		Result: *result,
		Output: output,
	}, nil
}

func typeName[T any]() string {
	t := reflect.TypeFor[T]()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	name := t.Name()
	if name == "" {
		name = "output"
	}

	return strings.ToLower(name)
}
