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

import { useTranslate } from "@probo/i18n";
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
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { ScopeCardFragment$key } from "#/__generated__/core/ScopeCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { BoundaryActions } from "./BoundaryActions";
import { CreateBoundaryDialog } from "./CreateBoundaryDialog";
import { CreateNodeDialog } from "./CreateNodeDialog";
import { CreateProcessDialog } from "./CreateProcessDialog";
import { CreateScenarioInScopeDialog } from "./CreateScenarioInScopeDialog";
import { CreateThreatDialog } from "./CreateThreatDialog";
import { NodeActions } from "./NodeActions";
import { ProcessActions } from "./ProcessActions";
import { ScenarioInScopeActions } from "./ScenarioInScopeActions";
import { ScopeActions } from "./ScopeActions";
import { ScopeDiagram } from "./ScopeDiagram";
import { ThreatActions } from "./ThreatActions";

export const scopeCardFragment = graphql`
  fragment ScopeCardFragment on RiskAssessmentScope {
    id
    name
    nodes(first: 100)
      @connection(key: "RiskAssessmentScope_nodes", filters: []) {
      __id
      edges {
        node { id nodeType name boundaryId }
      }
    }
    boundaries(first: 100)
      @connection(key: "RiskAssessmentScope_boundaries", filters: []) {
      __id
      edges {
        node { id name parentBoundaryId }
      }
    }
    processes(first: 100)
      @connection(key: "RiskAssessmentScope_processes", filters: []) {
      __id
      edges {
        node { id sourceNodeId targetNodeId name }
      }
    }
    threats(first: 100)
      @connection(key: "RiskAssessmentScope_threats", filters: []) {
      __id
      edges {
        node { id processId name category }
      }
    }
    scenarios(first: 100)
      @connection(key: "RiskAssessmentScope_scenarios", filters: []) {
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
    ...ScopeDiagram_scope
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

export function ScopeCard(props: {
  scopeRef: ScopeCardFragment$key;
  scopesConnectionId: string;
}) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const [isOpen, setIsOpen] = useState(true);
  const scope = useFragment(scopeCardFragment, props.scopeRef);
  const { scopesConnectionId } = props;

  const nodes = scope.nodes?.edges.map(e => e.node) ?? [];
  const boundaries = scope.boundaries?.edges.map(e => e.node) ?? [];
  const processes = scope.processes?.edges.map(e => e.node) ?? [];
  const threats = scope.threats?.edges.map(e => e.node) ?? [];
  const scenarios = scope.scenarios?.edges.map(e => e.node) ?? [];
  const nodeMap = new Map(nodes.map(n => [n.id, n]));
  const boundaryMap = new Map(boundaries.map(b => [b.id, b]));
  const boundaryOptions = boundaries.map(b => ({ id: b.id, name: b.name }));
  const nodesConnId = scope.nodes?.__id ?? "";
  const boundariesConnId = scope.boundaries?.__id ?? "";
  const processesConnId = scope.processes?.__id ?? "";
  const threatsConnId = scope.threats?.__id ?? "";
  const scenariosConnId = scope.scenarios?.__id ?? "";

  const ChevronIcon = isOpen ? IconChevronDown : IconChevronRight;

  return (
    <Card>
      <button
        type="button"
        className="flex w-full items-center justify-between px-4 py-3"
        onClick={() => setIsOpen(v => !v)}
      >
        <div className="text-left">
          <h3 className="text-sm font-semibold">{scope.name}</h3>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-txt-tertiary">
            {nodes.length}
            {" "}
            {__("nodes")}
            {" · "}
            {processes.length}
            {" "}
            {__("processes")}
            {" · "}
            {threats.length}
            {" "}
            {__("threats")}
            {" · "}
            {scenarios.length}
            {" "}
            {__("scenarios")}
          </span>
          <div
            onClick={e => e.stopPropagation()}
            onKeyDown={e => e.stopPropagation()}
          >
            <ScopeActions
              scope={{ id: scope.id, name: scope.name }}
              connectionId={scopesConnectionId}
            />
          </div>
          <ChevronIcon size={16} className="text-txt-tertiary" />
        </div>
      </button>

      {isOpen && (
        <div className="border-t border-border-low px-4 py-4 space-y-6">
          <div>
            <div className="mb-3">
              <h3 className="text-sm font-semibold">{__("Diagram")}</h3>
              <p className="text-xs text-txt-tertiary mt-1">
                {__("Visualization of nodes, processes, and threats in this scope.")}
              </p>
            </div>
            <ScopeDiagram scopeKey={scope} />
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div>
              <SectionHeader
                title={`${__("Nodes")} (${nodes.length})`}
                hint={__("Entities, boundaries, assets, and data involved in this scope.")}
              >
                <CreateNodeDialog scopeId={scope.id} connectionId={nodesConnId} boundaries={boundaryOptions} />
              </SectionHeader>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{__("Name")}</Th>
                    <Th>{__("Type")}</Th>
                    <Th>{__("Boundary")}</Th>
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
                      <Td colSpan={4} className="text-center text-txt-secondary">{__("No nodes")}</Td>
                    </Tr>
                  )}
                </Tbody>
              </Table>
            </div>

            <div>
              <SectionHeader
                title={`${__("Processes")} (${processes.length})`}
                hint={__("Data flows and interactions between nodes.")}
              >
                <CreateProcessDialog
                  scopeId={scope.id}
                  nodes={nodes.map(n => ({ id: n.id, name: n.name }))}
                  connectionId={processesConnId}
                />
              </SectionHeader>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{__("Name")}</Th>
                    <Th>{__("From")}</Th>
                    <Th>{__("To")}</Th>
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
                      <Td colSpan={4} className="text-center text-txt-secondary">{__("No processes")}</Td>
                    </Tr>
                  )}
                </Tbody>
              </Table>
            </div>
          </div>

          <div>
            <SectionHeader
              title={`${__("Boundaries")} (${boundaries.length})`}
              hint={__("Groupings that contain nodes and can be nested inside other boundaries.")}
            >
              <CreateBoundaryDialog
                scopeId={scope.id}
                connectionId={boundariesConnId}
                boundaries={boundaryOptions}
              />
            </SectionHeader>
            <Table>
              <Thead>
                <Tr>
                  <Th>{__("Name")}</Th>
                  <Th>{__("Parent")}</Th>
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
                    <Td colSpan={3} className="text-center text-txt-secondary">{__("No boundaries")}</Td>
                  </Tr>
                )}
              </Tbody>
            </Table>
          </div>

          <div>
            <SectionHeader
              title={`${__("Threats")} (${threats.length})`}
              hint={__("Potential threats targeting a process. Link threats to risks via scenarios.")}
            >
              <CreateThreatDialog
                scopeId={scope.id}
                processes={processes.map(p => ({ id: p.id, name: p.name }))}
                connectionId={threatsConnId}
              />
            </SectionHeader>
            <Table>
              <Thead>
                <Tr>
                  <Th>{__("Threat")}</Th>
                  <Th>{__("Category")}</Th>
                  <Th>{__("Process")}</Th>
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
                    <Td colSpan={4} className="text-center text-txt-secondary">{__("No threats")}</Td>
                  </Tr>
                )}
              </Tbody>
            </Table>
          </div>

          <div>
            <SectionHeader
              title={`${__("Scenarios")} (${scenarios.length})`}
              hint={__("Risk scenarios linking threats to risks.")}
            >
              <CreateScenarioInScopeDialog
                scopeId={scope.id}
                threats={threats.map(t => ({ id: t.id, name: t.name }))}
                connectionId={scenariosConnId}
              />
            </SectionHeader>
            <Table>
              <Thead>
                <Tr>
                  <Th>{__("Scenario")}</Th>
                  <Th>{__("Risks")}</Th>
                  <Th>{__("Threats")}</Th>
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
                                  to={`/organizations/${organizationId}/risks/${risk.id}`}
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
                        <ScenarioInScopeActions
                          scenario={{
                            id: scenario.id,
                            name: scenario.name,
                            description: scenario.description ?? null,
                            risks: scenarioRisks,
                            threats: scenarioThreats,
                          }}
                          scopeThreats={threats.map(t => ({ id: t.id, name: t.name }))}
                          connectionId={scenariosConnId}
                        />
                      </Td>
                    </Tr>
                  );
                })}
                {scenarios.length === 0 && (
                  <Tr>
                    <Td colSpan={4} className="text-center text-txt-secondary">{__("No scenarios")}</Td>
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
