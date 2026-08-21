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

import { Anchor, Card } from "@probo/ui";
import { useTranslation } from "react-i18next";

import {
  OnboardingStepActions,
  OnboardingStepDoCard,
} from "./OnboardingStepActions";

type Props = {
  onContinue: () => void;
  onDefer: () => void;
};

export function McpOnboardingDoPanel(props: Props) {
  const { onContinue, onDefer } = props;
  const { t } = useTranslation("organizations/onboarding");

  const mcpUrl = `${window.location.origin}/mcp/v1`;

  return (
    <OnboardingStepDoCard
      title={t("steps.mcp.doTitle")}
      description={t("steps.mcp.doDescription")}
      actions={(
        <OnboardingStepActions
          onContinue={onContinue}
          onDefer={onDefer}
        />
      )}
    >
      <ol className="list-decimal space-y-3 pl-5 text-2 text-txt-secondary">
        <li>{t("steps.mcp.steps.oauth")}</li>
        <li>{t("steps.mcp.steps.url")}</li>
        <li>{t("steps.mcp.steps.skills")}</li>
      </ol>
      <Card padded className="mt-4 bg-level-2">
        <p className="text-1 text-txt-tertiary mb-1">{t("steps.mcp.endpointLabel")}</p>
        <code className="text-2 break-all text-txt-primary">{mcpUrl}</code>
      </Card>
      <p className="text-2 text-txt-secondary">
        <Anchor
          href="https://github.com/getprobo/probo/tree/main/packages/skills"
          target="_blank"
          rel="noreferrer"
        >
          {t("steps.mcp.docsLink")}
        </Anchor>
      </p>
    </OnboardingStepDoCard>
  );
}
