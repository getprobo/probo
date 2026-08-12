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

import {
  Badge,
  Card,
  IconChevronDown,
  IconChevronRight,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { type ReactNode, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { DiagramCardFragment$key } from "#/__generated__/core/DiagramCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { BoundaryActions } from "./BoundaryActions";
import { CreateBoundaryDialog } from "./CreateBoundaryDialog";
import { CreateNodeDialog } from "./CreateNodeDialog";
import { CreateProcessDialog } from "./CreateProcessDialog";
import { CreateScenarioInDiagramDialog } from "./CreateScenarioInDiagramDialog";
import { CreateThreatDialog } from "./CreateThreatDialog";
import { DiagramActions } from "./DiagramActions";
import { DiagramChart } from "./DiagramChart";
import { NodeActions } from "./NodeActions";
import { ProcessActions } from "./ProcessActions";
import { ScenarioInDiagramActions } from "./ScenarioInDiagramActions";
import { ThreatActions } from "./ThreatActions";

export const diagramCardFragment = graphql`
  fragment DiagramCardFragment on RiskAnalysisDiagram {
    id
    name
    nodes(first: 100)
      @connection(key: "RiskAnalysisDiagram_nodes", filters: []) {
      __id
      edges {
        node { id nodeType name boundaryId }
      }
    }
    boundaries(first: 100)
      @connection(key: "RiskAnalysisDiagram_boundaries", filters: []) {
      __id
      edges {
        node { id name parentBoundaryId }
      }
    }
    processes(first: 100)
      @connection(key: "RiskAnalysisDiagram_processes", filters: []) {
      __id
      edges {
        node { id sourceNodeId targetNodeId name }
      }
    }
    threats(first: 100)
      @connection(key: "RiskAnalysisDiagram_threats", filters: []) {
      __id
      edges {
        node { id processId name category }
      }
    }
    scenarios(first: 100)
      @connection(key: "RiskAnalysisDiagram_scenarios", filters: []) {
      __id
      edges {
        node {
          id name description
          risks(first: 10) {
            edges { node { id name } }
          }
          threats(first: 10) {
            edges { node { id name } }
          }
        }
      }
    }
    ...DiagramChart_diagram
  }
`;

function SectionHeader(props: { title: string; hint?: string; children: ReactNode }) {
  return (
    <div className="mb-3">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">{props.title}</h3>
        {props.children}
      </div>
      {props.hint && (
        <p className="text-xs text-txt-tertiary mt-1">{props.hint}</p>
      )}
    </div>
  );
}

export function DiagramCard(props: {
  diagramRef: DiagramCardFragment$key;
  diagramsConnectionId: string;
}) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const [isOpen, setIsOpen] = useState(true);
  const diagram = useFragment(diagramCardFragment, props.diagramRef);
  const { diagramsConnectionId } = props;

  const nodes = diagram.nodes?.edges.map(e => e.node) ?? [];
  const boundaries = diagram.boundaries?.edges.map(e => e.node) ?? [];
  const processes = diagram.processes?.edges.map(e => e.node) ?? [];
  const threats = diagram.threats?.edges.map(e => e.node) ?? [];
  const scenarios = diagram.scenarios?.edges.map(e => e.node) ?? [];
  const nodeMap = new Map(nodes.map(n => [n.id, n]));
  const boundaryMap = new Map(boundaries.map(b => [b.id, b]));
  const boundaryOptions = boundaries.map(b => ({ id: b.id, name: b.name }));
  const nodesConnId = diagram.nodes?.__id ?? "";
  const boundariesConnId = diagram.boundaries?.__id ?? "";
  const processesConnId = diagram.processes?.__id ?? "";
  const threatsConnId = diagram.threats?.__id ?? "";
  const scenariosConnId = diagram.scenarios?.__id ?? "";

  const ChevronIcon = isOpen ? IconChevronDown : IconChevronRight;

  return (
    <Card>
      <button
        type="button"
        className="flex w-full items-center justify-between px-4 py-3"
        onClick={() => setIsOpen(v => !v)}
      >
        <div className="text-left">
          <h3 className="text-sm font-semibold">{diagram.name}</h3>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-txt-tertiary">
            {t("diagramCard.summary", {
              nodes: nodes.length,
              processes: processes.length,
              threats: threats.length,
              scenarios: scenarios.length,
            })}
          </span>
          <div
            onClick={e => e.stopPropagation()}
            onKeyDown={e => e.stopPropagation()}
          >
            <DiagramActions
              diagram={{ id: diagram.id, name: diagram.name }}
              connectionId={diagramsConnectionId}
            />
          </div>
          <ChevronIcon size={16} className="text-txt-tertiary" />
        </div>
      </button>

      {isOpen && (
        <div className="border-t border-border-low px-4 py-4 space-y-6">
          <DiagramChart diagramKey={diagram} />

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div>
              <SectionHeader
                title={t("diagramCard.sectionTitle.nodes", { count: nodes.length })}
                hint={t("diagramCard.hints.nodes")}
              >
                <CreateNodeDialog diagramId={diagram.id} connectionId={nodesConnId} boundaries={boundaryOptions} />
              </SectionHeader>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{t("diagramCard.columns.name")}</Th>
                    <Th>{t("diagramCard.columns.type")}</Th>
                    <Th>{t("diagramCard.columns.boundary")}</Th>
                    <Th className="w-12" />
                  </Tr>
                </Thead>
                <Tbody>
                  {nodes.map(node => (
                    <Tr key={node.id}>
                      <Td className="font-medium">{node.name}</Td>
                      <Td><Badge>{node.nodeType}</Badge></Td>
                      <Td className="text-txt-secondary">{node.boundaryId ? boundaryMap.get(node.boundaryId)?.name ?? "—" : "—"}</Td>
                      <Td>
                        <NodeActions
                          node={{
                            id: node.id,
                            name: node.name,
                            nodeType: node.nodeType,
                            boundaryId: node.boundaryId ?? null,
                          }}
                          boundaries={boundaryOptions}
                          connectionId={nodesConnId}
                        />
                      </Td>
                    </Tr>
                  ))}
                  {nodes.length === 0 && (
                    <Tr>
                      <Td colSpan={4} className="text-center text-txt-secondary">{t("diagramCard.empty.nodes")}</Td>
                    </Tr>
                  )}
                </Tbody>
              </Table>
            </div>

            <div>
              <SectionHeader
                title={t("diagramCard.sectionTitle.processes", { count: processes.length })}
                hint={t("diagramCard.hints.processes")}
              >
                <CreateProcessDialog
                  diagramId={diagram.id}
                  nodes={nodes.map(n => ({ id: n.id, name: n.name }))}
                  connectionId={processesConnId}
                />
              </SectionHeader>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{t("diagramCard.columns.name")}</Th>
                    <Th>{t("diagramCard.columns.from")}</Th>
                    <Th>{t("diagramCard.columns.to")}</Th>
                    <Th className="w-12" />
                  </Tr>
                </Thead>
                <Tbody>
                  {processes.map(process => (
                    <Tr key={process.id}>
                      <Td className="font-medium">{process.name}</Td>
                      <Td className="text-txt-secondary">{nodeMap.get(process.sourceNodeId)?.name ?? "—"}</Td>
                      <Td className="text-txt-secondary">{nodeMap.get(process.targetNodeId)?.name ?? "—"}</Td>
                      <Td>
                        <ProcessActions
                          process={{
                            id: process.id,
                            name: process.name,
                            sourceNodeId: process.sourceNodeId,
                            targetNodeId: process.targetNodeId,
                          }}
                          nodes={nodes.map(n => ({ id: n.id, name: n.name }))}
                          connectionId={processesConnId}
                        />
                      </Td>
                    </Tr>
                  ))}
                  {processes.length === 0 && (
                    <Tr>
                      <Td colSpan={4} className="text-center text-txt-secondary">{t("diagramCard.empty.processes")}</Td>
                    </Tr>
                  )}
                </Tbody>
              </Table>
            </div>
          </div>

          <div>
            <SectionHeader
              title={t("diagramCard.sectionTitle.boundaries", { count: boundaries.length })}
              hint={t("diagramCard.hints.boundaries")}
            >
              <CreateBoundaryDialog
                diagramId={diagram.id}
                connectionId={boundariesConnId}
                boundaries={boundaryOptions}
              />
            </SectionHeader>
            <Table>
              <Thead>
                <Tr>
                  <Th>{t("diagramCard.columns.name")}</Th>
                  <Th>{t("diagramCard.columns.parent")}</Th>
                  <Th className="w-12" />
                </Tr>
              </Thead>
              <Tbody>
                {boundaries.map(boundary => (
                  <Tr key={boundary.id}>
                    <Td className="font-medium">{boundary.name}</Td>
                    <Td className="text-txt-secondary">{boundary.parentBoundaryId ? boundaryMap.get(boundary.parentBoundaryId)?.name ?? "—" : "—"}</Td>
                    <Td>
                      <BoundaryActions
                        boundary={{
                          id: boundary.id,
                          name: boundary.name,
                          parentBoundaryId: boundary.parentBoundaryId ?? null,
                        }}
                        boundaries={boundaryOptions}
                        connectionId={boundariesConnId}
                      />
                    </Td>
                  </Tr>
                ))}
                {boundaries.length === 0 && (
                  <Tr>
                    <Td colSpan={3} className="text-center text-txt-secondary">{t("diagramCard.empty.boundaries")}</Td>
                  </Tr>
                )}
              </Tbody>
            </Table>
          </div>

          <div>
            <SectionHeader
              title={t("diagramCard.sectionTitle.threats", { count: threats.length })}
              hint={t("diagramCard.hints.threats")}
            >
              <CreateThreatDialog
                diagramId={diagram.id}
                processes={processes.map(p => ({ id: p.id, name: p.name }))}
                connectionId={threatsConnId}
              />
            </SectionHeader>
            <Table>
              <Thead>
                <Tr>
                  <Th>{t("diagramCard.columns.threat")}</Th>
                  <Th>{t("diagramCard.columns.category")}</Th>
                  <Th>{t("diagramCard.columns.process")}</Th>
                  <Th className="w-12" />
                </Tr>
              </Thead>
              <Tbody>
                {threats.map((threat) => {
                  const process = processes.find(p => p.id === threat.processId);
                  return (
                    <Tr key={threat.id}>
                      <Td className="font-medium">{threat.name}</Td>
                      <Td><Badge>{threat.category}</Badge></Td>
                      <Td className="text-txt-secondary">{process?.name ?? "—"}</Td>
                      <Td>
                        <ThreatActions
                          threat={{ id: threat.id, name: threat.name, category: threat.category }}
                          connectionId={threatsConnId}
                        />
                      </Td>
                    </Tr>
                  );
                })}
                {threats.length === 0 && (
                  <Tr>
                    <Td colSpan={4} className="text-center text-txt-secondary">{t("diagramCard.empty.threats")}</Td>
                  </Tr>
                )}
              </Tbody>
            </Table>
          </div>

          <div>
            <SectionHeader
              title={t("diagramCard.sectionTitle.scenarios", { count: scenarios.length })}
              hint={t("diagramCard.hints.scenarios")}
            >
              <CreateScenarioInDiagramDialog
                diagramId={diagram.id}
                threats={threats.map(t => ({ id: t.id, name: t.name }))}
                connectionId={scenariosConnId}
              />
            </SectionHeader>
            <Table>
              <Thead>
                <Tr>
                  <Th>{t("diagramCard.columns.scenario")}</Th>
                  <Th>{t("diagramCard.columns.risks")}</Th>
                  <Th>{t("diagramCard.columns.threats")}</Th>
                  <Th className="w-12" />
                </Tr>
              </Thead>
              <Tbody>
                {scenarios.map((scenario) => {
                  const scenarioRisks = scenario.risks?.edges.map(e => e.node) ?? [];
                  const scenarioThreats = scenario.threats?.edges.map(e => e.node) ?? [];
                  return (
                    <Tr key={scenario.id}>
                      <Td className="font-medium">{scenario.name}</Td>
                      <Td className="text-txt-secondary">
                        {scenarioRisks.length > 0
                          ? scenarioRisks.map((risk, i) => (
                              <span key={risk.id}>
                                {i > 0 && ", "}
                                <Link
                                  to={`/organizations/${organizationId}/risk-management/risks/${risk.id}`}
                                  className="text-txt-primary underline"
                                >
                                  {risk.name}
                                </Link>
                              </span>
                            ))
                          : "—"}
                      </Td>
                      <Td className="text-txt-secondary">
                        {scenarioThreats.length > 0
                          ? scenarioThreats.map(t => t.name).join(", ")
                          : "—"}
                      </Td>
                      <Td>
                        <ScenarioInDiagramActions
                          scenario={{
                            id: scenario.id,
                            name: scenario.name,
                            description: scenario.description ?? null,
                            risks: scenarioRisks,
                            threats: scenarioThreats,
                          }}
                          diagramThreats={threats.map(t => ({ id: t.id, name: t.name }))}
                          connectionId={scenariosConnId}
                        />
                      </Td>
                    </Tr>
                  );
                })}
                {scenarios.length === 0 && (
                  <Tr>
                    <Td colSpan={4} className="text-center text-txt-secondary">{t("diagramCard.empty.scenarios")}</Td>
                  </Tr>
                )}
              </Tbody>
            </Table>
          </div>
        </div>
      )}
    </Card>
  );
}
