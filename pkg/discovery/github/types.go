// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package github

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	// StartAgentName is stored on agent_runs.start_agent_name for discovery jobs.
	StartAgentName = "github-discovery"

	thirdPartyName = "GitHub"

	maxMeasureCreatesPerRun = 50
)

type (
	// RunInput is persisted in agent_runs.input_messages.
	RunInput struct {
		ConnectorID    gid.GID `json:"connector_id"`
		OrganizationID gid.GID `json:"organization_id"`
		RunKind        string  `json:"run_kind"`
	}

	// Fact is a deterministic scanner observation passed to synthesis.
	Fact struct {
		Check  Check  `json:"check"`
		Scope  string `json:"scope"`
		Value  any    `json:"value"`
		APIRef string `json:"api_ref,omitempty"`
		Repo   string `json:"repo,omitempty"`
	}

	// FactSheet is the full scanner output for one discovery run.
	FactSheet struct {
		GitHubOrg    string   `json:"github_org"`
		Limitations  []string `json:"limitations,omitempty"`
		Facts        []Fact   `json:"facts"`
		ReposScanned int      `json:"repos_scanned"`
	}

	// ExistingMeasure is a slim view for synthesis dedup.
	ExistingMeasure struct {
		ID          gid.GID               `json:"id"`
		Name        string                `json:"name"`
		Description *string               `json:"description,omitempty"`
		Category    string                `json:"category"`
		State       coredata.MeasureState `json:"state"`
	}

	// MeasurePlanUpdate updates an existing measure linked to GitHub.
	MeasurePlanUpdate struct {
		MeasureID       gid.GID               `json:"measure_id"`
		State           coredata.MeasureState `json:"state"`
		EvidenceSummary string                `json:"evidence_summary"`
		CheckRefs       []Check               `json:"check_refs"`
	}

	// MeasurePlanCreate creates a new measure for GitHub posture.
	MeasurePlanCreate struct {
		Name            string                `json:"name"`
		Description     string                `json:"description"`
		Category        string                `json:"category"`
		State           coredata.MeasureState `json:"state"`
		EvidenceSummary string                `json:"evidence_summary"`
		CheckRefs       []Check               `json:"check_refs"`
	}

	// MeasurePlanUnchanged records measures left as-is.
	MeasurePlanUnchanged struct {
		MeasureID gid.GID `json:"measure_id"`
		Reason    string  `json:"reason"`
	}

	// MeasurePlan is the structured synthesis output.
	MeasurePlan struct {
		Updates   []MeasurePlanUpdate    `json:"updates"`
		Creates   []MeasurePlanCreate    `json:"creates"`
		Unchanged []MeasurePlanUnchanged `json:"unchanged"`
	}

	// RunResult is persisted in agent_runs.result when discovery completes.
	RunResult struct {
		Integration      string         `json:"integration"`
		ThirdPartyID     gid.GID        `json:"third_party_id"`
		GitHubOrg        string         `json:"github_org"`
		CompletedAt      string         `json:"completed_at"`
		Limitations      []string       `json:"limitations,omitempty"`
		ReposScanned     int            `json:"repos_scanned"`
		MeasuresUpserted int            `json:"measures_upserted"`
		Summary          map[string]int `json:"summary"`
	}
)
