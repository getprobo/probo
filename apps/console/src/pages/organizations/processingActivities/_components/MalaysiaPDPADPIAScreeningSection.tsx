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

import { formatError, type GraphQLError } from "@probo/helpers";
import {
  Badge,
  Button,
  Checkbox,
  Field,
  Label,
  Textarea,
  useToast,
} from "@probo/ui";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { MalaysiaPDPADPIAScreeningSection_activity$key } from "#/__generated__/core/MalaysiaPDPADPIAScreeningSection_activity.graphql";
import { useUpdateProcessingActivity } from "#/hooks/graph/ProcessingActivityGraph";

const fragment = graphql`
  fragment MalaysiaPDPADPIAScreeningSection_activity on ProcessingActivity {
    id
    malaysiaPDPADPIATotalDataSubjects
    malaysiaPDPADPIASensitiveDataSubjects
    malaysiaPDPADPIALegalOrSignificantEffects
    malaysiaPDPADPIASystematicMonitoring
    malaysiaPDPADPIAInnovativeTechnology
    malaysiaPDPADPIADenialOrRestrictionOfRights
    malaysiaPDPADPIALocationOrBehaviourTracking
    malaysiaPDPADPIAChildrenOrVulnerableDataSubjects
    malaysiaPDPADPIAHighRiskAutomatedDecisionMaking
    malaysiaPDPADPIAOtherHighRiskFactors
    malaysiaPDPADPIARecommendation
    malaysiaPDPADPIAReasons
    malaysiaPDPADPIAAssessedAt
    malaysiaPDPADPIARuleVersion
    malaysiaPDPADPIARuleSource
    canUpdate: permission(action: "core:processing-activity:update")
    canCreateDPIA: permission(
      action: "core:data-protection-impact-assessment:create"
    )
    dataProtectionImpactAssessment { id }
  }
`;

type ScreeningForm = {
  totalDataSubjects: number;
  sensitiveDataSubjects: number;
  legalOrSignificantEffects: boolean;
  systematicMonitoring: boolean;
  innovativeTechnology: boolean;
  denialOrRestrictionOfRights: boolean;
  locationOrBehaviourTracking: boolean;
  childrenOrVulnerableDataSubjects: boolean;
  highRiskAutomatedDecisionMaking: boolean;
  otherHighRiskFactors: string;
};

type Props = {
  activityKey: MalaysiaPDPADPIAScreeningSection_activity$key;
  onOpenDPIA: () => void;
};

const booleanFields = [
  "legalOrSignificantEffects",
  "systematicMonitoring",
  "innovativeTechnology",
  "denialOrRestrictionOfRights",
  "locationOrBehaviourTracking",
  "childrenOrVulnerableDataSubjects",
  "highRiskAutomatedDecisionMaking",
] as const;

export function MalaysiaPDPADPIAScreeningSection({
  activityKey,
  onOpenDPIA,
}: Props) {
  const activity = useFragment(fragment, activityKey);
  const { t, i18n } = useTranslation();
  const { toast } = useToast();
  const updateActivity = useUpdateProcessingActivity();
  const form = useForm<ScreeningForm>({
    defaultValues: {
      totalDataSubjects: activity.malaysiaPDPADPIATotalDataSubjects,
      sensitiveDataSubjects: activity.malaysiaPDPADPIASensitiveDataSubjects,
      legalOrSignificantEffects:
        activity.malaysiaPDPADPIALegalOrSignificantEffects,
      systematicMonitoring: activity.malaysiaPDPADPIASystematicMonitoring,
      innovativeTechnology: activity.malaysiaPDPADPIAInnovativeTechnology,
      denialOrRestrictionOfRights:
        activity.malaysiaPDPADPIADenialOrRestrictionOfRights,
      locationOrBehaviourTracking:
        activity.malaysiaPDPADPIALocationOrBehaviourTracking,
      childrenOrVulnerableDataSubjects:
        activity.malaysiaPDPADPIAChildrenOrVulnerableDataSubjects,
      highRiskAutomatedDecisionMaking:
        activity.malaysiaPDPADPIAHighRiskAutomatedDecisionMaking,
      otherHighRiskFactors:
        activity.malaysiaPDPADPIAOtherHighRiskFactors ?? "",
    },
  });

  const recommendation = activity.malaysiaPDPADPIARecommendation;
  const recommendationVariant = recommendation === "REQUIRED"
    ? "danger"
    : recommendation === "DPO_REVIEW_REQUIRED"
      ? "warning"
      : "neutral";
  const assessedAt = activity.malaysiaPDPADPIAAssessedAt
    ? new Intl.DateTimeFormat(i18n.language, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(new Date(activity.malaysiaPDPADPIAAssessedAt))
    : null;

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      await updateActivity({
        id: activity.id,
        malaysiaPDPADPIAScreening: {
          ...values,
          otherHighRiskFactors: values.otherHighRiskFactors || undefined,
        },
      });
      toast({
        title: t("processingActivityDetailsPage.malaysiaDPIA.savedTitle"),
        description: t("processingActivityDetailsPage.malaysiaDPIA.savedDescription"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: t("processingActivityDetailsPage.messages.error"),
        description: formatError(
          t("processingActivityDetailsPage.malaysiaDPIA.saveError"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  return (
    <section className="mb-8 space-y-6 border-b pb-8">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-xl font-semibold">
            {t("processingActivityDetailsPage.malaysiaDPIA.title")}
          </h2>
          <p className="mt-1 text-sm text-txt-secondary">
            {t("processingActivityDetailsPage.malaysiaDPIA.description")}
          </p>
        </div>
        <Badge variant={recommendationVariant}>
          {t(`processingActivityDetailsPage.malaysiaDPIA.recommendations.${recommendation}`)}
        </Badge>
      </div>

      {activity.malaysiaPDPADPIAReasons.length > 0 && (
        <div className="rounded-md bg-level-1 p-4">
          <p className="mb-2 text-sm font-medium">
            {t("processingActivityDetailsPage.malaysiaDPIA.triggeredCriteria")}
          </p>
          <ul className="list-disc space-y-1 pl-5 text-sm text-txt-secondary">
            {activity.malaysiaPDPADPIAReasons.map(reason => (
              <li key={reason}>
                {t(`processingActivityDetailsPage.malaysiaDPIA.reasons.${reason}`)}
              </li>
            ))}
          </ul>
        </div>
      )}

      {assessedAt && (
        <div className="space-y-1 text-xs text-txt-tertiary">
          <p>
            {t("processingActivityDetailsPage.malaysiaDPIA.assessmentMetadata", {
              date: assessedAt,
              version: activity.malaysiaPDPADPIARuleVersion,
            })}
          </p>
          {activity.malaysiaPDPADPIARuleSource && (
            <a
              className="underline"
              href={activity.malaysiaPDPADPIARuleSource}
              rel="noreferrer"
              target="_blank"
            >
              {t("processingActivityDetailsPage.malaysiaDPIA.ruleSource")}
            </a>
          )}
        </div>
      )}

      <form className="space-y-5" onSubmit={event => void onSubmit(event)}>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <Field
            label={t("processingActivityDetailsPage.malaysiaDPIA.totalDataSubjects")}
            type="number"
            min={0}
            {...form.register("totalDataSubjects", { valueAsNumber: true })}
            disabled={!activity.canUpdate}
            required
          />
          <Field
            label={t("processingActivityDetailsPage.malaysiaDPIA.sensitiveDataSubjects")}
            type="number"
            min={0}
            {...form.register("sensitiveDataSubjects", { valueAsNumber: true })}
            disabled={!activity.canUpdate}
            required
          />
        </div>

        <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
          {booleanFields.map(name => (
            <Controller
              key={name}
              control={form.control}
              name={name}
              render={({ field }) => (
                <Label className="flex items-start gap-2 font-normal">
                  <Checkbox
                    checked={field.value}
                    onChange={field.onChange}
                    disabled={!activity.canUpdate}
                  />
                  <span>
                    {t(`processingActivityDetailsPage.malaysiaDPIA.criteria.${name}`)}
                  </span>
                </Label>
              )}
            />
          ))}
        </div>

        <div>
          <Label htmlFor="malaysia-dpia-other-factors">
            {t("processingActivityDetailsPage.malaysiaDPIA.otherHighRiskFactors")}
          </Label>
          <Textarea
            id="malaysia-dpia-other-factors"
            rows={3}
            {...form.register("otherHighRiskFactors")}
            disabled={!activity.canUpdate}
          />
        </div>

        <p className="text-xs text-txt-tertiary">
          {t("processingActivityDetailsPage.malaysiaDPIA.disclaimer")}
        </p>

        <div className="flex flex-wrap justify-end gap-2">
          {recommendation !== "NOT_INDICATED"
            && !activity.dataProtectionImpactAssessment
            && activity.canCreateDPIA && (
            <Button type="button" variant="secondary" onClick={onOpenDPIA}>
              {t("processingActivityDetailsPage.malaysiaDPIA.openDPIA")}
            </Button>
          )}
          {activity.canUpdate && (
            <Button type="submit" disabled={form.formState.isSubmitting}>
              {form.formState.isSubmitting
                ? t("processingActivityDetailsPage.actions.saving")
                : t("processingActivityDetailsPage.malaysiaDPIA.runAssessment")}
            </Button>
          )}
        </div>
      </form>
    </section>
  );
}
