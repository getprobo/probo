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

package webhook

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

var (
	jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

func InsertEvent(
	ctx context.Context,
	conn pg.Conn,
	scope coredata.Scoper,
	organizationID gid.GID,
	eventType coredata.WebhookEventType,
	data any,
) error {
	var configs coredata.WebhookConfigurations
	exists, err := configs.ExistsByOrganizationIDAndEventType(ctx, conn, scope, organizationID, eventType)
	if err != nil {
		return fmt.Errorf("cannot check webhook configurations: %w", err)
	}

	if !exists {
		return nil
	}

	raw, err := MarshalData(data)
	if err != nil {
		return fmt.Errorf("cannot marshal webhook event data: %w", err)
	}

	event := &coredata.WebhookEvent{
		ID:             gid.New(scope.GetTenantID(), coredata.WebhookEventEntityType),
		OrganizationID: organizationID,
		EventType:      eventType,
		Status:         coredata.WebhookEventStatusPending,
		Data:           raw,
		CreatedAt:      time.Now(),
	}

	if err = event.Insert(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot insert webhook event: %w", err)
	}

	return nil
}

func MarshalData(v any) (json.RawMessage, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal webhook data: %w", err)
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("cannot unmarshal webhook data: %w", err)
	}

	for _, key := range nestedFieldKeys(v) {
		delete(m, key)
	}

	delete(m, "permission")

	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("cannot re-marshal webhook data: %w", err)
	}

	return data, nil
}

func nestedFieldKeys(v any) []string {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	var keys []string
	for i := range t.NumField() {
		field := t.Field(i)

		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		jsonKey, _, _ := strings.Cut(tag, ",")

		if isNestedType(field.Type) {
			keys = append(keys, jsonKey)
		}
	}

	return keys
}

func isNestedType(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() == reflect.Slice {
		return isNestedType(t.Elem())
	}

	if t.Kind() != reflect.Struct {
		return false
	}

	ptrType := reflect.PointerTo(t)

	if t.Implements(jsonMarshalerType) || ptrType.Implements(jsonMarshalerType) {
		return false
	}

	if t.Implements(textMarshalerType) || ptrType.Implements(textMarshalerType) {
		return false
	}

	return true
}
