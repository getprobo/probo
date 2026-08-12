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

package coredata

import (
	"encoding"
	"fmt"
)

type AiSystemCompanyRole string

const (
	AiSystemCompanyRoleProvider  AiSystemCompanyRole = "PROVIDER"
	AiSystemCompanyRoleDeployer  AiSystemCompanyRole = "DEPLOYER"
	AiSystemCompanyRoleUser      AiSystemCompanyRole = "USER"
	AiSystemCompanyRoleDeveloper AiSystemCompanyRole = "DEVELOPER"
)

var (
	_ fmt.Stringer             = AiSystemCompanyRole("")
	_ encoding.TextMarshaler   = AiSystemCompanyRole("")
	_ encoding.TextUnmarshaler = (*AiSystemCompanyRole)(nil)
)

func AiSystemCompanyRoles() []AiSystemCompanyRole {
	return []AiSystemCompanyRole{
		AiSystemCompanyRoleProvider,
		AiSystemCompanyRoleDeployer,
		AiSystemCompanyRoleUser,
		AiSystemCompanyRoleDeveloper,
	}
}

func (v AiSystemCompanyRole) IsValid() bool {
	switch v {
	case
		AiSystemCompanyRoleProvider,
		AiSystemCompanyRoleDeployer,
		AiSystemCompanyRoleUser,
		AiSystemCompanyRoleDeveloper:
		return true
	}

	return false
}

func (v AiSystemCompanyRole) String() string {
	return string(v)
}

func (v AiSystemCompanyRole) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AiSystemCompanyRole) UnmarshalText(text []byte) error {
	val := AiSystemCompanyRole(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AiSystemCompanyRole value: %q", string(text))
	}

	*v = val

	return nil
}
