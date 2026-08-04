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

import { Badge, Button, Card, Field, Option, Select } from "@probo/ui";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { ConnectionHandler } from "relay-runtime";

import type { MalaysiaPDPABreachTransitionForm_incident$key } from "#/__generated__/core/MalaysiaPDPABreachTransitionForm_incident.graphql";
import type { MalaysiaPDPABreachTransitionFormMutation } from "#/__generated__/core/MalaysiaPDPABreachTransitionFormMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import {
  getAllowedBreachStatusTransitions,
  getBreachStatusBadgeVariant,
  MALAYSIA_PDPA_BREACH_HISTORY_CONNECTION_KEY,
  type MalaysiaPDPABreachStatus,
} from "../_lib/breachDisplay";

const incidentFragment = graphql`
  fragment MalaysiaPDPABreachTransitionForm_incident on MalaysiaPDPABreachIncident {
    id
    status
    canTransition: permission(action: "core:malaysia-pdpa-breach:transition")
  }
`;

const transitionMutation = graphql`
  mutation MalaysiaPDPABreachTransitionFormMutation(
    $input: TransitionMalaysiaPDPABreachStatusInput!
    $connections: [ID!]!
  ) {
    transitionMalaysiaPDPABreachStatus(input: $input) {
      incident {
        status
        ...MalaysiaPDPABreachTransitionForm_incident
        ...MalaysiaPDPABreachSummarySection_incident
        ...MalaysiaPDPABreachListItem_incident
      }
      historyEdge @prependEdge(connections: $connections) {
        node {
          id
          ...MalaysiaPDPABreachStatusHistoryListItem_history
        }
      }
    }
  }
`;

interface MalaysiaPDPABreachTransitionFormProps {
  incidentKey: MalaysiaPDPABreachTransitionForm_incident$key;
}

export function MalaysiaPDPABreachTransitionForm({
  incidentKey,
}: MalaysiaPDPABreachTransitionFormProps) {
  const { t } = useTranslation("organizations/data-breaches");
  const incident = useFragment(incidentFragment, incidentKey);
  const allowedTransitions = getAllowedBreachStatusTransitions(incident.status);
  const [toStatus, setToStatus] = useState<MalaysiaPDPABreachStatus>(
    allowedTransitions[0],
  );
  const [reason, setReason] = useState("");
  const [transitionMalaysiaPDPABreach, isTransitioning]
    = useMutation<MalaysiaPDPABreachTransitionFormMutation>(
      transitionMutation,
      {
        successMessage: t("messages.transitioned"),
        errorToast: t("errors.transition"),
      },
    );

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const connectionId = ConnectionHandler.getConnectionID(
      incident.id,
      MALAYSIA_PDPA_BREACH_HISTORY_CONNECTION_KEY,
    );

    try {
      const response = await transitionMalaysiaPDPABreach({
        variables: {
          connections: [connectionId],
          input: {
            id: incident.id,
            toStatus,
            reason: reason.trim() || null,
          },
        },
      });
      const nextAllowed = getAllowedBreachStatusTransitions(
        response.transitionMalaysiaPDPABreachStatus.incident.status,
      );
      setToStatus(nextAllowed[0]);
      setReason("");
    } catch {
      // useMutation already displays the localized server error.
    }
  }

  return (
    <Card padded className="space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-1">
          <h2 className="text-base font-medium">{t("transition.title")}</h2>
          <p className="text-sm text-txt-tertiary">
            {t("transition.description")}
          </p>
        </div>
        <Badge variant={getBreachStatusBadgeVariant(incident.status)}>
          {t(`statuses.${incident.status}`)}
        </Badge>
      </div>

      {incident.canTransition
        ? (
            <form
              onSubmit={event => void onSubmit(event)}
              className="grid items-end gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,2fr)_auto]"
            >
              <Field
                name="toStatus"
                label={t("transition.fields.toStatus")}
              >
                <Select<MalaysiaPDPABreachStatus>
                  value={toStatus}
                  disabled={isTransitioning}
                  onValueChange={value =>
                    setToStatus(value as MalaysiaPDPABreachStatus)}
                >
                  {allowedTransitions.map(status => (
                    <Option key={status} value={status}>
                      {t(`statuses.${status}`)}
                    </Option>
                  ))}
                </Select>
              </Field>
              <Field
                name="transitionReason"
                type="text"
                maxLength={5000}
                label={t("transition.fields.reason")}
                placeholder={t("transition.fields.reasonPlaceholder")}
                value={reason}
                disabled={isTransitioning}
                onValueChange={setReason}
              />
              <Button type="submit" disabled={isTransitioning}>
                {isTransitioning
                  ? t("transition.actions.updating")
                  : t("transition.actions.update")}
              </Button>
            </form>
          )
        : (
            <p className="text-sm text-txt-secondary">
              {t("transition.readOnly")}
            </p>
          )}
      {incident.status === "CONTAINED" && (
        <p className="text-xs text-txt-tertiary">
          {t("transition.closeRequirement")}
        </p>
      )}
    </Card>
  );
}
