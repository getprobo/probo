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

	"go.probo.inc/probo/pkg/gid"
)

type ErrSCIMConfigurationNotFound struct {
	ID gid.GID
}

func (e *ErrSCIMConfigurationNotFound) Error() string {
	return fmt.Sprintf("SCIM configuration %s not found", e.ID)
}

func NewSCIMConfigurationNotFoundError(id gid.GID) *ErrSCIMConfigurationNotFound {
	return &ErrSCIMConfigurationNotFound{ID: id}
}

type ErrSCIMConfigurationAlreadyExists struct {
	OrganizationID gid.GID
}

func (e *ErrSCIMConfigurationAlreadyExists) Error() string {
	return fmt.Sprintf("SCIM configuration already exists for organization %s", e.OrganizationID)
}

func NewSCIMConfigurationAlreadyExistsError(organizationID gid.GID) *ErrSCIMConfigurationAlreadyExists {
	return &ErrSCIMConfigurationAlreadyExists{OrganizationID: organizationID}
}

type ErrSCIMInvalidToken struct{}

func (e *ErrSCIMInvalidToken) Error() string {
	return "invalid SCIM bearer token"
}

func NewSCIMInvalidTokenError() *ErrSCIMInvalidToken {
	return &ErrSCIMInvalidToken{}
}
