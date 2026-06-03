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
import { graphql, useLazyLoadQuery, useMutation } from "react-relay";

import type { ScenarioInScopeActionsDeleteMutation } from "#/__generated__/core/ScenarioInScopeActionsDeleteMutation.graphql";
import type { ScenarioInScopeActionsLinkRiskMutation } from "#/__generated__/core/ScenarioInScopeActionsLinkRiskMutation.graphql";
import type { ScenarioInScopeActionsLinkThreatMutation } from "#/__generated__/core/ScenarioInScopeActionsLinkThreatMutation.graphql";
import type { ScenarioInScopeActionsRisksQuery } from "#/__generated__/core/ScenarioInScopeActionsRisksQuery.graphql";
import type { ScenarioInScopeActionsUnlinkRiskMutation } from "#/__generated__/core/ScenarioInScopeActionsUnlinkRiskMutation.graphql";
import type { ScenarioInScopeActionsUnlinkThreatMutation } from "#/__generated__/core/ScenarioInScopeActionsUnlinkThreatMutation.graphql";
import type { ScenarioInScopeActionsUpdateMutation } from "#/__generated__/core/ScenarioInScopeActionsUpdateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const updateScenarioMutation = graphql`
  mutation ScenarioInScopeActionsUpdateMutation($input: UpdateRiskAssessmentScenarioInput!) {
    updateRiskAssessmentScenario(input: $input) {
      riskAssessmentScenario {
        id name description
        risks(first: 10) { edges { node { id name } } }
        threats(first: 10) { edges { node { id name } } }
      }
    }
  }
`;

const deleteScenarioMutation = graphql`
  mutation ScenarioInScopeActionsDeleteMutation(
    $input: DeleteRiskAssessmentScenarioInput!
    $connections: [ID!]!
  ) {
    deleteRiskAssessmentScenario(input: $input) {
      deletedRiskAssessmentScenarioId @deleteEdge(connections: $connections)
    }
  }
`;

const linkThreatMutation = graphql`
  mutation ScenarioInScopeActionsLinkThreatMutation($input: LinkRiskAssessmentScenarioThreatInput!) {
    linkRiskAssessmentScenarioThreat(input: $input) {
      riskAssessmentScenario {
        id
        threats(first: 10) { edges { node { id name } } }
      }
    }
  }
`;

const unlinkThreatMutation = graphql`
  mutation ScenarioInScopeActionsUnlinkThreatMutation($input: UnlinkRiskAssessmentScenarioThreatInput!) {
    unlinkRiskAssessmentScenarioThreat(input: $input) {
      riskAssessmentScenario {
        id
        threats(first: 10) { edges { node { id name } } }
      }
    }
  }
`;

const linkRiskMutation = graphql`
  mutation ScenarioInScopeActionsLinkRiskMutation($input: LinkRiskAssessmentScenarioRiskInput!) {
    linkRiskAssessmentScenarioRisk(input: $input) {
      riskAssessmentScenario {
        id
        risks(first: 10) { edges { node { id name } } }
      }
      riskAssessmentScenarioEdge { node { id } }
    }
  }
`;

const unlinkRiskMutation = graphql`
  mutation ScenarioInScopeActionsUnlinkRiskMutation($input: UnlinkRiskAssessmentScenarioRiskInput!) {
    unlinkRiskAssessmentScenarioRisk(input: $input) {
      riskAssessmentScenario {
        id
        risks(first: 10) { edges { node { id name } } }
      }
      deletedRiskAssessmentScenarioId
    }
  }
`;

const risksQuery = graphql`
  query ScenarioInScopeActionsRisksQuery($organizationId: ID!) {
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
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const [linkRisk] = useMutation<ScenarioInScopeActionsLinkRiskMutation>(linkRiskMutation);
  const data = useLazyLoadQuery<ScenarioInScopeActionsRisksQuery>(
    risksQuery,
    { organizationId },
    { fetchPolicy: "store-or-network" },
  );
  const allRisks = data.node?.risks?.edges?.map(e => e.node) ?? [];
  const availableRisks = allRisks.filter(r => !props.linkedRiskIds.has(r.id));

  if (availableRisks.length === 0) {
    return <p className="text-xs text-txt-tertiary">{__("No more risks available.")}</p>;
  }

  return (
    <Select
      placeholder={__("Select a risk to link...")}
      onValueChange={(riskId) => {
        if (typeof riskId !== "string") return;
        linkRisk({
          variables: { input: { riskAssessmentScenarioId: props.scenarioId, riskId } },
        });
      }}
    >
      {availableRisks.map(r => (
        <Option key={r.id} value={r.id}>{r.name}</Option>
      ))}
    </Select>
  );
}

export function ScenarioInScopeActions(props: {
  scenario: {
    id: string;
    name: string;
    description: string | null;
    risks: readonly { id: string; name: string }[];
    threats: readonly { id: string; name: string }[];
  };
  scopeThreats: readonly { id: string; name: string }[];
  connectionId: string;
}) {
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const dialogRef = useDialogRef();
  const [updateScenario] = useMutation<ScenarioInScopeActionsUpdateMutation>(updateScenarioMutation);
  const [deleteScenario] = useMutation<ScenarioInScopeActionsDeleteMutation>(deleteScenarioMutation);
  const [linkThreat] = useMutation<ScenarioInScopeActionsLinkThreatMutation>(linkThreatMutation);
  const [unlinkThreat] = useMutation<ScenarioInScopeActionsUnlinkThreatMutation>(unlinkThreatMutation);
  const [unlinkRisk] = useMutation<ScenarioInScopeActionsUnlinkRiskMutation>(unlinkRiskMutation);
  const { register, handleSubmit } = useForm({
    values: { name: props.scenario.name, description: props.scenario.description ?? "" },
  });

  const linkedThreatIds = new Set(props.scenario.threats.map(t => t.id));
  const linkedRiskIds = new Set(props.scenario.risks.map(r => r.id));
  const availableThreats = props.scopeThreats.filter(t => !linkedThreatIds.has(t.id));

  return (
    <>
      <ActionDropdown>
        <DropdownItem icon={IconPencil} onSelect={() => dialogRef.current?.open()}>
          {__("Edit")}
        </DropdownItem>
        <DropdownItem
          icon={IconTrashCan}
          variant="danger"
          onSelect={() => confirm(
            () => {
              deleteScenario({
                variables: {
                  input: { riskAssessmentScenarioId: props.scenario.id },
                  connections: [props.connectionId],
                },
              });
            },
            { message: __("Delete this scenario?") },
          )}
        >
          {__("Delete")}
        </DropdownItem>
      </ActionDropdown>
      <Dialog className="max-w-lg" ref={dialogRef} title={<Breadcrumb items={[__("Scenarios"), __("Edit")]} />}>
        <form onSubmit={e => void handleSubmit((d) => {
          updateScenario({
            variables: { input: { id: props.scenario.id, name: d.name, description: d.description || null } },
            onCompleted: () => { dialogRef.current?.close(); },
          });
        })(e)}
        >
          <DialogContent padded className="space-y-4">
            <Field label={__("Name")} {...register("name", { required: __("This field is required") })} type="text" />
            <Field label={__("Description")} {...register("description")} type="textarea" rows={3} />

            <div>
              <div className="text-sm font-medium mb-2">{__("Threats")}</div>
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
                              input: { riskAssessmentScenarioId: props.scenario.id, threatId: threat.id },
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
                  placeholder={__("Select a threat to link...")}
                  onValueChange={(threatId) => {
                    if (typeof threatId !== "string") return;
                    linkThreat({ variables: { input: { riskAssessmentScenarioId: props.scenario.id, threatId } } });
                  }}
                >
                  {availableThreats.map(t => (
                    <Option key={t.id} value={t.id}>{t.name}</Option>
                  ))}
                </Select>
              )}
            </div>

            <div>
              <div className="text-sm font-medium mb-2">{__("Risks")}</div>
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
                              input: { riskAssessmentScenarioId: props.scenario.id, riskId: risk.id },
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
              <Suspense fallback={<p className="text-xs text-txt-tertiary">{__("Loading risks...")}</p>}>
                <RiskSelector
                  scenarioId={props.scenario.id}
                  linkedRiskIds={linkedRiskIds}
                />
              </Suspense>
            </div>
          </DialogContent>
          <DialogFooter><Button type="submit">{__("Save")}</Button></DialogFooter>
        </form>
      </Dialog>
    </>
  );
}
