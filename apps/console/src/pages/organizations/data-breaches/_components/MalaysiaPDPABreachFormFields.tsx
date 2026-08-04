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
  Card,
  Checkbox,
  Field,
  Input,
  Label,
  Option,
  Select,
} from "@probo/ui";
import { useTranslation } from "react-i18next";

import type {
  BreachFormErrors,
  BreachFormField,
  BreachFormValues,
  MalaysiaPDPABreachNotificationDecision,
} from "../_lib/breachForm";

const notificationDecisions: MalaysiaPDPABreachNotificationDecision[] = [
  "PENDING",
  "NOT_REQUIRED",
  "COMMISSIONER_ONLY",
  "COMMISSIONER_AND_DATA_SUBJECTS",
];

const harmFields = [
  "potentialPhysicalHarm",
  "potentialFinancialLoss",
  "potentialCreditOrPropertyDamage",
  "potentialIllegalUse",
  "sensitivePersonalData",
  "potentialIdentityFraud",
] as const;

interface MalaysiaPDPABreachFormFieldsProps {
  disabled: boolean;
  errors: BreachFormErrors;
  values: BreachFormValues;
  onValueChange: <FieldName extends BreachFormField>(
    field: FieldName,
    value: BreachFormValues[FieldName],
  ) => void;
}

export function MalaysiaPDPABreachFormFields(
  props: MalaysiaPDPABreachFormFieldsProps,
) {
  const { disabled, errors, values, onValueChange } = props;
  const { t } = useTranslation("organizations/data-breaches");

  return (
    <>
      <Card padded className="space-y-5">
        <SectionHeader
          title={t("form.incident.title")}
          description={t("form.incident.description")}
        />

        <Field
          name="title"
          type="text"
          required
          maxLength={1000}
          label={t("form.fields.title.label")}
          placeholder={t("form.fields.title.placeholder")}
          value={values.title}
          error={errors.title}
          disabled={disabled}
          onValueChange={value => onValueChange("title", value)}
        />
        <Field
          name="description"
          type="textarea"
          rows={4}
          maxLength={5000}
          label={t("form.fields.description.label")}
          placeholder={t("form.fields.description.placeholder")}
          value={values.description}
          error={errors.description}
          disabled={disabled}
          onValueChange={value => onValueChange("description", value)}
        />

        <div className="grid gap-4 md:grid-cols-3">
          <Field
            name="occurredAt"
            label={t("form.fields.occurredAt.label")}
            help={t("form.fields.occurredAt.help")}
            error={errors.occurredAt}
          >
            <Input
              name="occurredAt"
              type="datetime-local"
              value={values.occurredAt}
              disabled={disabled}
              onValueChange={value => onValueChange("occurredAt", value)}
            />
          </Field>
          <Field
            name="discoveredAt"
            label={t("form.fields.discoveredAt.label")}
            help={t("form.fields.discoveredAt.help")}
            error={errors.discoveredAt}
          >
            <Input
              name="discoveredAt"
              type="datetime-local"
              required
              value={values.discoveredAt}
              disabled={disabled}
              onValueChange={value => onValueChange("discoveredAt", value)}
            />
          </Field>
          <Field
            name="awarenessAt"
            label={t("form.fields.awarenessAt.label")}
            help={t("form.fields.awarenessAt.help")}
            error={errors.awarenessAt}
          >
            <Input
              name="awarenessAt"
              type="datetime-local"
              required
              value={values.awarenessAt}
              disabled={disabled}
              onValueChange={value => onValueChange("awarenessAt", value)}
            />
          </Field>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <Field
            name="affectedDataSubjects"
            type="number"
            min={0}
            step={1}
            required
            label={t("form.fields.affectedDataSubjects.label")}
            help={t("form.fields.affectedDataSubjects.help")}
            value={values.affectedDataSubjects}
            error={errors.affectedDataSubjects}
            disabled={disabled}
            onValueChange={value =>
              onValueChange("affectedDataSubjects", value)}
          />
          <Field
            name="affectedDataRecords"
            type="number"
            min={0}
            step={1}
            required
            label={t("form.fields.affectedDataRecords.label")}
            help={t("form.fields.affectedDataRecords.help")}
            value={values.affectedDataRecords}
            error={errors.affectedDataRecords}
            disabled={disabled}
            onValueChange={value => onValueChange("affectedDataRecords", value)}
          />
        </div>

        <Field
          name="personalDataTypes"
          type="textarea"
          rows={3}
          required
          maxLength={5000}
          label={t("form.fields.personalDataTypes.label")}
          help={t("form.fields.personalDataTypes.help")}
          value={values.personalDataTypes}
          error={errors.personalDataTypes}
          disabled={disabled}
          onValueChange={value => onValueChange("personalDataTypes", value)}
        />
        <Field
          name="affectedSystem"
          type="text"
          maxLength={5000}
          label={t("form.fields.affectedSystem.label")}
          value={values.affectedSystem}
          error={errors.affectedSystem}
          disabled={disabled}
          onValueChange={value => onValueChange("affectedSystem", value)}
        />
        <div className="grid gap-4 md:grid-cols-2">
          <Field
            name="likelyConsequences"
            type="textarea"
            rows={4}
            maxLength={5000}
            label={t("form.fields.likelyConsequences.label")}
            value={values.likelyConsequences}
            error={errors.likelyConsequences}
            disabled={disabled}
            onValueChange={value => onValueChange("likelyConsequences", value)}
          />
          <Field
            name="containmentActions"
            type="textarea"
            rows={4}
            maxLength={5000}
            label={t("form.fields.containmentActions.label")}
            value={values.containmentActions}
            error={errors.containmentActions}
            disabled={disabled}
            onValueChange={value => onValueChange("containmentActions", value)}
          />
        </div>
      </Card>

      <Card padded className="space-y-5">
        <SectionHeader
          title={t("form.harm.title")}
          description={t("form.harm.description")}
        />
        <div className="grid gap-4 md:grid-cols-2">
          {harmFields.map(field => (
            <Label
              key={field}
              className="flex items-start gap-3 mb-0 cursor-pointer"
            >
              <Checkbox
                checked={values[field]}
                disabled={disabled}
                onChange={checked => onValueChange(field, checked)}
              />
              <span className="space-y-1">
                <span className="block text-sm font-medium">
                  {t(`form.harm.fields.${field}.label`)}
                </span>
                <span className="block text-xs font-normal text-txt-tertiary">
                  {t(`form.harm.fields.${field}.help`)}
                </span>
              </span>
            </Label>
          ))}
        </div>
      </Card>

      <Card padded className="space-y-5">
        <SectionHeader
          title={t("form.decision.title")}
          description={t("form.decision.description")}
        />
        <Field
          name="notificationDecision"
          label={t("form.fields.notificationDecision.label")}
          help={t("form.fields.notificationDecision.help")}
          error={errors.notificationDecision}
        >
          <Select<MalaysiaPDPABreachNotificationDecision>
            value={values.notificationDecision}
            disabled={disabled}
            onValueChange={value =>
              onValueChange(
                "notificationDecision",
                value as MalaysiaPDPABreachNotificationDecision,
              )}
          >
            {notificationDecisions.map(decision => (
              <Option key={decision} value={decision}>
                {t(`decisions.${decision}`)}
              </Option>
            ))}
          </Select>
        </Field>
        <div className="grid gap-4 md:grid-cols-2">
          <Field
            name="decisionRationale"
            type="textarea"
            rows={4}
            maxLength={5000}
            label={t("form.fields.decisionRationale.label")}
            help={t("form.fields.decisionRationale.help")}
            value={values.decisionRationale}
            error={errors.decisionRationale}
            disabled={disabled}
            onValueChange={value => onValueChange("decisionRationale", value)}
          />
          <Field
            name="decisionEvidence"
            type="textarea"
            rows={4}
            maxLength={5000}
            label={t("form.fields.decisionEvidence.label")}
            help={t("form.fields.decisionEvidence.help")}
            value={values.decisionEvidence}
            error={errors.decisionEvidence}
            disabled={disabled}
            onValueChange={value => onValueChange("decisionEvidence", value)}
          />
        </div>
      </Card>

      <Card padded className="space-y-5">
        <SectionHeader
          title={t("form.notifications.title")}
          description={t("form.notifications.description")}
        />

        <div className="grid gap-4 md:grid-cols-2">
          <Field
            name="commissionerNotifiedAt"
            label={t("form.fields.commissionerNotifiedAt.label")}
            error={errors.commissionerNotifiedAt}
          >
            <Input
              name="commissionerNotifiedAt"
              type="datetime-local"
              value={values.commissionerNotifiedAt}
              disabled={disabled}
              onValueChange={value =>
                onValueChange("commissionerNotifiedAt", value)}
            />
          </Field>
          <Field
            name="commissionerNotificationReference"
            type="text"
            maxLength={1000}
            label={t("form.fields.commissionerNotificationReference.label")}
            help={t("form.fields.commissionerNotificationReference.help")}
            value={values.commissionerNotificationReference}
            error={errors.commissionerNotificationReference}
            disabled={disabled}
            onValueChange={value =>
              onValueChange("commissionerNotificationReference", value)}
          />
          <Field
            name="commissionerConfirmationReceivedAt"
            label={t("form.fields.commissionerConfirmationReceivedAt.label")}
            error={errors.commissionerConfirmationReceivedAt}
          >
            <Input
              name="commissionerConfirmationReceivedAt"
              type="datetime-local"
              value={values.commissionerConfirmationReceivedAt}
              disabled={disabled}
              onValueChange={value =>
                onValueChange("commissionerConfirmationReceivedAt", value)}
            />
          </Field>
          <Field
            name="commissionerConfirmationReference"
            type="text"
            maxLength={1000}
            label={t("form.fields.commissionerConfirmationReference.label")}
            value={values.commissionerConfirmationReference}
            error={errors.commissionerConfirmationReference}
            disabled={disabled}
            onValueChange={value =>
              onValueChange("commissionerConfirmationReference", value)}
          />
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <Field
            name="delayedNotificationReason"
            type="textarea"
            rows={3}
            maxLength={5000}
            label={t("form.fields.delayedNotificationReason.label")}
            help={t("form.fields.delayedNotificationReason.help")}
            value={values.delayedNotificationReason}
            error={errors.delayedNotificationReason}
            disabled={disabled}
            onValueChange={value =>
              onValueChange("delayedNotificationReason", value)}
          />
          <Field
            name="delayedNotificationEvidence"
            type="textarea"
            rows={3}
            maxLength={5000}
            label={t("form.fields.delayedNotificationEvidence.label")}
            help={t("form.fields.delayedNotificationEvidence.help")}
            value={values.delayedNotificationEvidence}
            error={errors.delayedNotificationEvidence}
            disabled={disabled}
            onValueChange={value =>
              onValueChange("delayedNotificationEvidence", value)}
          />
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <Field
            name="dataSubjectsNotifiedAt"
            label={t("form.fields.dataSubjectsNotifiedAt.label")}
            error={errors.dataSubjectsNotifiedAt}
          >
            <Input
              name="dataSubjectsNotifiedAt"
              type="datetime-local"
              value={values.dataSubjectsNotifiedAt}
              disabled={disabled}
              onValueChange={value =>
                onValueChange("dataSubjectsNotifiedAt", value)}
            />
          </Field>
          <Field
            name="dataSubjectsNotificationEvidence"
            type="textarea"
            rows={3}
            maxLength={5000}
            label={t("form.fields.dataSubjectsNotificationEvidence.label")}
            help={t("form.fields.dataSubjectsNotificationEvidence.help")}
            value={values.dataSubjectsNotificationEvidence}
            error={errors.dataSubjectsNotificationEvidence}
            disabled={disabled}
            onValueChange={value =>
              onValueChange("dataSubjectsNotificationEvidence", value)}
          />
        </div>
      </Card>
    </>
  );
}

interface SectionHeaderProps {
  description: string;
  title: string;
}

function SectionHeader({ description, title }: SectionHeaderProps) {
  return (
    <div className="space-y-1">
      <h2 className="text-base font-medium">{title}</h2>
      <p className="text-sm text-txt-tertiary">{description}</p>
    </div>
  );
}
