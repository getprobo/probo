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

import { ButtonLink, Link } from "@probo/ui";
import { useTranslation } from "react-i18next";

import type { OnboardingCompletion } from "../_lib/onboardingSteps";
import { INTEGRATION_STEP_IDS } from "../_lib/onboardingSteps";
import type { IntegrationStepId } from "../_lib/onboardingSteps";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CongratsOnboardingIllustration } from "./illustrations/OnboardingIllustrations";
import { OnboardingStepDoCard } from "./OnboardingStepActions";

type Props = {
  completion: OnboardingCompletion;
  deferredSteps: ReadonlySet<IntegrationStepId>;
};

export function CongratsOnboardingStep(props: Props) {
  const { completion, deferredSteps } = props;
  const { t } = useTranslation("organizations/onboarding");
  const organizationId = useOrganizationId();

  const pendingSteps = INTEGRATION_STEP_IDS.filter(
    stepId => !completion[stepId],
  );

  const settingsPaths: Record<IntegrationStepId, string> = {
    scim: `/organizations/${organizationId}/settings/scim`,
    accessReview: `/organizations/${organizationId}/access-reviews/sources`,
    agent: `/organizations/${organizationId}/devices`,
  };

  return (
    <div className="mx-auto max-w-xl space-y-8 text-center">
      <div
        aria-hidden
        className="mx-auto flex min-h-48 max-w-sm items-center justify-center"
      >
        <CongratsOnboardingIllustration />
      </div>
      <div className="space-y-2">
        <h2 className="text-5 font-semibold text-txt-primary">
          {t("steps.congrats.title")}
        </h2>
        <p className="text-2 text-txt-secondary">{t("steps.congrats.description")}</p>
      </div>

      {pendingSteps.length > 0 && (
        <OnboardingStepDoCard
          title={t("steps.congrats.pendingTitle")}
          actions={<span />}
        >
          <ul className="space-y-2 text-left text-2">
            {pendingSteps.map((stepId) => {
              const deferred = deferredSteps.has(stepId);
              return (
                <li key={stepId} className="flex items-center justify-between gap-2">
                  <span className="text-txt-secondary">
                    {t(`stepper.${stepId}`)}
                    {deferred && (
                      <span className="text-txt-tertiary">
                        {" "}
                        —
                        {" "}
                        {t("status.later")}
                      </span>
                    )}
                  </span>
                  <Link to={settingsPaths[stepId]} size="2">
                    {t("actions.finishLater")}
                  </Link>
                </li>
              );
            })}
          </ul>
        </OnboardingStepDoCard>
      )}

      <ButtonLink to={`/organizations/${organizationId}/tasks`} variant="secondary">
        {t("steps.congrats.goToConsole")}
      </ButtonLink>
      <p className="text-2 text-txt-tertiary">{t("steps.congrats.closeHint")}</p>
    </div>
  );
}
