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

package types

import (
	"time"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

type (
	ProbotIdentityBinding struct {
		ID               gid.GID   `json:"id"`
		Provider         string    `json:"provider"`
		ExternalTenantID string    `json:"externalTenantId"`
		ExternalUserID   string    `json:"externalUserId"`
		CreatedAt        time.Time `json:"createdAt"`
		UpdatedAt        time.Time `json:"updatedAt"`
	}

	ProbotIdentityBindPreview struct {
		Provider           string `json:"provider"`
		ExternalTenantID   string `json:"externalTenantId"`
		ExternalUserID     string `json:"externalUserId"`
		ExternalTenantName string `json:"externalTenantName"`
		ExternalUserName   string `json:"externalUserName"`
	}
)

func NewProbotIdentityBinding(
	binding *identitybinding.Binding,
) *ProbotIdentityBinding {
	return &ProbotIdentityBinding{
		ID:               binding.ID,
		Provider:         binding.Provider,
		ExternalTenantID: binding.ExternalTenantID,
		ExternalUserID:   binding.ExternalUserID,
		CreatedAt:        binding.CreatedAt,
		UpdatedAt:        binding.UpdatedAt,
	}
}

func NewProbotIdentityBindings(
	bindings []*identitybinding.Binding,
) []*ProbotIdentityBinding {
	result := make([]*ProbotIdentityBinding, len(bindings))
	for i, binding := range bindings {
		result[i] = NewProbotIdentityBinding(binding)
	}

	return result
}

func (b ProbotIdentityBinding) GetID() gid.GID {
	return b.ID
}

func NewProbotIdentityBindPreview(
	subject identitybinding.Subject,
) *ProbotIdentityBindPreview {
	return &ProbotIdentityBindPreview{
		Provider:           subject.Provider,
		ExternalTenantID:   subject.ExternalTenantID,
		ExternalUserID:     subject.ExternalUserID,
		ExternalTenantName: subject.ExternalTenantName,
		ExternalUserName:   subject.ExternalUserName,
	}
}
