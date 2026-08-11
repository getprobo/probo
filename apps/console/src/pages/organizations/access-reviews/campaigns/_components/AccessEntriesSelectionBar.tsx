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

import { Button, Option, Select } from "@probo/ui";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";

import type { AccessReviewEntryDecision } from "#/__generated__/core/CampaignDetailPageBulkDecisionMutation.graphql";
import type { AccessReviewEntryFlag } from "#/__generated__/core/CampaignDetailPageBulkFlagMutation.graphql";

import { flagGroups } from "../../_components/accessReviewHelpers";

import { FilterMultiSelect } from "./FilterMultiSelect";
import { accessEntriesSelectionBar } from "./variants";

// Radix Select reserves "" — use a sentinel for the clear-decision option.
const NO_DECISION_VALUE = "__none__";

interface AccessEntriesSelectionBarProps {
  selectedCount: number;
  selectedIds: ReadonlyArray<string>;
  allEntryIds: string[];
  hasUnloadedEntries: boolean;
  onClear: () => void;
  onSelectAll: () => void;
  bulkDecision: AccessReviewEntryDecision | null;
  onBulkDecisionChange: (decision: AccessReviewEntryDecision | null) => void;
  bulkFlagSelection: AccessReviewEntryFlag[];
  onBulkFlagSelectionChange: (flags: AccessReviewEntryFlag[]) => void;
  bulkFlagsDirty: boolean;
  isSubmitting: boolean;
  onSubmit: () => void;
}

// Fixed bottom bar matching compliance-portal documents selection chrome.
export function AccessEntriesSelectionBar({
  selectedCount,
  selectedIds,
  allEntryIds,
  hasUnloadedEntries,
  onClear,
  onSelectAll,
  bulkDecision,
  onBulkDecisionChange,
  bulkFlagSelection,
  onBulkFlagSelectionChange,
  bulkFlagsDirty,
  isSubmitting,
  onSubmit,
}: AccessEntriesSelectionBarProps) {
  const { t } = useTranslation();
  const { bar, inner, actions } = accessEntriesSelectionBar();

  const flagOptions = useMemo(
    () => flagGroups.flatMap(group =>
      group.flags.map(flag => ({
        value: flag.value,
        label: t(`campaignDetailPage.flags.${flag.value.toLowerCase()}`),
      })),
    ),
    [t],
  );

  const allSelected = useMemo(() => {
    if (allEntryIds.length === 0) {
      return false;
    }
    const selectedIdSet = new Set(selectedIds);
    return allEntryIds.every(id => selectedIdSet.has(id));
  }, [allEntryIds, selectedIds]);
  const canSubmit = bulkDecision !== null || bulkFlagsDirty;

  if (selectedCount === 0) {
    return null;
  }

  return (
    <div className={bar()}>
      <div className={inner()}>
        <div className="flex flex-col">
          <span className="text-sm font-medium text-txt-primary">
            {t("campaignDetailPage.selected", { count: selectedCount })}
          </span>
          {hasUnloadedEntries && (
            <span className="text-xs text-txt-tertiary">
              {t("campaignDetailPage.selectionLoadedOnly")}
            </span>
          )}
        </div>
        <div className={actions()}>
          <Button variant="tertiary" onClick={onClear} disabled={isSubmitting}>
            {t("campaignDetailPage.actions.clearSelection")}
          </Button>
          <Button
            variant="tertiary"
            onClick={onSelectAll}
            disabled={isSubmitting || allEntryIds.length === 0 || allSelected}
          >
            {hasUnloadedEntries
              ? t("campaignDetailPage.actions.selectAllLoaded", {
                  loaded: allEntryIds.length,
                })
              : t("campaignDetailPage.actions.selectAll")}
          </Button>
          <Select
            variant="editor"
            value={bulkDecision ?? undefined}
            placeholder={t("campaignDetailPage.decisionPlaceholder")}
            disabled={isSubmitting}
            onValueChange={(value) => {
              if (value === NO_DECISION_VALUE) {
                onBulkDecisionChange(null);
                return;
              }
              onBulkDecisionChange(value as AccessReviewEntryDecision);
            }}
          >
            <Option value={NO_DECISION_VALUE}>
              {t("campaignDetailPage.actions.noDecision")}
            </Option>
            <Option value={"APPROVED" satisfies AccessReviewEntryDecision}>
              {t("campaignDetailPage.actions.approve")}
            </Option>
            <Option value={"REVOKE" satisfies AccessReviewEntryDecision}>
              {t("campaignDetailPage.actions.revoke")}
            </Option>
            <Option value={"DEFER" satisfies AccessReviewEntryDecision}>
              {t("campaignDetailPage.actions.modify")}
            </Option>
            <Option value={"ESCALATE" satisfies AccessReviewEntryDecision}>
              {t("campaignDetailPage.actions.escalate")}
            </Option>
          </Select>
          <div className="w-44">
            <FilterMultiSelect
              placeholder={t("campaignDetailPage.flagsPlaceholder")}
              options={flagOptions}
              value={bulkFlagSelection}
              onChange={values => onBulkFlagSelectionChange(values as AccessReviewEntryFlag[])}
              side="top"
              disabled={isSubmitting}
            />
          </div>
          <Button
            disabled={!canSubmit || isSubmitting}
            onClick={onSubmit}
          >
            {t("campaignDetailPage.actions.submit")}
          </Button>
        </div>
      </div>
    </div>
  );
}
