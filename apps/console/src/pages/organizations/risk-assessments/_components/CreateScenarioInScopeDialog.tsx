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
import { useState } from "react";
import { useForm } from "react-hook-form";
import { graphql, useMutation } from "react-relay";

import type { CreateScenarioInScopeDialogLinkThreatMutation } from "#/__generated__/core/CreateScenarioInScopeDialogLinkThreatMutation.graphql";
import type { CreateScenarioInScopeDialogMutation } from "#/__generated__/core/CreateScenarioInScopeDialogMutation.graphql";

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
      riskAssessmentScenario { id }
    }
  }
`;

export function CreateScenarioInScopeDialog(props: {
  scopeId: string;
  threats: { id: string; name: string }[];
  connectionId: string;
}) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [selectedThreats, setSelectedThreats] = useState<Map<string, string>>(new Map());
  const [createScenario, isCreating] = useMutation<CreateScenarioInScopeDialogMutation>(createScenarioMutation);
  const [linkThreat] = useMutation<CreateScenarioInScopeDialogLinkThreatMutation>(linkThreatMutation);
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
        reset();
        setSelectedThreats(new Map());
        dialogRef.current?.close();
      },
    });
  };

  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      trigger={<Button icon={IconPlusLarge} variant="secondary">{__("Add")}</Button>}
      title={<Breadcrumb items={[__("Scenarios"), __("Add Scenario")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name", { required: __("This field is required") })}
            type="text"
            error={formState.errors.name?.message}
            placeholder={__("e.g. Data breach via compromised API")}
          />
          <Field
            label={__("Description")}
            {...register("description")}
            type="textarea"
            rows={3}
          />
          {props.threats.length > 0 && (
            <div>
              <div className="text-sm font-medium mb-2">{__("Threats")}</div>
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
                  placeholder={__("Select a threat to link...")}
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
        </DialogContent>
        <DialogFooter><Button type="submit" disabled={isCreating}>{__("Add")}</Button></DialogFooter>
      </form>
    </Dialog>
  );
}
