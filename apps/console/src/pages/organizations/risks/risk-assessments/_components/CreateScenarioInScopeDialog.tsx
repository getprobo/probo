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
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  IconCrossLargeX,
  IconPlusLarge,
  Option,
  Select,
  useDialogRef,
} from "@probo/ui";
import { Suspense, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, useMutation } from "react-relay";

import type { CreateScenarioInScopeDialogLinkRiskMutation } from "#/__generated__/core/CreateScenarioInScopeDialogLinkRiskMutation.graphql";
import type { CreateScenarioInScopeDialogLinkThreatMutation } from "#/__generated__/core/CreateScenarioInScopeDialogLinkThreatMutation.graphql";
import type { CreateScenarioInScopeDialogMutation } from "#/__generated__/core/CreateScenarioInScopeDialogMutation.graphql";
import type { CreateScenarioInScopeDialogRisksQuery } from "#/__generated__/core/CreateScenarioInScopeDialogRisksQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const createScenarioMutation = graphql`
  mutation CreateScenarioInScopeDialogMutation(
    $input: CreateRiskAssessmentScenarioInput!
    $connections: [ID!]!
  ) {
    createRiskAssessmentScenario(input: $input) {
      riskAssessmentScenarioEdge @appendEdge(connections: $connections) {
        node {
          id name description
          risks(first: 10) { edges { node { id name } } }
          threats(first: 10) { edges { node { id name } } }
        }
      }
    }
  }
`;

const linkThreatMutation = graphql`
  mutation CreateScenarioInScopeDialogLinkThreatMutation(
    $input: LinkRiskAssessmentScenarioThreatInput!
  ) {
    linkRiskAssessmentScenarioThreat(input: $input) {
      riskAssessmentScenario {
        id
        threats(first: 10) { edges { node { id name } } }
      }
    }
  }
`;

const linkRiskMutation = graphql`
  mutation CreateScenarioInScopeDialogLinkRiskMutation(
    $input: LinkRiskAssessmentScenarioRiskInput!
  ) {
    linkRiskAssessmentScenarioRisk(input: $input) {
      riskAssessmentScenario {
        id
        risks(first: 10) { edges { node { id name } } }
      }
      riskAssessmentScenarioEdge { node { id } }
    }
  }
`;

const risksQuery = graphql`
  query CreateScenarioInScopeDialogRisksQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        risks(first: 100) {
          edges { node { id name } }
        }
      }
    }
  }
`;

function RiskSelector(props: {
  selectedRisks: Map<string, string>;
  onSelect: (id: string, name: string) => void;
}) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const data = useLazyLoadQuery<CreateScenarioInScopeDialogRisksQuery>(
    risksQuery,
    { organizationId },
    { fetchPolicy: "store-or-network" },
  );
  const allRisks = data.node?.risks?.edges?.map(e => e.node) ?? [];
  const available = allRisks.filter(r => !props.selectedRisks.has(r.id));

  if (available.length === 0) {
    return <p className="text-xs text-txt-tertiary">{t("createScenarioInScopeDialog.noMoreRisks")}</p>;
  }

  return (
    <Select
      key={props.selectedRisks.size}
      placeholder={t("createScenarioInScopeDialog.placeholders.riskToLink")}
      onValueChange={(riskId) => {
        if (typeof riskId !== "string") return;
        const risk = allRisks.find(r => r.id === riskId);
        if (risk) props.onSelect(risk.id, risk.name);
      }}
    >
      {available.map(r => (
        <Option key={r.id} value={r.id}>{r.name}</Option>
      ))}
    </Select>
  );
}

export function CreateScenarioInScopeDialog(props: {
  scopeId: string;
  threats: { id: string; name: string }[];
  connectionId: string;
}) {
  const { t } = useTranslation();
  const dialogRef = useDialogRef();
  const [selectedThreats, setSelectedThreats] = useState<Map<string, string>>(new Map());
  const [selectedRisks, setSelectedRisks] = useState<Map<string, string>>(new Map());
  const [createScenario, isCreating] = useMutation<CreateScenarioInScopeDialogMutation>(createScenarioMutation);
  const [linkThreat] = useMutation<CreateScenarioInScopeDialogLinkThreatMutation>(linkThreatMutation);
  const [linkRisk] = useMutation<CreateScenarioInScopeDialogLinkRiskMutation>(linkRiskMutation);
  const { register, handleSubmit, reset, formState } = useForm({
    defaultValues: { name: "", description: "" },
  });

  const availableThreats = props.threats.filter(t => !selectedThreats.has(t.id));

  const onSubmit = (data: { name: string; description: string }) => {
    createScenario({
      variables: {
        input: {
          riskAssessmentScopeId: props.scopeId,
          name: data.name,
          description: data.description || null,
        },
        connections: [props.connectionId],
      },
      onCompleted(response) {
        const scenarioId = response.createRiskAssessmentScenario.riskAssessmentScenarioEdge.node.id;
        for (const threatId of selectedThreats.keys()) {
          linkThreat({
            variables: {
              input: { riskAssessmentScenarioId: scenarioId, threatId },
            },
          });
        }
        for (const riskId of selectedRisks.keys()) {
          linkRisk({
            variables: {
              input: { riskAssessmentScenarioId: scenarioId, riskId },
            },
          });
        }
        reset();
        setSelectedThreats(new Map());
        setSelectedRisks(new Map());
        dialogRef.current?.close();
      },
    });
  };

  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      trigger={<Button icon={IconPlusLarge} variant="secondary">{t("createScenarioInScopeDialog.actions.add")}</Button>}
      title={<Breadcrumb items={[t("createScenarioInScopeDialog.breadcrumb.scenarios"), t("createScenarioInScopeDialog.breadcrumb.addScenario")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={t("createScenarioInScopeDialog.fields.name")}
            {...register("name", { required: t("createScenarioInScopeDialog.validation.nameRequired") })}
            type="text"
            error={formState.errors.name?.message}
            placeholder={t("createScenarioInScopeDialog.placeholders.name")}
          />
          <Field
            label={t("createScenarioInScopeDialog.fields.description")}
            {...register("description")}
            type="textarea"
            rows={3}
          />
          {props.threats.length > 0 && (
            <div>
              <div className="text-sm font-medium mb-2">{t("createScenarioInScopeDialog.fields.threats")}</div>
              {selectedThreats.size > 0 && (
                <div className="flex flex-wrap gap-1 mb-2">
                  {[...selectedThreats.entries()].map(([id, name]) => (
                    <Badge key={id}>
                      {name}
                      <button
                        type="button"
                        className="ml-1 hover:text-txt-danger"
                        onClick={() => {
                          setSelectedThreats((prev) => {
                            const next = new Map(prev);
                            next.delete(id);
                            return next;
                          });
                        }}
                      >
                        <IconCrossLargeX size={12} />
                      </button>
                    </Badge>
                  ))}
                </div>
              )}
              {availableThreats.length > 0 && (
                <Select
                  key={selectedThreats.size}
                  placeholder={t("createScenarioInScopeDialog.placeholders.threatToLink")}
                  onValueChange={(threatId) => {
                    if (typeof threatId !== "string") return;
                    const threat = props.threats.find(t => t.id === threatId);
                    if (threat) {
                      setSelectedThreats((prev) => {
                        const next = new Map(prev);
                        next.set(threat.id, threat.name);
                        return next;
                      });
                    }
                  }}
                >
                  {availableThreats.map(t => (
                    <Option key={t.id} value={t.id}>{t.name}</Option>
                  ))}
                </Select>
              )}
            </div>
          )}

          <div>
            <div className="text-sm font-medium mb-2">{t("createScenarioInScopeDialog.fields.risks")}</div>
            {selectedRisks.size > 0 && (
              <div className="flex flex-wrap gap-1 mb-2">
                {[...selectedRisks.entries()].map(([id, name]) => (
                  <Badge key={id}>
                    {name}
                    <button
                      type="button"
                      className="ml-1 hover:text-txt-danger"
                      onClick={() => {
                        setSelectedRisks((prev) => {
                          const next = new Map(prev);
                          next.delete(id);
                          return next;
                        });
                      }}
                    >
                      <IconCrossLargeX size={12} />
                    </button>
                  </Badge>
                ))}
              </div>
            )}
            <Suspense fallback={<p className="text-xs text-txt-tertiary">{t("createScenarioInScopeDialog.loadingRisks")}</p>}>
              <RiskSelector
                selectedRisks={selectedRisks}
                onSelect={(id, name) => {
                  setSelectedRisks((prev) => {
                    const next = new Map(prev);
                    next.set(id, name);
                    return next;
                  });
                }}
              />
            </Suspense>
          </div>
        </DialogContent>
        <DialogFooter><Button type="submit" disabled={isCreating}>{t("createScenarioInScopeDialog.actions.add")}</Button></DialogFooter>
      </form>
    </Dialog>
  );
}
