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

import { usePageTitle } from "@probo/hooks";
import { Role } from "@probo/helpers";
import { PageHeader } from "@probo/ui";
import { useCallback, useEffect, useMemo, use } from "react";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  graphql,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { Navigate, useSearchParams } from "react-router";

import type { onboardingIamQueriesOnboardingIamStatusQuery } from "#/__generated__/iam/onboardingIamQueriesOnboardingIamStatusQuery.graphql";
import type { OnboardingPageQuery } from "#/__generated__/core/OnboardingPageQuery.graphql";
import type { OnboardingPageStatusFragment$key } from "#/__generated__/core/OnboardingPageStatusFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

import { AccessReviewOnboardingDoPanel } from "./_components/AccessReviewOnboardingDoPanel";
import { AgentOnboardingDoPanel } from "./_components/AgentOnboardingDoPanel";
import { CongratsOnboardingStep } from "./_components/CongratsOnboardingStep";
import {
  AccessReviewOnboardingIllustration,
  AgentOnboardingIllustration,
  McpOnboardingIllustration,
  ScimOnboardingIllustration,
} from "./_components/illustrations/OnboardingIllustrations";
import { McpOnboardingDoPanel } from "./_components/McpOnboardingDoPanel";
import { OnboardingLearnPanel } from "./_components/OnboardingLearnPanel";
import { OnboardingStepper } from "./_components/OnboardingStepper";
import { ScimOnboardingDoPanel } from "#/pages/iam/organizations/onboarding/ScimOnboardingDoPanel";
import { OnboardingSessionProvider, useOnboardingSession } from "./_lib/OnboardingSessionContext";
import {
  allIntegrationsComplete,
  firstIncompleteStep,
  isOnboardingStepId,
  ONBOARDING_STEP_IDS,
  type OnboardingCompletion,
  type OnboardingStepId,
} from "./_lib/onboardingSteps";
import type { IntegrationStepId } from "./_lib/onboardingSteps";

export const onboardingPageQuery = graphql`
  query OnboardingPageQuery($organizationId: ID!) {
    accessReviewDrivers {
      ...AddAccessReviewSourceDialogConnectorProviderInfoFragment
    }
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...OnboardingPageStatusFragment
        ...AccessReviewOnboardingDoPanelFragment
        ...AgentOnboardingDoPanelFragment
      }
    }
  }
`;

const onboardingPageStatusFragment = graphql`
  fragment OnboardingPageStatusFragment on Organization {
    accessReviewSources(first: 50) {
      edges {
        node {
          id
        }
      }
    }
    devices(first: 1) {
      edges {
        node {
          id
        }
      }
    }
  }
`;

import { onboardingIamStatusQuery } from "#/pages/iam/organizations/onboarding/onboardingIamQueries";(current: OnboardingStepId): OnboardingStepId {
  const index = ONBOARDING_STEP_IDS.indexOf(current);
  return ONBOARDING_STEP_IDS[Math.min(index + 1, ONBOARDING_STEP_IDS.length - 1)]!;
}

function resolveInitialStep(
  completion: OnboardingCompletion,
  deferred: ReadonlySet<IntegrationStepId>,
  welcome: boolean,
): OnboardingStepId {
  if (welcome) return "scim";
  if (allIntegrationsComplete(completion)) return "congrats";
  return firstIncompleteStep(completion, deferred);
}

function OnboardingPageInner(props: {
  coreQueryRef: PreloadedQuery<OnboardingPageQuery>;
  iamQueryRef: PreloadedQuery<onboardingIamQueriesOnboardingIamStatusQuery>;
  refetchIamStatus: () => void;
}) {
  const { coreQueryRef, iamQueryRef, refetchIamStatus } = props;
  const { t } = useTranslation("organizations/onboarding");
  const role = use(CurrentUser).role;
  const organizationId = useOrganizationId();
  const [searchParams, setSearchParams] = useSearchParams();
  const { clearDeferred, deferStep, deferredSteps } = useOnboardingSession();

  const { accessReviewDrivers, organization } = usePreloadedQuery<OnboardingPageQuery>(
    onboardingPageQuery,
    coreQueryRef,
  );
  const iamData = usePreloadedQuery<onboardingIamQueriesOnboardingIamStatusQuery>(
    onboardingIamStatusQuery,
    iamQueryRef,
  );

  if (organization.__typename !== "Organization") {
    throw new Error("invalid organization node");
  }
  if (iamData.organization.__typename !== "Organization") {
    throw new Error("invalid organization node");
  }

  const orgStatus = useFragment<OnboardingPageStatusFragment$key>(
    onboardingPageStatusFragment,
    organization,
  );

  usePageTitle(t("pageTitle"));

  const completion = useMemo<OnboardingCompletion>(
    () => ({
      scim: !!iamData.organization.scimConfiguration?.id,
      accessReview: orgStatus.accessReviewSources.edges.length > 0,
      agent: orgStatus.devices.edges.length > 0,
    }),
    [
      iamData.organization.scimConfiguration?.id,
      orgStatus.accessReviewSources.edges.length,
      orgStatus.devices.edges.length,
    ],
  );

  useEffect(() => {
    for (const stepId of ["scim", "accessReview", "agent"] as const) {
      if (completion[stepId]) {
        clearDeferred(stepId);
      }
    }
  }, [clearDeferred, completion]);

  const welcome = searchParams.get("welcome") === "1";
  const stepParam = searchParams.get("step");
  const activeStep: OnboardingStepId = isOnboardingStepId(stepParam)
    ? stepParam
    : resolveInitialStep(completion, deferredSteps, welcome);

  const refetchIamScim = useCallback(() => {
    refetchIamStatus();
  }, [refetchIamStatus]);

  const goToStep = useCallback(
    (stepId: OnboardingStepId) => {
      const params = new URLSearchParams(searchParams);
      params.set("step", stepId);
      params.delete("welcome");
      setSearchParams(params);
    },
    [searchParams, setSearchParams],
  );

  const handleContinue = useCallback(() => {
    goToStep(nextStep(activeStep));
  }, [activeStep, goToStep]);

  const handleDefer = useCallback(() => {
    if (activeStep === "scim" || activeStep === "accessReview" || activeStep === "agent") {
      deferStep(activeStep);
    }
    goToStep(nextStep(activeStep));
  }, [activeStep, deferStep, goToStep]);

  if (role === Role.EMPLOYEE) {
    return <Navigate to={`/organizations/${organizationId}/employee`} replace />;
  }
  if (role === Role.AUDITOR) {
    return <Navigate to={`/organizations/${organizationId}/measures`} replace />;
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <PageHeader
        title={t("title")}
        description={t("description")}
      />
      <OnboardingStepper
        activeStep={activeStep}
        completion={completion}
        deferredSteps={deferredSteps}
        onStepClick={goToStep}
      />

      {activeStep === "congrats"
        ? (
            <CongratsOnboardingStep
              completion={completion}
              deferredSteps={deferredSteps}
            />
          )
        : (
            <div className="grid gap-8 lg:grid-cols-2 lg:gap-12">
              <OnboardingLearnPanel
                stepId={activeStep}
                illustration={
                  activeStep === "scim"
                    ? <ScimOnboardingIllustration />
                    : activeStep === "accessReview"
                      ? <AccessReviewOnboardingIllustration />
                      : activeStep === "agent"
                        ? <AgentOnboardingIllustration />
                        : <McpOnboardingIllustration />
                }
              />
              <div>
                {activeStep === "scim" && (
                  <ScimOnboardingDoPanel
                    onContinue={handleContinue}
                    onDefer={handleDefer}
                    onScimUpdated={refetchIamScim}
                  />
                )}
                {activeStep === "accessReview" && (
                  <AccessReviewOnboardingDoPanel
                    accessReviewDrivers={accessReviewDrivers}
                    organizationFKey={organization}
                    onContinue={handleContinue}
                    onDefer={handleDefer}
                  />
                )}
                {activeStep === "agent" && (
                  <AgentOnboardingDoPanel
                    organizationFKey={organization}
                    onContinue={handleContinue}
                    onDefer={handleDefer}
                  />
                )}
                {activeStep === "mcp" && (
                  <McpOnboardingDoPanel
                    onContinue={handleContinue}
                    onDefer={handleDefer}
                  />
                )}
              </div>
            </div>
          )}
    </div>
  );
}

export function OnboardingPage(props: {
  coreQueryRef: PreloadedQuery<OnboardingPageQuery>;
  iamQueryRef: PreloadedQuery<onboardingIamQueriesOnboardingIamStatusQuery>;
  refetchIamStatus: () => void;
}) {
  return (
    <OnboardingSessionProvider>
      <OnboardingPageInner {...props} />
    </OnboardingSessionProvider>
  );
}
