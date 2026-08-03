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

import { Badge, Button } from "@probo/ui";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

type OnboardingStepActionsProps = {
  continueLabel?: string;
  continueDisabled?: boolean;
  onContinue: () => void;
  onDefer?: () => void;
  showDefer?: boolean;
  settingsLink?: ReactNode;
};

export function OnboardingStepActions(props: OnboardingStepActionsProps) {
  const {
    continueDisabled = false,
    continueLabel,
    onContinue,
    onDefer,
    settingsLink,
    showDefer = true,
  } = props;
  const { t } = useTranslation("organizations/onboarding");

  return (
    <div className="flex flex-col gap-4 border-t border-border-mid pt-6">
      {settingsLink}
      <div className="flex flex-col-reverse gap-3 sm:flex-row sm:items-center sm:justify-between">
        {showDefer && onDefer
          ? (
              <Button variant="ghost" type="button" onClick={onDefer}>
                {t("actions.doLater")}
              </Button>
            )
          : <span />}
        <Button
          type="button"
          disabled={continueDisabled}
          onClick={onContinue}
          className="sm:min-w-40"
        >
          {continueLabel ?? t("actions.continue")}
        </Button>
      </div>
    </div>
  );
}

export function OnboardingStepDoCard(props: {
  title: string;
  description?: string;
  complete?: boolean;
  children: ReactNode;
  actions: ReactNode;
}) {
  const { actions, children, complete, description, title } = props;
  const { t } = useTranslation("organizations/onboarding");

  return (
    <div className="space-y-6 rounded-4 border border-border-mid bg-level-1 p-6 motion-safe:animate-in motion-safe:fade-in motion-safe:slide-in-from-bottom-2">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div className="space-y-1">
          <h3 className="text-3 font-medium text-txt-primary">{title}</h3>
          {description && (
            <p className="text-2 text-txt-secondary">{description}</p>
          )}
        </div>
        {complete && (
          <Badge variant="success">{t("status.complete")}</Badge>
        )}
      </div>
      {children}
      {actions}
    </div>
  );
}
