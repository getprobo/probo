// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { RichEditor } from "@probo/ui";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ErrorInfo, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { RiskAnalysisDescriptionSection_riskAnalysis$key } from "#/__generated__/core/RiskAnalysisDescriptionSection_riskAnalysis.graphql";
import { isRichEditorContentEmpty } from "#/pages/organizations/_lib/richEditorContent";

import { riskAnalysisDescriptionField } from "../variants";

const riskAnalysisDescriptionSectionFragment = graphql`
  fragment RiskAnalysisDescriptionSection_riskAnalysis on RiskAnalysis {
    description
  }
`;

interface RiskAnalysisDescriptionSectionProps {
  riskAnalysisKey: RiskAnalysisDescriptionSection_riskAnalysis$key;
  fallback?: ReactNode;
  onError?: (error: unknown, info: ErrorInfo) => void;
}

function RiskAnalysisDescriptionSectionContent({
  riskAnalysisKey,
}: {
  riskAnalysisKey: RiskAnalysisDescriptionSection_riskAnalysis$key;
}) {
  const { t } = useTranslation();
  const riskAnalysis = useFragment(
    riskAnalysisDescriptionSectionFragment,
    riskAnalysisKey,
  );
  const description = riskAnalysis.description ?? "";

  if (isRichEditorContentEmpty(description)) {
    return null;
  }

  return (
    <RichEditor
      key={description}
      className={riskAnalysisDescriptionField({ density: "display" }).editor()}
      content={description}
      disabled
      aria-label={t("riskAnalysisDetailPage.fields.description")}
    />
  );
}

export function RiskAnalysisDescriptionSection({
  riskAnalysisKey,
  fallback,
  onError,
}: RiskAnalysisDescriptionSectionProps) {
  const { t } = useTranslation();

  return (
    <ErrorBoundary
      fallback={fallback ?? (
        <Text size={2} color="faint">
          {t("riskAnalysisDetailPage.errors.content")}
        </Text>
      )}
      onError={onError}
    >
      <RiskAnalysisDescriptionSectionContent riskAnalysisKey={riskAnalysisKey} />
    </ErrorBoundary>
  );
}
