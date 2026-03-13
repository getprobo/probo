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
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	name := t.Name()
	if name == "" {
		name = "output"
	}

	return strings.ToLower(name)
}
