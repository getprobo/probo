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

import type { TFunction } from "i18next";
import { z } from "zod";

export type MalaysiaPDPABreachNotificationDecision
  = | "PENDING"
    | "NOT_REQUIRED"
    | "COMMISSIONER_ONLY"
    | "COMMISSIONER_AND_DATA_SUBJECTS";

export interface BreachFormValues {
  title: string;
  description: string;
  occurredAt: string;
  discoveredAt: string;
  awarenessAt: string;
  affectedDataSubjects: string;
  affectedDataRecords: string;
  personalDataTypes: string;
  affectedSystem: string;
  likelyConsequences: string;
  containmentActions: string;
  potentialPhysicalHarm: boolean;
  potentialFinancialLoss: boolean;
  potentialCreditOrPropertyDamage: boolean;
  potentialIllegalUse: boolean;
  sensitivePersonalData: boolean;
  potentialIdentityFraud: boolean;
  notificationDecision: MalaysiaPDPABreachNotificationDecision;
  decisionRationale: string;
  decisionEvidence: string;
  commissionerNotifiedAt: string;
  commissionerNotificationReference: string;
  commissionerConfirmationReceivedAt: string;
  commissionerConfirmationReference: string;
  delayedNotificationReason: string;
  delayedNotificationEvidence: string;
  dataSubjectsNotifiedAt: string;
  dataSubjectsNotificationEvidence: string;
}

export type BreachFormField = keyof BreachFormValues;
export type BreachFormErrors = Partial<Record<BreachFormField, string>>;
export type ParsedBreachFormValues = Omit<
  BreachFormValues,
  "affectedDataSubjects" | "affectedDataRecords"
> & {
  affectedDataSubjects: number;
  affectedDataRecords: number;
};

const TITLE_MAX_LENGTH = 1000;
const CONTENT_MAX_LENGTH = 5000;
const COMMISSIONER_NOTIFICATION_WINDOW_MS = 72 * 60 * 60 * 1000;

export function createDefaultBreachFormValues(): BreachFormValues {
  const now = toDateTimeLocal(new Date().toISOString());

  return {
    title: "",
    description: "",
    occurredAt: "",
    discoveredAt: now,
    awarenessAt: now,
    affectedDataSubjects: "0",
    affectedDataRecords: "0",
    personalDataTypes: "",
    affectedSystem: "",
    likelyConsequences: "",
    containmentActions: "",
    potentialPhysicalHarm: false,
    potentialFinancialLoss: false,
    potentialCreditOrPropertyDamage: false,
    potentialIllegalUse: false,
    sensitivePersonalData: false,
    potentialIdentityFraud: false,
    notificationDecision: "PENDING",
    decisionRationale: "",
    decisionEvidence: "",
    commissionerNotifiedAt: "",
    commissionerNotificationReference: "",
    commissionerConfirmationReceivedAt: "",
    commissionerConfirmationReference: "",
    delayedNotificationReason: "",
    delayedNotificationEvidence: "",
    dataSubjectsNotifiedAt: "",
    dataSubjectsNotificationEvidence: "",
  };
}

export function createBreachFormSchema(t: TFunction) {
  const count = (requiredKey: string) =>
    z
      .string()
      .min(1, t(requiredKey))
      .regex(/^\d+$/, t("form.validation.nonNegativeInteger"))
      .transform(Number)
      .refine(Number.isSafeInteger, t("form.validation.numberTooLarge"));

  const content = z.string().trim().max(
    CONTENT_MAX_LENGTH,
    t("form.validation.contentTooLong"),
  );

  return z
    .object({
      title: z
        .string()
        .trim()
        .min(1, t("form.validation.titleRequired"))
        .max(TITLE_MAX_LENGTH, t("form.validation.titleTooLong")),
      description: content,
      occurredAt: z.string(),
      discoveredAt: z.string().min(1, t("form.validation.discoveredAtRequired")),
      awarenessAt: z.string().min(1, t("form.validation.awarenessAtRequired")),
      affectedDataSubjects: count("form.validation.affectedSubjectsRequired"),
      affectedDataRecords: count("form.validation.affectedRecordsRequired"),
      personalDataTypes: z
        .string()
        .trim()
        .min(1, t("form.validation.personalDataTypesRequired"))
        .max(CONTENT_MAX_LENGTH, t("form.validation.contentTooLong")),
      affectedSystem: content,
      likelyConsequences: content,
      containmentActions: content,
      potentialPhysicalHarm: z.boolean(),
      potentialFinancialLoss: z.boolean(),
      potentialCreditOrPropertyDamage: z.boolean(),
      potentialIllegalUse: z.boolean(),
      sensitivePersonalData: z.boolean(),
      potentialIdentityFraud: z.boolean(),
      notificationDecision: z.enum([
        "PENDING",
        "NOT_REQUIRED",
        "COMMISSIONER_ONLY",
        "COMMISSIONER_AND_DATA_SUBJECTS",
      ]),
      decisionRationale: content,
      decisionEvidence: content,
      commissionerNotifiedAt: z.string(),
      commissionerNotificationReference: z
        .string()
        .trim()
        .max(TITLE_MAX_LENGTH, t("form.validation.referenceTooLong")),
      commissionerConfirmationReceivedAt: z.string(),
      commissionerConfirmationReference: z
        .string()
        .trim()
        .max(TITLE_MAX_LENGTH, t("form.validation.referenceTooLong")),
      delayedNotificationReason: content,
      delayedNotificationEvidence: content,
      dataSubjectsNotifiedAt: z.string(),
      dataSubjectsNotificationEvidence: content,
    })
    .superRefine((value, context) => {
      const occurredAt = toDate(value.occurredAt);
      const discoveredAt = toDate(value.discoveredAt);
      const awarenessAt = toDate(value.awarenessAt);
      const commissionerNotifiedAt = toDate(value.commissionerNotifiedAt);
      const commissionerConfirmationAt = toDate(
        value.commissionerConfirmationReceivedAt,
      );
      const dataSubjectsNotifiedAt = toDate(value.dataSubjectsNotifiedAt);

      if (occurredAt && discoveredAt && occurredAt > discoveredAt) {
        addIssue(
          context,
          "occurredAt",
          t("form.validation.occurredAfterDiscovery"),
        );
      }
      if (discoveredAt && awarenessAt && discoveredAt > awarenessAt) {
        addIssue(
          context,
          "awarenessAt",
          t("form.validation.awarenessBeforeDiscovery"),
        );
      }
      if (value.notificationDecision !== "PENDING" && !value.decisionRationale) {
        addIssue(
          context,
          "decisionRationale",
          t("form.validation.decisionRationaleRequired"),
        );
      }

      validatePair(
        context,
        value.commissionerNotifiedAt,
        "commissionerNotifiedAt",
        value.commissionerNotificationReference,
        "commissionerNotificationReference",
        t,
      );
      validatePair(
        context,
        value.commissionerConfirmationReceivedAt,
        "commissionerConfirmationReceivedAt",
        value.commissionerConfirmationReference,
        "commissionerConfirmationReference",
        t,
      );
      validatePair(
        context,
        value.dataSubjectsNotifiedAt,
        "dataSubjectsNotifiedAt",
        value.dataSubjectsNotificationEvidence,
        "dataSubjectsNotificationEvidence",
        t,
      );

      if (commissionerNotifiedAt && awarenessAt) {
        if (commissionerNotifiedAt < awarenessAt) {
          addIssue(
            context,
            "commissionerNotifiedAt",
            t("form.validation.commissionerBeforeAwareness"),
          );
        }
        if (
          commissionerNotifiedAt.getTime()
          > awarenessAt.getTime() + COMMISSIONER_NOTIFICATION_WINDOW_MS
        ) {
          if (!value.delayedNotificationReason) {
            addIssue(
              context,
              "delayedNotificationReason",
              t("form.validation.delayReasonRequired"),
            );
          }
          if (!value.delayedNotificationEvidence) {
            addIssue(
              context,
              "delayedNotificationEvidence",
              t("form.validation.delayEvidenceRequired"),
            );
          }
        }
      }

      if (commissionerConfirmationAt) {
        if (!commissionerNotifiedAt) {
          addIssue(
            context,
            "commissionerConfirmationReceivedAt",
            t("form.validation.confirmationRequiresNotification"),
          );
        } else if (commissionerConfirmationAt < commissionerNotifiedAt) {
          addIssue(
            context,
            "commissionerConfirmationReceivedAt",
            t("form.validation.confirmationBeforeNotification"),
          );
        }
      }

      if (dataSubjectsNotifiedAt) {
        if (!commissionerNotifiedAt) {
          addIssue(
            context,
            "dataSubjectsNotifiedAt",
            t("form.validation.dataSubjectsRequireCommissioner"),
          );
        } else if (dataSubjectsNotifiedAt < commissionerNotifiedAt) {
          addIssue(
            context,
            "dataSubjectsNotifiedAt",
            t("form.validation.dataSubjectsBeforeCommissioner"),
          );
        }
      }
    });
}

export function toBreachFormErrors(error: z.ZodError): BreachFormErrors {
  const errors: BreachFormErrors = {};

  for (const issue of error.issues) {
    const [field] = issue.path;
    if (typeof field === "string" && !(field in errors)) {
      errors[field as BreachFormField] = issue.message;
    }
  }

  return errors;
}

export function toDateTimeLocal(value: string | null | undefined): string {
  if (!value) {
    return "";
  }

  const date = new Date(value);
  const localDate = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return localDate.toISOString().slice(0, 16);
}

export function toISOStringOrNull(value: string): string | null {
  return value ? new Date(value).toISOString() : null;
}

export function toNullable(value: string): string | null {
  const normalized = value.trim();
  return normalized ? normalized : null;
}

export function breachValuesToMutationFields(
  input: ParsedBreachFormValues,
) {
  return {
    title: input.title,
    description: toNullable(input.description),
    occurredAt: toISOStringOrNull(input.occurredAt),
    discoveredAt: new Date(input.discoveredAt).toISOString(),
    awarenessAt: new Date(input.awarenessAt).toISOString(),
    affectedDataSubjects: input.affectedDataSubjects,
    affectedDataRecords: input.affectedDataRecords,
    personalDataTypes: input.personalDataTypes,
    affectedSystem: toNullable(input.affectedSystem),
    likelyConsequences: toNullable(input.likelyConsequences),
    containmentActions: toNullable(input.containmentActions),
    potentialPhysicalHarm: input.potentialPhysicalHarm,
    potentialFinancialLoss: input.potentialFinancialLoss,
    potentialCreditOrPropertyDamage: input.potentialCreditOrPropertyDamage,
    potentialIllegalUse: input.potentialIllegalUse,
    sensitivePersonalData: input.sensitivePersonalData,
    potentialIdentityFraud: input.potentialIdentityFraud,
    notificationDecision: input.notificationDecision,
    decisionRationale: toNullable(input.decisionRationale),
    decisionEvidence: toNullable(input.decisionEvidence),
    commissionerNotifiedAt: toISOStringOrNull(input.commissionerNotifiedAt),
    commissionerNotificationReference: toNullable(
      input.commissionerNotificationReference,
    ),
    commissionerConfirmationReceivedAt: toISOStringOrNull(
      input.commissionerConfirmationReceivedAt,
    ),
    commissionerConfirmationReference: toNullable(
      input.commissionerConfirmationReference,
    ),
    delayedNotificationReason: toNullable(input.delayedNotificationReason),
    delayedNotificationEvidence: toNullable(input.delayedNotificationEvidence),
    dataSubjectsNotifiedAt: toISOStringOrNull(input.dataSubjectsNotifiedAt),
    dataSubjectsNotificationEvidence: toNullable(
      input.dataSubjectsNotificationEvidence,
    ),
  };
}

function toDate(value: string): Date | null {
  return value ? new Date(value) : null;
}

function addIssue(
  context: z.RefinementCtx<unknown>,
  field: BreachFormField,
  message: string,
) {
  context.addIssue({ code: "custom", path: [field], message });
}

function validatePair(
  context: z.RefinementCtx<unknown>,
  firstValue: string,
  firstField: BreachFormField,
  secondValue: string,
  secondField: BreachFormField,
  t: TFunction,
) {
  if (firstValue && !secondValue) {
    addIssue(context, secondField, t("form.validation.pairedFieldRequired"));
  }
  if (secondValue && !firstValue) {
    addIssue(context, firstField, t("form.validation.pairedFieldRequired"));
  }
}
