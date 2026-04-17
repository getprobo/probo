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

package probod

import (
	"context"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agents/vetting"
	"go.probo.inc/probo/pkg/probo"
)

// stubVendorAssessor satisfies probo.VendorAssessor with a deterministic
// result. It exists so end-to-end tests can exercise the assessVendor
// GraphQL mutation, MCP tool and CLI command without running the full
// LLM and browser pipeline.
type stubVendorAssessor struct{}

var _ probo.VendorAssessor = (*stubVendorAssessor)(nil)

func (stubVendorAssessor) Assess(
	_ context.Context,
	websiteURL string,
	_ string,
	_ agent.ProgressReporter,
) (*vetting.Result, error) {
	return &vetting.Result{
		Document: "stub vendor assessment report",
		Info: vetting.VendorInfo{
			Name:             "Stub Vendor",
			Description:      "Deterministic stub output used in e2e tests",
			Category:         "OTHER",
			VendorType:       "SAAS",
			PrivacyPolicyURL: websiteURL + "/privacy",
			Subprocessors: []vetting.Subprocessor{
				{
					Name:    "Example Subprocessor",
					Country: "United States",
					Purpose: "Data processing",
				},
			},
		},
	}, nil
}
