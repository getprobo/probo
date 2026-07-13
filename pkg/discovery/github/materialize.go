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
	return []materializeRule{
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
	}
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
