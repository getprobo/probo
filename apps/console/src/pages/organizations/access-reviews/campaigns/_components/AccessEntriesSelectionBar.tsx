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

interface AccessEntriesSelectionBarProps {
  selectedCount: number;
  allEntryIds: string[];
  onClear: () => void;
  onSelectAll: () => void;
  bulkDecision: AccessReviewEntryDecision | null;
  onBulkDecisionChange: (decision: AccessReviewEntryDecision) => void;
  bulkFlagSelection: AccessReviewEntryFlag[];
  onBulkFlagSelectionChange: (flags: AccessReviewEntryFlag[]) => void;
  onSubmit: () => void;
}

// Fixed bottom bar matching compliance-portal documents selection chrome.
export function AccessEntriesSelectionBar({
  selectedCount,
  allEntryIds,
  onClear,
  onSelectAll,
  bulkDecision,
  onBulkDecisionChange,
  bulkFlagSelection,
  onBulkFlagSelectionChange,
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

  const canSubmit = bulkDecision !== null || bulkFlagSelection.length > 0;

  if (selectedCount === 0) {
    return null;
  }

  return (
    <div className={bar()}>
      <div className={inner()}>
        <span className="text-sm font-medium text-txt-primary">
          {t("campaignDetailPage.selected", { count: selectedCount })}
        </span>
        <div className={actions()}>
          <Button variant="tertiary" onClick={onClear}>
            {t("campaignDetailPage.actions.clearSelection")}
          </Button>
          <Button
            variant="tertiary"
            onClick={onSelectAll}
            disabled={allEntryIds.length === 0}
          >
            {t("campaignDetailPage.actions.selectAll")}
          </Button>
          <Select
            variant="editor"
            value={bulkDecision ?? undefined}
            placeholder={t("campaignDetailPage.decisionPlaceholder")}
            onValueChange={value => onBulkDecisionChange(value as AccessReviewEntryDecision)}
          >
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
            />
          </div>
          <Button
            disabled={!canSubmit}
            onClick={onSubmit}
          >
            {t("campaignDetailPage.actions.submit")}
          </Button>
        </div>
      </div>
    </div>
  );
}
