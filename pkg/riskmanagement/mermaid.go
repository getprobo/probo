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

package riskmanagement

import (
	"context"
	"fmt"
	"strings"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func (s *Service) BuildScopeMermaidChart(ctx context.Context, scope coredata.Scoper, scopeID gid.GID) (string, error) {
	var (
		nodes     coredata.RiskAssessmentNodes
		processes coredata.RiskAssessmentProcesses
		threats   coredata.RiskAssessmentThreats
	)

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := nodes.LoadAllByRiskAssessmentScopeID(ctx, conn, scope, scopeID); err != nil {
			return fmt.Errorf("cannot load nodes: %w", err)
		}
		if err := processes.LoadAllByRiskAssessmentScopeID(ctx, conn, scope, scopeID); err != nil {
			return fmt.Errorf("cannot load processes: %w", err)
		}
		if err := threats.LoadAllByRiskAssessmentScopeID(ctx, conn, scope, scopeID); err != nil {
			return fmt.Errorf("cannot load threats: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return buildScopeMermaidChart(nodes, processes, threats), nil
}

func buildScopeMermaidChart(
	nodes coredata.RiskAssessmentNodes,
	processes coredata.RiskAssessmentProcesses,
	threats coredata.RiskAssessmentThreats,
) string {
	if len(nodes) == 0 {
		return ""
	}

	nodeAlias := make(map[gid.GID]string, len(nodes))
	for i, n := range nodes {
		nodeAlias[n.ID] = fmt.Sprintf("n%d", i)
	}

	var b strings.Builder
	b.WriteString("flowchart LR\n")

	for _, n := range nodes {
		id := nodeAlias[n.ID]
		fmt.Fprintf(&b, "  %s\n", mermaidNodeShape(n.NodeType, id, n.Name))
		fmt.Fprintf(&b, "  class %s %s\n", id, mermaidNodeClass(n.NodeType))
	}

	for _, p := range processes {
		src, srcOK := nodeAlias[p.SourceNodeID]
		dst, dstOK := nodeAlias[p.TargetNodeID]
		if !srcOK || !dstOK {
			continue
		}
		fmt.Fprintf(&b, "  %s -- \"%s\" --> %s\n", src, escapeMermaidLabel(p.Name), dst)
	}

	processTarget := make(map[gid.GID]gid.GID, len(processes))
	for _, p := range processes {
		processTarget[p.ID] = p.TargetNodeID
	}

	for i, t := range threats {
		target, ok := processTarget[t.ProcessID]
		if !ok {
			continue
		}
		targetAlias, ok := nodeAlias[target]
		if !ok {
			continue
		}
		tid := fmt.Sprintf("t%d", i)
		label := escapeMermaidLabel(fmt.Sprintf("%s (%s)", t.Name, t.Category))
		fmt.Fprintf(&b, "  %s{{\"%s\"}}\n", tid, label)
		fmt.Fprintf(&b, "  class %s nodeThreat\n", tid)
		fmt.Fprintf(&b, "  %s -.-> %s\n", tid, targetAlias)
	}

	b.WriteString("  classDef nodeEntity fill:#dbeafe,stroke:#1d4ed8,color:#1e3a8a\n")
	b.WriteString("  classDef nodeBoundary fill:#fef3c7,stroke:#b45309,color:#78350f\n")
	b.WriteString("  classDef nodeAsset fill:#e5e7eb,stroke:#374151,color:#111827\n")
	b.WriteString("  classDef nodeData fill:#dcfce7,stroke:#15803d,color:#14532d\n")
	b.WriteString("  classDef nodeThreat fill:#fee2e2,stroke:#b91c1c,color:#7f1d1d\n")

	return strings.TrimRight(b.String(), "\n")
}

func mermaidNodeShape(t coredata.RiskAssessmentNodeType, id, name string) string {
	label := `"` + escapeMermaidLabel(name) + `"`
	switch t {
	case coredata.RiskAssessmentNodeTypeEntity:
		return fmt.Sprintf("%s([%s])", id, label)
	case coredata.RiskAssessmentNodeTypeBoundary:
		return fmt.Sprintf("%s{{%s}}", id, label)
	case coredata.RiskAssessmentNodeTypeData:
		return fmt.Sprintf("%s[(%s)]", id, label)
	case coredata.RiskAssessmentNodeTypeAsset:
		fallthrough
	default:
		return fmt.Sprintf("%s[%s]", id, label)
	}
}

func mermaidNodeClass(t coredata.RiskAssessmentNodeType) string {
	switch t {
	case coredata.RiskAssessmentNodeTypeEntity:
		return "nodeEntity"
	case coredata.RiskAssessmentNodeTypeBoundary:
		return "nodeBoundary"
	case coredata.RiskAssessmentNodeTypeData:
		return "nodeData"
	case coredata.RiskAssessmentNodeTypeAsset:
		fallthrough
	default:
		return "nodeAsset"
	}
}

var mermaidLabelReplacer = strings.NewReplacer(
	"&", "&amp;",
	`"`, "#quot;",
	"<", "&lt;",
	">", "&gt;",
	"\r\n", " ",
	"\n", " ",
)

func escapeMermaidLabel(s string) string {
	return mermaidLabelReplacer.Replace(s)
}
