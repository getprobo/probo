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

import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import type { OnboardingStepId } from "../_lib/onboardingSteps";

type OnboardingLearnPanelProps = {
  stepId: Exclude<OnboardingStepId, "congrats">;
  illustration: ReactNode;
};

export function OnboardingLearnPanel(props: OnboardingLearnPanelProps) {
  const { stepId, illustration } = props;
  const { t } = useTranslation("organizations/onboarding");

  const bullets = t(`steps.${stepId}.bullets`, {
    returnObjects: true,
  }) as string[];

  return (
    <div className="flex flex-col gap-6 lg:sticky lg:top-8">
      <div
        aria-hidden
        className="flex min-h-48 items-center justify-center rounded-4 bg-level-1 p-6 motion-safe:animate-in motion-safe:fade-in motion-safe:slide-in-from-bottom-2"
      >
        {illustration}
      </div>
      <div className="space-y-3">
        <h2 className="text-4 font-semibold text-txt-primary">
          {t(`steps.${stepId}.learnTitle`)}
        </h2>
        <p className="text-2 text-txt-secondary leading-relaxed">
          {t(`steps.${stepId}.purpose`)}
        </p>
        <ul className="list-disc space-y-2 pl-5 text-2 text-txt-secondary">
          {bullets.map(bullet => (
            <li key={bullet}>{bullet}</li>
          ))}
        </ul>
      </div>
    </div>
  );
}
