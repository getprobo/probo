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

import { IconCheckmark1 } from "@probo/ui";
import { useTranslation } from "react-i18next";

import {
  INTEGRATION_STEP_IDS,
  ONBOARDING_STEP_IDS,
  type OnboardingCompletion,
  type OnboardingStepId,
} from "../_lib/onboardingSteps";
import type { IntegrationStepId } from "../_lib/onboardingSteps";

type OnboardingStepperProps = {
  activeStep: OnboardingStepId;
  completion: OnboardingCompletion;
  deferredSteps: ReadonlySet<IntegrationStepId>;
  onStepClick: (stepId: OnboardingStepId) => void;
};

function stepLabelKey(stepId: OnboardingStepId): string {
  if (stepId === "congrats") return "stepper.congrats";
  return `stepper.${stepId}`;
}

export function OnboardingStepper(props: OnboardingStepperProps) {
  const { activeStep, completion, deferredSteps, onStepClick } = props;
  const { t } = useTranslation("organizations/onboarding");

  const steps = ONBOARDING_STEP_IDS;

  return (
    <nav aria-label={t("stepper.ariaLabel")} className="mb-8">
      <ol className="flex flex-wrap gap-2">
        {steps.map((stepId, index) => {
          const isActive = stepId === activeStep;
          const isIntegration = (INTEGRATION_STEP_IDS as readonly string[]).includes(
            stepId,
          );
          const complete = isIntegration
            && completion[stepId as IntegrationStepId];
          const deferred = isIntegration
            && deferredSteps.has(stepId as IntegrationStepId)
            && !complete;

          const clickable = stepId === "congrats"
            ? activeStep !== "congrats"
            : complete || deferred || steps.indexOf(activeStep) >= index;

          return (
            <li key={stepId}>
              <button
                type="button"
                disabled={!clickable && !isActive}
                onClick={() => clickable && onStepClick(stepId)}
                className={[
                  "flex items-center gap-2 rounded-3 px-3 py-2 text-2 transition-colors",
                  isActive
                    ? "bg-accent-3 text-accent-11 font-medium"
                    : clickable
                      ? "text-txt-secondary hover:bg-level-2"
                      : "text-txt-tertiary cursor-default",
                ].join(" ")}
              >
                <span
                  className={[
                    "flex size-6 shrink-0 items-center justify-center rounded-full text-1 font-medium",
                    complete
                      ? "bg-success-9 text-white"
                      : isActive
                        ? "bg-accent-9 text-white"
                        : "bg-level-3 text-txt-secondary",
                  ].join(" ")}
                >
                  {complete
                    ? <IconCheckmark1 size={14} />
                    : index + 1}
                </span>
                <span>{t(stepLabelKey(stepId))}</span>
                {deferred && (
                  <span className="text-1 text-txt-tertiary">
                    ({t("status.later")})
                  </span>
                )}
              </button>
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
