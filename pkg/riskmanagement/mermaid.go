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

package riskmanagement

import (
	"context"
	"fmt"
	"strings"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func (s *Service) BuildDiagramMermaidChart(ctx context.Context, scope coredata.Scoper, diagramID gid.GID) (string, error) {
	var (
		nodes      coredata.RiskAnalysisNodes
		boundaries coredata.RiskAnalysisBoundaries
		processes  coredata.RiskAnalysisProcesses
		threats    coredata.RiskAnalysisThreats
	)

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loadedNodes, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.RiskAnalysisNodeOrderField]{
				Field:     coredata.RiskAnalysisNodeOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisNodeOrderField]) ([]*coredata.RiskAnalysisNode, error) {
				var batch coredata.RiskAnalysisNodes
				if err := batch.LoadByRiskAnalysisDiagramID(ctx, conn, scope, diagramID, cursor); err != nil {
					return nil, fmt.Errorf("cannot load nodes: %w", err)
				}

				return batch, nil
			},
		)
		if err != nil {
			return err
		}

		nodes = loadedNodes

		loadedBoundaries, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.RiskAnalysisBoundaryOrderField]{
				Field:     coredata.RiskAnalysisBoundaryOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisBoundaryOrderField]) ([]*coredata.RiskAnalysisBoundary, error) {
				var batch coredata.RiskAnalysisBoundaries
				if err := batch.LoadByRiskAnalysisDiagramID(ctx, conn, scope, diagramID, cursor); err != nil {
					return nil, fmt.Errorf("cannot load boundaries: %w", err)
				}

				return batch, nil
			},
		)
		if err != nil {
			return err
		}

		boundaries = loadedBoundaries

		loadedProcesses, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.RiskAnalysisProcessOrderField]{
				Field:     coredata.RiskAnalysisProcessOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisProcessOrderField]) ([]*coredata.RiskAnalysisProcess, error) {
				var batch coredata.RiskAnalysisProcesses
				if err := batch.LoadByRiskAnalysisDiagramID(ctx, conn, scope, diagramID, cursor); err != nil {
					return nil, fmt.Errorf("cannot load processes: %w", err)
				}

				return batch, nil
			},
		)
		if err != nil {
			return err
		}

		processes = loadedProcesses

		loadedThreats, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.RiskAnalysisThreatOrderField]{
				Field:     coredata.RiskAnalysisThreatOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisThreatOrderField]) ([]*coredata.RiskAnalysisThreat, error) {
				var batch coredata.RiskAnalysisThreats
				if err := batch.LoadByRiskAnalysisDiagramID(ctx, conn, scope, diagramID, cursor); err != nil {
					return nil, fmt.Errorf("cannot load threats: %w", err)
				}

				return batch, nil
			},
		)
		if err != nil {
			return err
		}

		threats = loadedThreats

		return nil
	})
	if err != nil {
		return "", err
	}

	return buildDiagramMermaidChart(nodes, boundaries, processes, threats), nil
}

func buildDiagramMermaidChart(
	nodes coredata.RiskAnalysisNodes,
	boundaries coredata.RiskAnalysisBoundaries,
	processes coredata.RiskAnalysisProcesses,
	threats coredata.RiskAnalysisThreats,
) string {
	if len(nodes) == 0 && len(boundaries) == 0 {
		return ""
	}

	nodeAlias := make(map[gid.GID]string, len(nodes))
	for i, n := range nodes {
		nodeAlias[n.ID] = fmt.Sprintf("n%d", i)
	}

	boundaryAlias := make(map[gid.GID]string, len(boundaries))
	for i, bnd := range boundaries {
		boundaryAlias[bnd.ID] = fmt.Sprintf("b%d", i)
	}

	// Group boundaries by their parent so nested boundaries become nested subgraphs.
	childBoundaries := make(map[gid.GID]coredata.RiskAnalysisBoundaries)

	var rootBoundaries coredata.RiskAnalysisBoundaries

	for _, bnd := range boundaries {
		if bnd.ParentBoundaryID != nil {
			if _, ok := boundaryAlias[*bnd.ParentBoundaryID]; ok {
				childBoundaries[*bnd.ParentBoundaryID] = append(childBoundaries[*bnd.ParentBoundaryID], bnd)
				continue
			}
		}

		rootBoundaries = append(rootBoundaries, bnd)
	}

	// Group nodes by the boundary that contains them; nodes without a
	// boundary (or referencing an unknown one) are rendered at the top level.
	nodesByBoundary := make(map[gid.GID]coredata.RiskAnalysisNodes)

	var rootNodes coredata.RiskAnalysisNodes

	for _, n := range nodes {
		if n.BoundaryID != nil {
			if _, ok := boundaryAlias[*n.BoundaryID]; ok {
				nodesByBoundary[*n.BoundaryID] = append(nodesByBoundary[*n.BoundaryID], n)
				continue
			}
		}

		rootNodes = append(rootNodes, n)
	}

	var b strings.Builder
	b.WriteString("flowchart LR\n")

	// class statements must live at the flowchart level, not inside a
	// subgraph block, so collect them and emit once all shapes are written.
	var classLines []string

	emitNode := func(n *coredata.RiskAnalysisNode, indent string) {
		id := nodeAlias[n.ID]
		fmt.Fprintf(&b, "%s%s\n", indent, mermaidNodeShape(n.NodeType, id, n.Name))
		classLines = append(classLines, fmt.Sprintf("  class %s %s", id, mermaidNodeClass(n.NodeType)))
	}

	var emitBoundary func(bnd *coredata.RiskAnalysisBoundary, indent string)

	emitBoundary = func(bnd *coredata.RiskAnalysisBoundary, indent string) {
		alias := boundaryAlias[bnd.ID]
		fmt.Fprintf(&b, "%ssubgraph %s[\"%s\"]\n", indent, alias, escapeMermaidLabel(bnd.Name))

		inner := indent + "  "
		for _, child := range childBoundaries[bnd.ID] {
			emitBoundary(child, inner)
		}

		for _, n := range nodesByBoundary[bnd.ID] {
			emitNode(n, inner)
		}

		fmt.Fprintf(&b, "%send\n", indent)

		classLines = append(classLines, fmt.Sprintf("  class %s nodeBoundary", alias))
	}

	for _, bnd := range rootBoundaries {
		emitBoundary(bnd, "  ")
	}

	for _, n := range rootNodes {
		emitNode(n, "  ")
	}

	for _, line := range classLines {
		b.WriteString(line + "\n")
	}

	for _, p := range processes {
		src, srcOK := nodeAlias[p.SourceNodeID]

		dst, dstOK := nodeAlias[p.TargetNodeID]
		if !srcOK || !dstOK {
			continue
		}

		fmt.Fprintf(&b, "  %s -- \"%s\" --> %s\n", src, mermaidEdgeLabel(p.Name), dst)
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
	b.WriteString("  classDef nodeBoundary fill:#ffffff,stroke:#b45309,color:#78350f\n")
	b.WriteString("  classDef nodeAsset fill:#e5e7eb,stroke:#374151,color:#111827\n")
	b.WriteString("  classDef nodeData fill:#dcfce7,stroke:#15803d,color:#14532d\n")
	b.WriteString("  classDef nodeThreat fill:#fee2e2,stroke:#b91c1c,color:#7f1d1d\n")

	return strings.TrimRight(b.String(), "\n")
}

func mermaidNodeShape(t coredata.RiskAnalysisNodeType, id, name string) string {
	label := `"` + escapeMermaidLabel(name) + `"`

	switch t {
	case coredata.RiskAnalysisNodeTypeEntity:
		return fmt.Sprintf("%s([%s])", id, label)
	case coredata.RiskAnalysisNodeTypeData:
		return fmt.Sprintf("%s[(%s)]", id, label)
	case coredata.RiskAnalysisNodeTypeAsset:
		fallthrough
	default:
		return fmt.Sprintf("%s[%s]", id, label)
	}
}

func mermaidNodeClass(t coredata.RiskAnalysisNodeType) string {
	switch t {
	case coredata.RiskAnalysisNodeTypeEntity:
		return "nodeEntity"
	case coredata.RiskAnalysisNodeTypeData:
		return "nodeData"
	case coredata.RiskAnalysisNodeTypeAsset:
		fallthrough
	default:
		return "nodeAsset"
	}
}

const mermaidEdgeLabelWrapWidth = 28

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

// mermaidEdgeLabel escapes a process name and inserts <br> breaks so long
// edge labels wrap. Mermaid's wrappingWidth only applies to nodes, not edges.
func mermaidEdgeLabel(s string) string {
	return strings.ReplaceAll(wrapWords(escapeMermaidLabel(s), mermaidEdgeLabelWrapWidth), "\n", "<br>")
}

func wrapWords(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}

	words := strings.Fields(s)
	if len(words) <= 1 {
		return s
	}

	var (
		b       strings.Builder
		lineLen int
	)

	for _, word := range words {
		if lineLen > 0 && lineLen+1+len(word) > width {
			b.WriteByte('\n')

			lineLen = 0
		}

		if lineLen > 0 {
			b.WriteByte(' ')

			lineLen++
		}

		b.WriteString(word)

		lineLen += len(word)
	}

	return b.String()
}
