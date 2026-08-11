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

import { dateTimeFormat } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  Checkbox,
  Field,
  Input,
  Label,
  Option,
  Select,
} from "@probo/ui";
import type { TFunction } from "i18next";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql, type RecordSourceSelectorProxy } from "relay-runtime";
import { z } from "zod";

import type { MalaysiaPDPAProfileForm_organization$key } from "#/__generated__/core/MalaysiaPDPAProfileForm_organization.graphql";
import type { MalaysiaPDPAProfileForm_profile$key } from "#/__generated__/core/MalaysiaPDPAProfileForm_profile.graphql";
import type { MalaysiaPDPAProfileFormUpdateMutation } from "#/__generated__/core/MalaysiaPDPAProfileFormUpdateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const NO_DPO_PROFILE = "__none__";
const MAX_REFERENCE_LENGTH = 1000;

const profileFragment = graphql`
  fragment MalaysiaPDPAProfileForm_profile on MalaysiaPDPAProfile {
    totalDataSubjects
    sensitiveDataSubjects
    regularSystematicMonitoring
    dpoRequired
    dpoRequirementReasons
    assessedByProfileId
    assessedAt
    dpoProfileId
    dpoAppointedAt
    commissionerNotificationDueAt
    commissionerNotificationOverdue
    commissionerNotifiedAt
    commissionerNotificationReference
  }
`;

const organizationFragment = graphql`
  fragment MalaysiaPDPAProfileForm_organization on Organization {
    id
    canUpdateMalaysiaPDPAProfile: permission(
      action: "core:malaysia-pdpa-profile:update"
    )
    malaysiaPDPAProfile {
      ...MalaysiaPDPAProfileForm_profile
    }
    profiles(
      first: 1000
      orderBy: { direction: ASC, field: FULL_NAME }
    ) {
      edges {
        node {
          id
          fullName
          state
        }
      }
    }
  }
`;

const updateMutation = graphql`
  mutation MalaysiaPDPAProfileFormUpdateMutation(
    $input: UpdateMalaysiaPDPAProfileInput!
  ) {
    updateMalaysiaPDPAProfile(input: $input) {
      malaysiaPDPAProfile {
        ...MalaysiaPDPAProfileForm_profile
      }
    }
  }
`;

interface FormValues {
  totalDataSubjects: string;
  sensitiveDataSubjects: string;
  regularSystematicMonitoring: boolean;
  dpoProfileId: string | null;
  dpoAppointedAt: string;
  commissionerNotifiedAt: string;
  commissionerNotificationReference: string;
}

type FormField = keyof FormValues;
type FormErrors = Partial<Record<FormField, string>>;

interface MalaysiaPDPAProfileFormProps {
  organizationKey: MalaysiaPDPAProfileForm_organization$key;
}

export function MalaysiaPDPAProfileForm({
  organizationKey,
}: MalaysiaPDPAProfileFormProps) {
  const { i18n, t } = useTranslation("organizations/settings");
  const organization = useFragment(organizationFragment, organizationKey);
  const profile = useFragment(
    profileFragment,
    organization.malaysiaPDPAProfile as MalaysiaPDPAProfileForm_profile$key,
  );
  const [values, setValues] = useState<FormValues>(() => ({
    totalDataSubjects: String(profile.totalDataSubjects),
    sensitiveDataSubjects: String(profile.sensitiveDataSubjects),
    regularSystematicMonitoring: profile.regularSystematicMonitoring,
    dpoProfileId: profile.dpoProfileId ?? null,
    dpoAppointedAt: toDateTimeLocal(profile.dpoAppointedAt),
    commissionerNotifiedAt: toDateTimeLocal(profile.commissionerNotifiedAt),
    commissionerNotificationReference:
      profile.commissionerNotificationReference ?? "",
  }));
  const [errors, setErrors] = useState<FormErrors>({});
  const [updateMalaysiaPDPAProfile, isUpdating]
    = useMutation<MalaysiaPDPAProfileFormUpdateMutation>(updateMutation, {
      successMessage: t("malaysiaPDPA.messages.updated"),
      errorToast: t("malaysiaPDPA.errors.update"),
    });

  const profiles = organization.profiles.edges.map(({ node }) => node);
  const selectableProfiles = profiles.filter(
    candidate =>
      candidate.state === "ACTIVE" || candidate.id === profile.dpoProfileId,
  );
  const assessedBy = profiles.find(
    candidate => candidate.id === profile.assessedByProfileId,
  );

  function updateValue<Field extends FormField>(
    field: Field,
    value: FormValues[Field],
  ) {
    setValues(current => ({ ...current, [field]: value }));
    setErrors(current => ({ ...current, [field]: undefined }));
  }

  function onDPOProfileChange(profileId: string) {
    if (profileId === NO_DPO_PROFILE) {
      setValues(current => ({
        ...current,
        dpoProfileId: null,
        dpoAppointedAt: "",
        commissionerNotifiedAt: "",
        commissionerNotificationReference: "",
      }));
      setErrors(current => ({
        ...current,
        dpoProfileId: undefined,
        dpoAppointedAt: undefined,
        commissionerNotifiedAt: undefined,
        commissionerNotificationReference: undefined,
      }));
      return;
    }

    updateValue("dpoProfileId", profileId);
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const result = createFormSchema(t).safeParse(values);
    if (!result.success) {
      setErrors(toFormErrors(result.error));
      return;
    }

    setErrors({});
    const input = result.data;

    try {
      await updateMalaysiaPDPAProfile({
        variables: {
          input: {
            organizationId: organization.id,
            totalDataSubjects: input.totalDataSubjects,
            sensitiveDataSubjects: input.sensitiveDataSubjects,
            regularSystematicMonitoring:
              input.regularSystematicMonitoring,
            dpoProfileId: input.dpoProfileId,
            dpoAppointedAt: toISOString(input.dpoAppointedAt),
            commissionerNotifiedAt: toISOString(
              input.commissionerNotifiedAt,
            ),
            commissionerNotificationReference:
              input.commissionerNotificationReference || null,
          },
        },
        updater: store => updateProfileLink(store, organization.id),
      });
    } catch {
      // useMutation already displays the localized server error.
    }
  }

  const assessmentBadge = getAssessmentBadge(profile.assessedAt, profile.dpoRequired);

  return (
    <form onSubmit={event => void onSubmit(event)} className="space-y-6">
      {!organization.canUpdateMalaysiaPDPAProfile && (
        <Card padded className="border-border-mid bg-subtle">
          <p className="text-sm text-txt-secondary">
            {t("malaysiaPDPA.page.readOnly")}
          </p>
        </Card>
      )}

      <Card padded className="space-y-4">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <h3 className="text-sm font-semibold">
              {t("malaysiaPDPA.status.title")}
            </h3>
            <p className="text-xs text-txt-tertiary">
              {t("malaysiaPDPA.status.description")}
            </p>
          </div>
          <Badge variant={assessmentBadge.variant}>
            {t(assessmentBadge.translationKey)}
          </Badge>
        </div>

        {profile.dpoRequirementReasons.length > 0 && (
          <div className="space-y-2">
            <p className="text-sm font-medium">
              {t("malaysiaPDPA.status.reasonsTitle")}
            </p>
            <ul className="list-disc pl-5 space-y-1 text-sm text-txt-secondary">
              {profile.dpoRequirementReasons.map(reason => (
                <li key={reason}>{t(reasonTranslationKey(reason))}</li>
              ))}
            </ul>
          </div>
        )}

        <p className="text-xs text-txt-tertiary">
          {t("malaysiaPDPA.status.thresholdNote")}
        </p>

        {profile.assessedAt && (
          <div className="flex flex-wrap gap-x-6 gap-y-1 text-xs text-txt-tertiary">
            <span>
              {t("malaysiaPDPA.status.assessedAt", {
                date: dateTimeFormat(i18n.language, profile.assessedAt),
              })}
            </span>
            <span>
              {t("malaysiaPDPA.status.assessedBy", {
                name:
                  assessedBy?.fullName
                  ?? t("malaysiaPDPA.status.unknownAssessor"),
              })}
            </span>
          </div>
        )}
      </Card>

      <Card padded className="space-y-5">
        <div className="space-y-1">
          <h3 className="text-sm font-semibold">
            {t("malaysiaPDPA.assessment.title")}
          </h3>
          <p className="text-xs text-txt-tertiary">
            {t("malaysiaPDPA.assessment.description")}
          </p>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <Field
            name="totalDataSubjects"
            type="number"
            min={0}
            step={1}
            label={t("malaysiaPDPA.assessment.totalDataSubjects.label")}
            help={t("malaysiaPDPA.assessment.totalDataSubjects.help")}
            value={values.totalDataSubjects}
            error={errors.totalDataSubjects}
            disabled={!organization.canUpdateMalaysiaPDPAProfile || isUpdating}
            onValueChange={value =>
              updateValue("totalDataSubjects", value)}
          />
          <Field
            name="sensitiveDataSubjects"
            type="number"
            min={0}
            step={1}
            label={t("malaysiaPDPA.assessment.sensitiveDataSubjects.label")}
            help={t("malaysiaPDPA.assessment.sensitiveDataSubjects.help")}
            value={values.sensitiveDataSubjects}
            error={errors.sensitiveDataSubjects}
            disabled={!organization.canUpdateMalaysiaPDPAProfile || isUpdating}
            onValueChange={value =>
              updateValue("sensitiveDataSubjects", value)}
          />
        </div>

        <Label className="flex items-start gap-3 mb-0 cursor-pointer">
          <Checkbox
            checked={values.regularSystematicMonitoring}
            disabled={!organization.canUpdateMalaysiaPDPAProfile || isUpdating}
            onChange={checked =>
              updateValue("regularSystematicMonitoring", checked)}
          />
          <span className="space-y-1">
            <span className="block text-sm font-medium">
              {t("malaysiaPDPA.assessment.regularMonitoring.label")}
            </span>
            <span className="block text-xs font-normal text-txt-tertiary">
              {t("malaysiaPDPA.assessment.regularMonitoring.help")}
            </span>
          </span>
        </Label>
      </Card>

      <Card padded className="space-y-5">
        <div className="space-y-1">
          <h3 className="text-sm font-semibold">
            {t("malaysiaPDPA.appointment.title")}
          </h3>
          <p className="text-xs text-txt-tertiary">
            {t("malaysiaPDPA.appointment.description")}
          </p>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <Field
            name="dpoProfileId"
            label={t("malaysiaPDPA.appointment.dpo.label")}
            help={t("malaysiaPDPA.appointment.dpo.help")}
            error={errors.dpoProfileId}
          >
            <Select<string>
              value={values.dpoProfileId ?? NO_DPO_PROFILE}
              disabled={!organization.canUpdateMalaysiaPDPAProfile || isUpdating}
              onValueChange={onDPOProfileChange}
            >
              <Option value={NO_DPO_PROFILE}>
                {t("malaysiaPDPA.appointment.dpo.none")}
              </Option>
              {selectableProfiles.map(candidate => (
                <Option key={candidate.id} value={candidate.id}>
                  {candidate.fullName}
                  {candidate.state !== "ACTIVE"
                    ? ` (${t(`malaysiaPDPA.profileStates.${candidate.state}`)})`
                    : ""}
                </Option>
              ))}
            </Select>
          </Field>

          <Field
            name="dpoAppointedAt"
            label={t("malaysiaPDPA.appointment.appointedAt.label")}
            help={t("malaysiaPDPA.appointment.appointedAt.help")}
            error={errors.dpoAppointedAt}
          >
            <Input
              name="dpoAppointedAt"
              type="datetime-local"
              value={values.dpoAppointedAt}
              disabled={
                !organization.canUpdateMalaysiaPDPAProfile
                || isUpdating
                || !values.dpoProfileId
              }
              onValueChange={value => updateValue("dpoAppointedAt", value)}
            />
          </Field>
        </div>
      </Card>

      <Card padded className="space-y-5">
        <div className="space-y-1">
          <h3 className="text-sm font-semibold">
            {t("malaysiaPDPA.notification.title")}
          </h3>
          <p className="text-xs text-txt-tertiary">
            {t("malaysiaPDPA.notification.description")}
          </p>
        </div>

        {profile.commissionerNotificationDueAt && (
          <div className="flex items-center justify-between gap-4 rounded-xl bg-subtle p-3">
            <span className="text-sm text-txt-secondary">
              {t("malaysiaPDPA.notification.dueAt.label")}
            </span>
            <div className="flex items-center gap-2">
              {profile.commissionerNotificationOverdue && (
                <Badge variant="danger">
                  {t("malaysiaPDPA.notification.overdue")}
                </Badge>
              )}
              <span className="text-sm font-medium">
                {dateTimeFormat(
                  i18n.language,
                  profile.commissionerNotificationDueAt,
                )}
              </span>
            </div>
          </div>
        )}

        <div className="grid gap-4 md:grid-cols-2">
          <Field
            name="commissionerNotifiedAt"
            label={t("malaysiaPDPA.notification.notifiedAt.label")}
            help={t("malaysiaPDPA.notification.notifiedAt.help")}
            error={errors.commissionerNotifiedAt}
          >
            <Input
              name="commissionerNotifiedAt"
              type="datetime-local"
              value={values.commissionerNotifiedAt}
              disabled={
                !organization.canUpdateMalaysiaPDPAProfile
                || isUpdating
                || !values.dpoProfileId
              }
              onValueChange={value =>
                updateValue("commissionerNotifiedAt", value)}
            />
          </Field>
          <Field
            name="commissionerNotificationReference"
            type="text"
            maxLength={MAX_REFERENCE_LENGTH}
            label={t("malaysiaPDPA.notification.reference.label")}
            help={t("malaysiaPDPA.notification.reference.help")}
            value={values.commissionerNotificationReference}
            error={errors.commissionerNotificationReference}
            disabled={
              !organization.canUpdateMalaysiaPDPAProfile
              || isUpdating
              || !values.commissionerNotifiedAt
            }
            onValueChange={value =>
              updateValue("commissionerNotificationReference", value)}
          />
        </div>
      </Card>

      {organization.canUpdateMalaysiaPDPAProfile && (
        <div className="flex justify-end">
          <Button type="submit" disabled={isUpdating}>
            {isUpdating
              ? t("malaysiaPDPA.actions.saving")
              : t("malaysiaPDPA.actions.save")}
          </Button>
        </div>
      )}
    </form>
  );
}

function createFormSchema(t: TFunction) {
  const count = (requiredMessage: string) =>
    z
      .string()
      .min(1, requiredMessage)
      .regex(/^\d+$/, t("malaysiaPDPA.validation.nonNegativeInteger"))
      .transform(Number)
      .refine(Number.isSafeInteger, t("malaysiaPDPA.validation.numberTooLarge"));

  return z
    .object({
      totalDataSubjects: count(
        t("malaysiaPDPA.validation.totalDataSubjectsRequired"),
      ),
      sensitiveDataSubjects: count(
        t("malaysiaPDPA.validation.sensitiveDataSubjectsRequired"),
      ),
      regularSystematicMonitoring: z.boolean(),
      dpoProfileId: z.string().nullable(),
      dpoAppointedAt: z.string(),
      commissionerNotifiedAt: z.string(),
      commissionerNotificationReference: z
        .string()
        .trim()
        .max(
          MAX_REFERENCE_LENGTH,
          t("malaysiaPDPA.validation.referenceTooLong"),
        ),
    })
    .superRefine((value, context) => {
      if (value.sensitiveDataSubjects > value.totalDataSubjects) {
        context.addIssue({
          code: "custom",
          path: ["sensitiveDataSubjects"],
          message: t("malaysiaPDPA.validation.sensitiveExceedsTotal"),
        });
      }

      const hasDPO = value.dpoProfileId !== null;
      const hasAppointment = value.dpoAppointedAt !== "";
      if (hasDPO !== hasAppointment) {
        context.addIssue({
          code: "custom",
          path: [hasDPO ? "dpoAppointedAt" : "dpoProfileId"],
          message: t("malaysiaPDPA.validation.dpoAndAppointmentRequired"),
        });
      }

      if (value.commissionerNotifiedAt && !hasAppointment) {
        context.addIssue({
          code: "custom",
          path: ["commissionerNotifiedAt"],
          message: t("malaysiaPDPA.validation.notificationRequiresDPO"),
        });
      }

      if (
        value.commissionerNotifiedAt
        && value.dpoAppointedAt
        && new Date(value.commissionerNotifiedAt)
        < new Date(value.dpoAppointedAt)
      ) {
        context.addIssue({
          code: "custom",
          path: ["commissionerNotifiedAt"],
          message: t("malaysiaPDPA.validation.notificationBeforeAppointment"),
        });
      }

      if (
        value.commissionerNotificationReference
        && !value.commissionerNotifiedAt
      ) {
        context.addIssue({
          code: "custom",
          path: ["commissionerNotificationReference"],
          message: t("malaysiaPDPA.validation.referenceRequiresNotification"),
        });
      }
    });
}

function toFormErrors(error: z.ZodError): FormErrors {
  const errors: FormErrors = {};

  for (const issue of error.issues) {
    const [field] = issue.path;
    if (typeof field === "string" && !(field in errors)) {
      errors[field as FormField] = issue.message;
    }
  }

  return errors;
}

function toDateTimeLocal(value: string | null | undefined): string {
  if (!value) {
    return "";
  }

  const date = new Date(value);
  const localDate = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return localDate.toISOString().slice(0, 16);
}

function toISOString(value: string): string | null {
  return value ? new Date(value).toISOString() : null;
}

function updateProfileLink(
  store: RecordSourceSelectorProxy,
  organizationId: string,
) {
  const payload = store.getRootField("updateMalaysiaPDPAProfile");
  const updatedProfile = payload?.getLinkedRecord("malaysiaPDPAProfile");
  const organization = store.get(organizationId);

  if (updatedProfile && organization) {
    organization.setLinkedRecord(updatedProfile, "malaysiaPDPAProfile");
  }
}

function getAssessmentBadge(
  assessedAt: string | null | undefined,
  dpoRequired: boolean,
) {
  if (!assessedAt) {
    return {
      variant: "neutral" as const,
      translationKey: "malaysiaPDPA.status.notAssessed",
    };
  }

  if (dpoRequired) {
    return {
      variant: "warning" as const,
      translationKey: "malaysiaPDPA.status.required",
    };
  }

  return {
    variant: "success" as const,
    translationKey: "malaysiaPDPA.status.notRequired",
  };
}

function reasonTranslationKey(reason: string): string {
  switch (reason) {
    case "PERSONAL_DATA_VOLUME":
      return "malaysiaPDPA.status.reasons.personalDataVolume";
    case "SENSITIVE_OR_FINANCIAL_DATA_VOLUME":
      return "malaysiaPDPA.status.reasons.sensitiveDataVolume";
    case "REGULAR_AND_SYSTEMATIC_MONITORING":
      return "malaysiaPDPA.status.reasons.regularMonitoring";
    default:
      return "malaysiaPDPA.status.reasons.unknown";
  }
}
