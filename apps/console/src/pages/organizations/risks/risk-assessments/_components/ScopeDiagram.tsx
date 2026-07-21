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

import { CheckIcon, CopyIcon } from "@phosphor-icons/react";
import { Button, MermaidDiagram, useToast } from "@probo/ui";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchQuery, graphql, useFragment, useRelayEnvironment } from "react-relay";

import type { ScopeDiagram_scope$key } from "#/__generated__/core/ScopeDiagram_scope.graphql";
import type { ScopeDiagramMermaidQuery } from "#/__generated__/core/ScopeDiagramMermaidQuery.graphql";

const scopeDiagramFragment = graphql`
  fragment ScopeDiagram_scope on RiskAssessmentScope {
    id
    mermaidChart
    nodes(first: 100)
      @connection(key: "RiskAssessmentScope_nodes", filters: []) {
      edges {
        node {
          id
          name
          nodeType
          boundaryId
        }
      }
    }
    boundaries(first: 100)
      @connection(key: "RiskAssessmentScope_boundaries", filters: []) {
      edges {
        node {
          id
          name
          parentBoundaryId
        }
      }
    }
    processes(first: 100)
      @connection(key: "RiskAssessmentScope_processes", filters: []) {
      edges {
        node {
          id
          name
          sourceNodeId
          targetNodeId
        }
      }
    }
    threats(first: 100)
      @connection(key: "RiskAssessmentScope_threats", filters: []) {
      edges {
        node {
          id
          name
          processId
          category
        }
      }
    }
  }
`;

const scopeDiagramMermaidQuery = graphql`
  query ScopeDiagramMermaidQuery($scopeId: ID!) {
    node(id: $scopeId) {
      ... on RiskAssessmentScope {
        id
        mermaidChart
      }
    }
  }
`;

interface ScopeDiagramProps {
  scopeKey: ScopeDiagram_scope$key;
}

export function ScopeDiagram({ scopeKey }: ScopeDiagramProps) {
  const { t } = useTranslation();
  const environment = useRelayEnvironment();
  const scope = useFragment(scopeDiagramFragment, scopeKey);
  const mermaidChart = scope.mermaidChart;

  const nodeSignature = scope.nodes?.edges
    .map(e => `${e.node.id}|${e.node.name}|${e.node.nodeType}|${e.node.boundaryId ?? ""}`)
    .join(";") ?? "";
  const boundarySignature = scope.boundaries?.edges
    .map(e => `${e.node.id}|${e.node.name}|${e.node.parentBoundaryId ?? ""}`)
    .join(";") ?? "";
  const processSignature = scope.processes?.edges
    .map(e => `${e.node.id}|${e.node.name}|${e.node.sourceNodeId}|${e.node.targetNodeId}`)
    .join(";") ?? "";
  const threatSignature = scope.threats?.edges
    .map(e => `${e.node.id}|${e.node.name}|${e.node.processId}|${e.node.category}`)
    .join(";") ?? "";
  const signature = `${nodeSignature}::${boundarySignature}::${processSignature}::${threatSignature}`;
  const previousSignature = useRef(signature);
  useEffect(() => {
    if (previousSignature.current === signature) {
      return;
    }
    previousSignature.current = signature;
    const subscription = fetchQuery<ScopeDiagramMermaidQuery>(
      environment,
      scopeDiagramMermaidQuery,
      { scopeId: scope.id },
      { fetchPolicy: "network-only" },
    ).subscribe({});
    return () => subscription.unsubscribe();
  }, [signature, environment, scope.id]);

  if (!mermaidChart) {
    return (
      <div className="text-center text-txt-secondary text-sm py-6">
        {t("scopeDiagram.empty")}
      </div>
    );
  }

  return (
    <div className="relative">
      <div className="absolute right-0 top-0 z-10">
        <CopyChartButton chart={mermaidChart} />
      </div>
      <div className="overflow-x-auto">
        <MermaidDiagram chart={mermaidChart} />
        <Legend />
      </div>
    </div>
  );
}

interface CopyChartButtonProps {
  chart: string;
}

function CopyChartButton({ chart }: CopyChartButtonProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [copied, setCopied] = useState(false);

  const onClick = () => {
    const onFailure = () => {
      toast({
        title: t("scopeDiagram.errorTitle"),
        description: t("scopeDiagram.copyError"),
        variant: "error",
      });
    };

    if (!navigator.clipboard?.writeText) {
      onFailure();
      return;
    }

    try {
      navigator.clipboard.writeText(chart).then(
        () => {
          setCopied(true);
          setTimeout(() => setCopied(false), 1500);
        },
        onFailure,
      );
    } catch {
      onFailure();
    }
  };

  return (
    <Button
      variant="secondary"
      icon={copied ? CheckIcon : CopyIcon}
      onClick={onClick}
      aria-label={t("scopeDiagram.copyMermaidSource")}
      title={t("scopeDiagram.copyMermaidSource")}
    />
  );
}

type LegendShape = "stadium" | "hexagon" | "rectangle" | "cylinder";

type LegendItem = {
  label: string;
  shape: LegendShape;
  fill: string;
  stroke: string;
  text: string;
};

function Legend() {
  const { t } = useTranslation();
  const items: LegendItem[] = [
    { label: t("scopeDiagram.legend.entity"), shape: "stadium", fill: "#dbeafe", stroke: "#1d4ed8", text: "#1e3a8a" },
    { label: t("scopeDiagram.legend.boundary"), shape: "rectangle", fill: "#ffffff", stroke: "#b45309", text: "#78350f" },
    { label: t("scopeDiagram.legend.asset"), shape: "rectangle", fill: "#e5e7eb", stroke: "#374151", text: "#111827" },
    { label: t("scopeDiagram.legend.data"), shape: "cylinder", fill: "#dcfce7", stroke: "#15803d", text: "#14532d" },
    { label: t("scopeDiagram.legend.threat"), shape: "hexagon", fill: "#fee2e2", stroke: "#b91c1c", text: "#7f1d1d" },
  ];
  return (
    <div className="flex flex-wrap items-center gap-3 mt-3">
      {items.map(item => (
        <LegendSwatch key={item.label} item={item} />
      ))}
    </div>
  );
}

interface LegendSwatchProps {
  item: LegendItem;
}

function LegendSwatch({ item }: LegendSwatchProps) {
  const w = 88;
  const h = 28;
  return (
    <svg
      width={w}
      height={h}
      viewBox={`0 0 ${w} ${h}`}
      aria-label={item.label}
    >
      <LegendShapeEl shape={item.shape} w={w} h={h} fill={item.fill} stroke={item.stroke} />
      <text
        x={w / 2}
        y={h / 2}
        textAnchor="middle"
        dominantBaseline="central"
        fontSize={11}
        fontFamily="inherit"
        fill={item.text}
      >
        {item.label}
      </text>
    </svg>
  );
}

interface LegendShapeElProps {
  shape: LegendShape;
  w: number;
  h: number;
  fill: string;
  stroke: string;
}

function LegendShapeEl({ shape, w, h, fill, stroke }: LegendShapeElProps) {
  const sw = 1.5;
  const common = { fill, stroke, strokeWidth: sw };

  switch (shape) {
    case "stadium":
      return <rect x={sw / 2} y={sw / 2} width={w - sw} height={h - sw} rx={(h - sw) / 2} {...common} />;
    case "rectangle":
      return <rect x={sw / 2} y={sw / 2} width={w - sw} height={h - sw} {...common} />;
    case "hexagon": {
      const inset = 8;
      const pts = [
        `${inset},${h / 2}`,
        `${inset + 4},${sw}`,
        `${w - inset - 4},${sw}`,
        `${w - inset},${h / 2}`,
        `${w - inset - 4},${h - sw}`,
        `${inset + 4},${h - sw}`,
      ].join(" ");
      return <polygon points={pts} {...common} />;
    }
    case "cylinder": {
      const ry = 4;
      return (
        <g>
          <path
            d={`M ${sw / 2} ${ry + sw / 2} A ${(w - sw) / 2} ${ry} 0 0 1 ${w - sw / 2} ${ry + sw / 2} L ${w - sw / 2} ${h - ry - sw / 2} A ${(w - sw) / 2} ${ry} 0 0 1 ${sw / 2} ${h - ry - sw / 2} Z`}
            {...common}
          />
          <path
            d={`M ${sw / 2} ${ry + sw / 2} A ${(w - sw) / 2} ${ry} 0 0 0 ${w - sw / 2} ${ry + sw / 2}`}
            fill="none"
            stroke={stroke}
            strokeWidth={sw}
          />
        </g>
      );
    }
  }
}
