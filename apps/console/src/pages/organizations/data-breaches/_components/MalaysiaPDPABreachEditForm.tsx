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

import { Button, Card } from "@probo/ui";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { MalaysiaPDPABreachEditForm_incident$key } from "#/__generated__/core/MalaysiaPDPABreachEditForm_incident.graphql";
import type { MalaysiaPDPABreachEditFormUpdateMutation } from "#/__generated__/core/MalaysiaPDPABreachEditFormUpdateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import {
  type BreachFormErrors,
  type BreachFormField,
  type BreachFormValues,
  breachValuesToMutationFields,
  createBreachFormSchema,
  toBreachFormErrors,
  toDateTimeLocal,
} from "../_lib/breachForm";

import { MalaysiaPDPABreachFormFields } from "./MalaysiaPDPABreachFormFields";

const incidentFragment = graphql`
  fragment MalaysiaPDPABreachEditForm_incident on MalaysiaPDPABreachIncident {
    id
    title
    description
    occurredAt
    discoveredAt
    awarenessAt
    affectedDataSubjects
    affectedDataRecords
    personalDataTypes
    affectedSystem
    likelyConsequences
    containmentActions
    potentialPhysicalHarm
    potentialFinancialLoss
    potentialCreditOrPropertyDamage
    potentialIllegalUse
    sensitivePersonalData
    potentialIdentityFraud
    notificationDecision
    decisionRationale
    decisionEvidence
    commissionerNotifiedAt
    commissionerNotificationReference
    commissionerConfirmationReceivedAt
    commissionerConfirmationReference
    delayedNotificationReason
    delayedNotificationEvidence
    dataSubjectsNotifiedAt
    dataSubjectsNotificationEvidence
    canUpdate: permission(action: "core:malaysia-pdpa-breach:update")
  }
`;

const updateMutation = graphql`
  mutation MalaysiaPDPABreachEditFormUpdateMutation(
    $input: UpdateMalaysiaPDPABreachIncidentInput!
  ) {
    updateMalaysiaPDPABreachIncident(input: $input) {
      incident {
        ...MalaysiaPDPABreachEditForm_incident
        ...MalaysiaPDPABreachSummarySection_incident
      }
    }
  }
`;

interface MalaysiaPDPABreachEditFormProps {
  incidentKey: MalaysiaPDPABreachEditForm_incident$key;
}

export function MalaysiaPDPABreachEditForm({
  incidentKey,
}: MalaysiaPDPABreachEditFormProps) {
  const { t } = useTranslation("organizations/data-breaches");
  const incident = useFragment(incidentFragment, incidentKey);
  const [values, setValues] = useState<BreachFormValues>(() => ({
    title: incident.title,
    description: incident.description ?? "",
    occurredAt: toDateTimeLocal(incident.occurredAt),
    discoveredAt: toDateTimeLocal(incident.discoveredAt),
    awarenessAt: toDateTimeLocal(incident.awarenessAt),
    affectedDataSubjects: String(incident.affectedDataSubjects),
    affectedDataRecords: String(incident.affectedDataRecords),
    personalDataTypes: incident.personalDataTypes,
    affectedSystem: incident.affectedSystem ?? "",
    likelyConsequences: incident.likelyConsequences ?? "",
    containmentActions: incident.containmentActions ?? "",
    potentialPhysicalHarm: incident.potentialPhysicalHarm,
    potentialFinancialLoss: incident.potentialFinancialLoss,
    potentialCreditOrPropertyDamage:
      incident.potentialCreditOrPropertyDamage,
    potentialIllegalUse: incident.potentialIllegalUse,
    sensitivePersonalData: incident.sensitivePersonalData,
    potentialIdentityFraud: incident.potentialIdentityFraud,
    notificationDecision: incident.notificationDecision,
    decisionRationale: incident.decisionRationale ?? "",
    decisionEvidence: incident.decisionEvidence ?? "",
    commissionerNotifiedAt: toDateTimeLocal(incident.commissionerNotifiedAt),
    commissionerNotificationReference:
      incident.commissionerNotificationReference ?? "",
    commissionerConfirmationReceivedAt: toDateTimeLocal(
      incident.commissionerConfirmationReceivedAt,
    ),
    commissionerConfirmationReference:
      incident.commissionerConfirmationReference ?? "",
    delayedNotificationReason: incident.delayedNotificationReason ?? "",
    delayedNotificationEvidence: incident.delayedNotificationEvidence ?? "",
    dataSubjectsNotifiedAt: toDateTimeLocal(incident.dataSubjectsNotifiedAt),
    dataSubjectsNotificationEvidence:
      incident.dataSubjectsNotificationEvidence ?? "",
  }));
  const [errors, setErrors] = useState<BreachFormErrors>({});
  const [updateMalaysiaPDPABreach, isUpdating]
    = useMutation<MalaysiaPDPABreachEditFormUpdateMutation>(updateMutation, {
      successMessage: t("messages.updated"),
      errorToast: t("errors.update"),
    });

  function updateValue<FieldName extends BreachFormField>(
    field: FieldName,
    value: BreachFormValues[FieldName],
  ) {
    setValues(current => ({ ...current, [field]: value }));
    setErrors(current => ({ ...current, [field]: undefined }));
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const result = createBreachFormSchema(t).safeParse(values);
    if (!result.success) {
      setErrors(toBreachFormErrors(result.error));
      return;
    }

    setErrors({});
    try {
      await updateMalaysiaPDPABreach({
        variables: {
          input: {
            id: incident.id,
            ...breachValuesToMutationFields(result.data),
          },
        },
      });
    } catch {
      // useMutation already displays the localized server error.
    }
  }

  return (
    <form onSubmit={event => void onSubmit(event)} className="space-y-6">
      {!incident.canUpdate && (
        <Card padded className="border-border-mid bg-subtle">
          <p className="text-sm text-txt-secondary">
            {t("detail.readOnly")}
          </p>
        </Card>
      )}
      <MalaysiaPDPABreachFormFields
        values={values}
        errors={errors}
        disabled={!incident.canUpdate || isUpdating}
        onValueChange={updateValue}
      />
      {incident.canUpdate && (
        <div className="flex justify-end">
          <Button type="submit" disabled={isUpdating}>
            {isUpdating ? t("common.saving") : t("common.save")}
          </Button>
        </div>
      )}
    </form>
  );
}
