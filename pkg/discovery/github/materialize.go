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
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	materializeRule struct {
		factKey     string
		name        string
		description string
		category    string
		evaluate    func(Fact) coredata.MeasureState
	}

	gidKey string
)

// MaterializeFromFacts builds a measure plan without an LLM (tests and fallback).
func MaterializeFromFacts(sheet *FactSheet, existing []ExistingMeasure) (*MeasurePlan, error) {
	rules := defaultMaterializeRules()
	byKey := map[string]Fact{}

	for _, fact := range sheet.Facts {
		byKey[fact.FactKey] = fact
	}

	plan := &MeasurePlan{
		Unchanged: []MeasurePlanUnchanged{},
	}

	used := map[gidKey]struct{}{}

	for _, rule := range rules {
		fact, ok := byKey[rule.factKey]
		if !ok {
			continue
		}

		state := rule.evaluate(fact)
		summary := fmt.Sprintf("%s (fact %s)", rule.description, fact.FactID)

		if match := findMeasureByName(existing, rule.name); match != nil {
			plan.Updates = append(plan.Updates, MeasurePlanUpdate{
				MeasureID:       match.ID,
				State:           state,
				EvidenceSummary: summary,
				FactRefs:        []string{fact.FactID},
			})
			used[gidKey(match.ID.String())] = struct{}{}

			continue
		}

		if len(plan.Creates) >= maxMeasureCreatesPerRun {
			continue
		}

		plan.Creates = append(plan.Creates, MeasurePlanCreate{
			Name:            rule.name,
			Description:     rule.description,
			Category:        rule.category,
			State:           state,
			EvidenceSummary: summary,
			FactRefs:        []string{fact.FactID},
		})
	}

	for _, m := range existing {
		if _, ok := used[gidKey(m.ID.String())]; ok {
			continue
		}

		plan.Unchanged = append(plan.Unchanged, MeasurePlanUnchanged{
			MeasureID: m.ID,
			Reason:    "No matching facts in this run.",
		})
	}

	return plan, nil
}

func findMeasureByName(existing []ExistingMeasure, name string) *ExistingMeasure {
	target := strings.ToLower(strings.TrimSpace(name))

	for i := range existing {
		if strings.ToLower(strings.TrimSpace(existing[i].Name)) == target {
			return &existing[i]
		}
	}

	return nil
}

func defaultMaterializeRules() []materializeRule {
	rules := []materializeRule{
		{
			factKey:     "org_mfa_required",
			name:        "Org-wide MFA enforcement",
			description: "Organization requires two-factor authentication for all members.",
			category:    "access",
			evaluate: func(f Fact) coredata.MeasureState {
				if b, ok := f.Value.(bool); ok && b {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			factKey:     "org_no_2fa_members",
			name:        "Members without 2FA",
			description: "No organization members lack two-factor authentication.",
			category:    "access",
			evaluate: func(f Fact) coredata.MeasureState {
				count, ok := factIntValue(f.Value)
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if count == 0 {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			factKey:     "org_base_permissions",
			name:        "Minimal default repository permissions",
			description: "Default repository permission is read or none.",
			category:    "access",
			evaluate: func(f Fact) coredata.MeasureState {
				perm, ok := f.Value.(string)
				if !ok {
					return coredata.MeasureStateUnknown
				}

				switch strings.ToLower(perm) {
				case "none", "read":
					return coredata.MeasureStateImplemented
				default:
					return coredata.MeasureStateNotImplemented
				}
			},
		},
		{
			factKey:     "org_no_public_repo_creation",
			name:        "Restrict public repository creation",
			description: "Members cannot create public repositories.",
			category:    "exposure",
			evaluate: func(f Fact) coredata.MeasureState {
				if b, ok := f.Value.(bool); ok && b {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			factKey:     "org_admin_minimization",
			name:        "Admin account minimization",
			description: "Organization admin accounts are limited.",
			category:    "access",
			evaluate: func(f Fact) coredata.MeasureState {
				return evaluateAdminMinimization(f.Value)
			},
		},
		{
			factKey:     "org_public_repos",
			name:        "Public repository exposure",
			description: "Unexpected public repositories are controlled.",
			category:    "exposure",
			evaluate: func(f Fact) coredata.MeasureState {
				public, total, ok := factCountPair(f.Value)
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if total == 0 || public == 0 {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			factKey:     "org_no_visibility_change",
			name:        "Restrict repository visibility changes",
			description: "Members cannot change repository visibility.",
			category:    "exposure",
			evaluate: func(f Fact) coredata.MeasureState {
				if b, ok := f.Value.(bool); ok && b {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			factKey:     "org_outside_collaborators",
			name:        "Outside collaborator inventory",
			description: "Outside collaborators are inventoried for review.",
			category:    "access",
			evaluate: func(f Fact) coredata.MeasureState {
				_, ok := factCountValue(f.Value, "count")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				return coredata.MeasureStateImplemented
			},
		},
		{
			factKey:     "org_actions_restricted",
			name:        "GitHub Actions usage restricted",
			description: "Organization restricts which actions and workflows may run.",
			category:    "ci_cd",
			evaluate: func(f Fact) coredata.MeasureState {
				restricted, ok := factBoolField(f.Value, "restricted")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if restricted {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			factKey:     "org_github_apps",
			name:        "GitHub App inventory",
			description: "Installed GitHub Apps are inventoried for review.",
			category:    "integrations",
			evaluate: func(f Fact) coredata.MeasureState {
				_, ok := factCountValue(f.Value, "installations")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				return coredata.MeasureStateImplemented
			},
		},
		{
			factKey:     "org_audit_log_accessible",
			name:        "Audit log accessible",
			description: "Organization audit log is accessible for review.",
			category:    "audit",
			evaluate: func(f Fact) coredata.MeasureState {
				if b, ok := f.Value.(bool); ok && b {
					return coredata.MeasureStateImplemented
				}

				if b, ok := f.Value.(bool); ok && !b {
					return coredata.MeasureStateNotApplicable
				}

				return coredata.MeasureStateUnknown
			},
		},
		{
			factKey:     "repo_branch_protection_coverage",
			name:        "Default branch protection",
			description: "Default branches are protected across scanned repositories.",
			category:    "code_review",
			evaluate:    evaluateFullCoverage,
		},
		{
			factKey:     "repo_pr_reviews_required_coverage",
			name:        "Pull request reviews required",
			description: "Default branches require pull request reviews.",
			category:    "code_review",
			evaluate:    evaluateFullCoverage,
		},
		{
			factKey:     "repo_signed_commits_required_coverage",
			name:        "Signed commits required",
			description: "Default branches require signed commits.",
			category:    "code_integrity",
			evaluate:    evaluateFullCoverage,
		},
		{
			factKey:     "repo_workflow_coverage",
			name:        "CI/CD workflows present",
			description: "Repositories run automated workflows.",
			category:    "ci_cd",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_security_md_coverage",
			name:        "Security disclosure policy",
			description: "Repositories publish a SECURITY.md disclosure policy.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_contributing_md_coverage",
			name:        "Contributing guidelines documented",
			description: "Repositories publish CONTRIBUTING.md guidance.",
			category:    "documentation",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_dependabot_config_coverage",
			name:        "Dependabot configuration",
			description: "Repositories configure Dependabot update automation.",
			category:    "dependencies",
			evaluate:    evaluateAnyCoverage,
		},
		{
			factKey:     "repo_dependabot_critical_open",
			name:        "Critical Dependabot alerts resolved",
			description: "No open critical Dependabot alerts in scanned repositories.",
			category:    "dependencies",
			evaluate: func(f Fact) coredata.MeasureState {
				count, ok := factCountValue(f.Value, "open_critical")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if count == 0 {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			factKey:     "repo_secret_scanning_alerts_open",
			name:        "Secret scanning alerts resolved",
			description: "No open secret scanning alerts in scanned repositories.",
			category:    "secrets",
			evaluate: func(f Fact) coredata.MeasureState {
				count, ok := factCountValue(f.Value, "open")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if count == 0 {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
		{
			factKey:     "repo_code_scanning_critical_open",
			name:        "Critical code scanning alerts resolved",
			description: "No open critical code scanning alerts in scanned repositories.",
			category:    "code_scanning",
			evaluate: func(f Fact) coredata.MeasureState {
				count, ok := factCountValue(f.Value, "open_critical")
				if !ok {
					return coredata.MeasureStateUnknown
				}

				if count == 0 {
					return coredata.MeasureStateImplemented
				}

				return coredata.MeasureStateNotImplemented
			},
		},
	}

	return append(rules, p0MaterializeRules()...)
}

func factIntValue(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func factCountPair(value any) (int, int, bool) {
	switch m := value.(type) {
	case map[string]int:
		return m["public"], m["total"], true
	case map[string]any:
		public, ok1 := toInt(m["public"])
		total, ok2 := toInt(m["total"])

		return public, total, ok1 && ok2
	default:
		return 0, 0, false
	}
}

func factCoveragePair(value any) (int, int, bool) {
	switch m := value.(type) {
	case map[string]int:
		return m["matched"], m["total"], m["total"] > 0
	case map[string]any:
		matched, ok1 := toInt(m["matched"])
		total, ok2 := toInt(m["total"])

		return matched, total, ok1 && ok2 && total > 0
	default:
		return 0, 0, false
	}
}

func factCountValue(value any, key string) (int, bool) {
	switch m := value.(type) {
	case map[string]int:
		v, ok := m[key]

		return v, ok
	case map[string]any:
		return toInt(m[key])
	default:
		return 0, false
	}
}

func factBoolField(value any, key string) (bool, bool) {
	m, ok := value.(map[string]any)
	if !ok {
		return false, false
	}

	v, ok := m[key].(bool)

	return v, ok
}

func evaluateFullCoverage(f Fact) coredata.MeasureState {
	matched, total, ok := factCoveragePair(f.Value)
	if !ok {
		return coredata.MeasureStateUnknown
	}

	if matched == total {
		return coredata.MeasureStateImplemented
	}

	if matched == 0 {
		return coredata.MeasureStateNotImplemented
	}

	return coredata.MeasureStateNotImplemented
}

func evaluateAnyCoverage(f Fact) coredata.MeasureState {
	matched, _, ok := factCoveragePair(f.Value)
	if !ok {
		return coredata.MeasureStateUnknown
	}

	if matched > 0 {
		return coredata.MeasureStateImplemented
	}

	return coredata.MeasureStateNotImplemented
}
