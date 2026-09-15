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

import { RichEditor } from "@probo/ui";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ErrorInfo, type ReactNode, useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { RiskAnalysisDescriptionSection_riskAnalysis$key } from "#/__generated__/core/RiskAnalysisDescriptionSection_riskAnalysis.graphql";

import { isRichEditorContentEmpty } from "../_lib/richEditorContent";
import { useDebouncedSerializedFieldSave } from "../_lib/useSerializedFieldSave";
import { useUpdateRiskAnalysis } from "../_lib/useUpdateRiskAnalysis";
import { riskAnalysisDescriptionSection } from "../variants";

const riskAnalysisDescriptionSaveDelayMs = 1000;

const riskAnalysisDescriptionSectionFragment = graphql`
  fragment RiskAnalysisDescriptionSection_riskAnalysis on RiskAnalysis {
    id
    description
    canUpdate: permission(action: "risk-management:risk-analysis:update")
  }
`;

interface RiskAnalysisDescriptionSectionProps {
  riskAnalysisKey: RiskAnalysisDescriptionSection_riskAnalysis$key;
  fallback?: ReactNode;
  onError?: (error: unknown, info: ErrorInfo) => void;
}

function normalizeContent(value: string) {
  return isRichEditorContentEmpty(value) ? "" : value;
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
  const [updateRiskAnalysis] = useUpdateRiskAnalysis();
  const saved = riskAnalysis.description ?? "";
  const [draft, setDraft] = useState(saved);
  const [savedContent, setSavedContent] = useState(saved);
  const [dirty, setDirty] = useState(false);
  const [editorGeneration, setEditorGeneration] = useState(0);

  if (saved !== savedContent) {
    setSavedContent(saved);
    if (!dirty) {
      setDraft(saved);
      setEditorGeneration(generation => generation + 1);
    }
  }

  const persist = useCallback(
    async (value: string) => {
      const next = normalizeContent(value);

      try {
        await updateRiskAnalysis({
          variables: {
            input: {
              id: riskAnalysis.id,
              description: next || null,
            },
          },
        });
        setDraft((current) => {
          if (current === value || normalizeContent(current) === next) {
            setDirty(false);
            return next;
          }
          return current;
        });
      } catch {
        setDraft((current) => {
          if (normalizeContent(current) === next) {
            setDirty(false);
            setEditorGeneration(generation => generation + 1);
            return saved;
          }
          return current;
        });
      }
    },
    [riskAnalysis.id, saved, updateRiskAnalysis],
  );
  const persistDebounced = useDebouncedSerializedFieldSave(
    persist,
    riskAnalysisDescriptionSaveDelayMs,
  );
  const { root, editor } = riskAnalysisDescriptionSection();

  return (
    <div className={root()}>
      {riskAnalysis.canUpdate
        ? (
            <RichEditor
              key={editorGeneration}
              className={editor()}
              content={draft}
              aria-label={t("riskAnalysisDetailPage.fields.description")}
              onChangeContent={(next) => {
                setDirty(true);
                setDraft(next);
                persistDebounced.schedule(next);
              }}
              onBlur={() => {
                persistDebounced.flush();
              }}
            />
          )
        : isRichEditorContentEmpty(saved)
          ? (
              <Text size={2} color="faint">
                {t("riskAnalysisDetailPage.noDescription")}
              </Text>
            )
          : (
              <RichEditor
                className={editor()}
                content={saved}
                disabled
                aria-label={t("riskAnalysisDetailPage.fields.description")}
              />
            )}
    </div>
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
