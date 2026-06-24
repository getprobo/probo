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
