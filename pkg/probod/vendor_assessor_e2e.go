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

//go:build e2e

package probod

import (
	"context"
	"fmt"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agents/vetting"
	"go.probo.inc/probo/pkg/probo"
)

// buildVendorAssessor returns a deterministic stub. It replaces the real
// LLM/browser-driven assessor only in binaries built with `-tags=e2e` so
// the production binary can never serve this fixture.
func (impl *Implm) buildVendorAssessor(
	_ *log.Logger,
	_ trace.TracerProvider,
	_ prometheus.Registerer,
) (probo.VendorAssessor, error) {
	return stubVendorAssessor{}, nil
}

type stubVendorAssessor struct{}

var _ probo.VendorAssessor = (*stubVendorAssessor)(nil)

func (stubVendorAssessor) Assess(
	_ context.Context,
	websiteURL string,
	_ string,
	_ agent.ProgressReporter,
) (*vetting.Result, error) {
	privacyURL, err := url.JoinPath(websiteURL, "privacy")
	if err != nil {
		return nil, fmt.Errorf("cannot build stub privacy URL: %w", err)
	}

	return &vetting.Result{
		Document: "stub vendor assessment report",
		Info: vetting.VendorInfo{
			Name:             "Stub Vendor",
			Description:      "Deterministic stub output used in e2e tests",
			Category:         "OTHER",
			VendorType:       "SAAS",
			PrivacyPolicyURL: privacyURL,
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
