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
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DropdownItem,
  Field,
  IconCrossLargeX,
  IconPencil,
  IconTrashCan,
  Option,
  Select,
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { Suspense } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, useMutation } from "react-relay";

import type { ScenarioInDiagramActionsDeleteMutation } from "#/__generated__/core/ScenarioInDiagramActionsDeleteMutation.graphql";
import type { ScenarioInDiagramActionsLinkRiskMutation } from "#/__generated__/core/ScenarioInDiagramActionsLinkRiskMutation.graphql";
import type { ScenarioInDiagramActionsLinkThreatMutation } from "#/__generated__/core/ScenarioInDiagramActionsLinkThreatMutation.graphql";
import type { ScenarioInDiagramActionsRisksQuery } from "#/__generated__/core/ScenarioInDiagramActionsRisksQuery.graphql";
import type { ScenarioInDiagramActionsUnlinkRiskMutation } from "#/__generated__/core/ScenarioInDiagramActionsUnlinkRiskMutation.graphql";
import type { ScenarioInDiagramActionsUnlinkThreatMutation } from "#/__generated__/core/ScenarioInDiagramActionsUnlinkThreatMutation.graphql";
import type { ScenarioInDiagramActionsUpdateMutation } from "#/__generated__/core/ScenarioInDiagramActionsUpdateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const updateScenarioMutation = graphql`
  mutation ScenarioInDiagramActionsUpdateMutation($input: UpdateRiskAnalysisScenarioInput!) {
    updateRiskAnalysisScenario(input: $input) {
      riskAnalysisScenario {
        id name description
        risks(first: 10) { edges { node { id name } } }
        threats(first: 10) { edges { node { id name } } }
      }
    }
  }
`;

const deleteScenarioMutation = graphql`
  mutation ScenarioInDiagramActionsDeleteMutation(
    $input: DeleteRiskAnalysisScenarioInput!
    $connections: [ID!]!
  ) {
    deleteRiskAnalysisScenario(input: $input) {
      deletedRiskAnalysisScenarioId @deleteEdge(connections: $connections)
    }
  }
`;

const linkThreatMutation = graphql`
  mutation ScenarioInDiagramActionsLinkThreatMutation($input: LinkRiskAnalysisScenarioThreatInput!) {
    linkRiskAnalysisScenarioThreat(input: $input) {
      riskAnalysisScenario {
        id
        threats(first: 10) { edges { node { id name } } }
      }
    }
  }
`;

const unlinkThreatMutation = graphql`
  mutation ScenarioInDiagramActionsUnlinkThreatMutation($input: UnlinkRiskAnalysisScenarioThreatInput!) {
    unlinkRiskAnalysisScenarioThreat(input: $input) {
      riskAnalysisScenario {
        id
        threats(first: 10) { edges { node { id name } } }
      }
    }
  }
`;

const linkRiskMutation = graphql`
  mutation ScenarioInDiagramActionsLinkRiskMutation($input: LinkRiskAnalysisScenarioRiskInput!) {
    linkRiskAnalysisScenarioRisk(input: $input) {
      riskAnalysisScenario {
        id
        risks(first: 10) { edges { node { id name } } }
      }
      riskAnalysisScenarioEdge { node { id } }
    }
  }
`;

const unlinkRiskMutation = graphql`
  mutation ScenarioInDiagramActionsUnlinkRiskMutation($input: UnlinkRiskAnalysisScenarioRiskInput!) {
    unlinkRiskAnalysisScenarioRisk(input: $input) {
      riskAnalysisScenario {
        id
        risks(first: 10) { edges { node { id name } } }
      }
      deletedRiskAnalysisScenarioId
    }
  }
`;

const risksQuery = graphql`
  query ScenarioInDiagramActionsRisksQuery($organizationId: ID!) {
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
  scenarioId: string;
  linkedRiskIds: Set<string>;
}) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const [linkRisk] = useMutation<ScenarioInDiagramActionsLinkRiskMutation>(linkRiskMutation);
  const data = useLazyLoadQuery<ScenarioInDiagramActionsRisksQuery>(
    risksQuery,
    { organizationId },
    { fetchPolicy: "store-or-network" },
  );
  const allRisks = data.node?.risks?.edges?.map(e => e.node) ?? [];
  const availableRisks = allRisks.filter(r => !props.linkedRiskIds.has(r.id));

  if (availableRisks.length === 0) {
    return <p className="text-xs text-txt-tertiary">{t("riskAnalysisScenarioActions.noMoreRisks")}</p>;
  }

  return (
    <Select
      placeholder={t("riskAnalysisScenarioActions.placeholders.riskToLink")}
      onValueChange={(riskId) => {
        if (typeof riskId !== "string") return;
        linkRisk({
          variables: { input: { riskAnalysisScenarioId: props.scenarioId, riskId } },
        });
      }}
    >
      {availableRisks.map(r => (
        <Option key={r.id} value={r.id}>{r.name}</Option>
      ))}
    </Select>
  );
}

export function ScenarioInDiagramActions(props: {
  scenario: {
    id: string;
    name: string;
    description: string | null;
    risks: readonly { id: string; name: string }[];
    threats: readonly { id: string; name: string }[];
  };
  diagramThreats: readonly { id: string; name: string }[];
  connectionId: string;
}) {
  const { t } = useTranslation();
  const confirm = useConfirm();
  const dialogRef = useDialogRef();
  const [updateScenario] = useMutation<ScenarioInDiagramActionsUpdateMutation>(updateScenarioMutation);
  const [deleteScenario] = useMutation<ScenarioInDiagramActionsDeleteMutation>(deleteScenarioMutation);
  const [linkThreat] = useMutation<ScenarioInDiagramActionsLinkThreatMutation>(linkThreatMutation);
  const [unlinkThreat] = useMutation<ScenarioInDiagramActionsUnlinkThreatMutation>(unlinkThreatMutation);
  const [unlinkRisk] = useMutation<ScenarioInDiagramActionsUnlinkRiskMutation>(unlinkRiskMutation);
  const { register, handleSubmit } = useForm({
    values: { name: props.scenario.name, description: props.scenario.description ?? "" },
  });

  const linkedThreatIds = new Set(props.scenario.threats.map(t => t.id));
  const linkedRiskIds = new Set(props.scenario.risks.map(r => r.id));
  const availableThreats = props.diagramThreats.filter(t => !linkedThreatIds.has(t.id));

  return (
    <>
      <ActionDropdown>
        <DropdownItem icon={IconPencil} onSelect={() => dialogRef.current?.open()}>
          {t("riskAnalysisScenarioActions.actions.edit")}
        </DropdownItem>
        <DropdownItem
          icon={IconTrashCan}
          variant="danger"
          onSelect={() => confirm(
            () => {
              deleteScenario({
                variables: {
                  input: { riskAnalysisScenarioId: props.scenario.id },
                  connections: [props.connectionId],
                },
              });
            },
            { message: t("riskAnalysisScenarioActions.deleteConfirmation") },
          )}
        >
          {t("riskAnalysisScenarioActions.actions.delete")}
        </DropdownItem>
      </ActionDropdown>
      <Dialog className="max-w-lg" ref={dialogRef} title={<Breadcrumb items={[t("riskAnalysisScenarioActions.breadcrumb.scenarios"), t("riskAnalysisScenarioActions.actions.edit")]} />}>
        <form onSubmit={e => void handleSubmit((d) => {
          updateScenario({
            variables: { input: { id: props.scenario.id, name: d.name, description: d.description || null } },
            onCompleted: () => { dialogRef.current?.close(); },
          });
        })(e)}
        >
          <DialogContent padded className="space-y-4">
            <Field label={t("riskAnalysisScenarioActions.fields.name")} {...register("name", { required: t("riskAnalysisScenarioActions.validation.nameRequired") })} type="text" />
            <Field label={t("riskAnalysisScenarioActions.fields.description")} {...register("description")} type="textarea" rows={3} />

            <div>
              <div className="text-sm font-medium mb-2">{t("riskAnalysisScenarioActions.fields.threats")}</div>
              {props.scenario.threats.length > 0 && (
                <div className="flex flex-wrap gap-1 mb-2">
                  {props.scenario.threats.map(threat => (
                    <Badge key={threat.id}>
                      {threat.name}
                      <button
                        type="button"
                        className="ml-1 hover:text-txt-danger"
                        onClick={() => {
                          unlinkThreat({
                            variables: {
                              input: { riskAnalysisScenarioId: props.scenario.id, threatId: threat.id },
                            },
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
                  placeholder={t("riskAnalysisScenarioActions.placeholders.threatToLink")}
                  onValueChange={(threatId) => {
                    if (typeof threatId !== "string") return;
                    linkThreat({ variables: { input: { riskAnalysisScenarioId: props.scenario.id, threatId } } });
                  }}
                >
                  {availableThreats.map(t => (
                    <Option key={t.id} value={t.id}>{t.name}</Option>
                  ))}
                </Select>
              )}
            </div>

            <div>
              <div className="text-sm font-medium mb-2">{t("riskAnalysisScenarioActions.fields.risks")}</div>
              {props.scenario.risks.length > 0 && (
                <div className="flex flex-wrap gap-1 mb-2">
                  {props.scenario.risks.map(risk => (
                    <Badge key={risk.id}>
                      {risk.name}
                      <button
                        type="button"
                        className="ml-1 hover:text-txt-danger"
                        onClick={() => {
                          unlinkRisk({
                            variables: {
                              input: { riskAnalysisScenarioId: props.scenario.id, riskId: risk.id },
                            },
                          });
                        }}
                      >
                        <IconCrossLargeX size={12} />
                      </button>
                    </Badge>
                  ))}
                </div>
              )}
              <Suspense fallback={<p className="text-xs text-txt-tertiary">{t("riskAnalysisScenarioActions.loadingRisks")}</p>}>
                <RiskSelector
                  scenarioId={props.scenario.id}
                  linkedRiskIds={linkedRiskIds}
                />
              </Suspense>
            </div>
          </DialogContent>
          <DialogFooter><Button type="submit">{t("riskAnalysisScenarioActions.actions.save")}</Button></DialogFooter>
        </form>
      </Dialog>
    </>
  );
}
