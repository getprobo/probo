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

package coredata

import "fmt"

type ErrResourceNotFound struct {
	Resource   string
	Identifier string
}

func (e ErrResourceNotFound) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.Identifier)
}

type ErrResourceAlreadyExists struct {
	Resource string
	Message  string
}

func (e ErrResourceAlreadyExists) Error() string {
	msg := fmt.Sprintf("%s already exists", e.Resource)

	if e.Message != "" {
		msg += ": " + e.Message
	}

	return msg
}

type ErrRestrictedOperation struct {
	Resource string
	Message  string
}

func (e ErrRestrictedOperation) Error() string {
	msg := fmt.Sprintf("restricted operation on %s", e.Resource)

	if e.Message != "" {
		msg += ": " + e.Message
	}

	return msg
}

type ErrInvalidValue struct {
	Field   string
	Value   string
	Message string
}

func (e ErrInvalidValue) Error() string {
	var msg string
	if e.Value != "" {
		msg = fmt.Sprintf("invalid value %s for %s", e.Value, e.Field)
	} else {
		msg = fmt.Sprintf("invalid value for %s", e.Field)
	}

	if e.Message != "" {
		msg += ": " + e.Message
	}

	return msg
}

type ErrNoChange struct {
	Resource string
	Message  string
}

func (e ErrNoChange) Error() string {
	msg := fmt.Sprintf("no changes detected for %s", e.Resource)

	if e.Message != "" {
		msg += ": " + e.Message
	}

	return msg
}
