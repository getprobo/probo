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

import { ArrowCounterClockwiseIcon, CheckIcon, ProhibitIcon } from "@phosphor-icons/react";
import type { CompliancePortalDocumentAccessInfo } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { rejectOrRevokeStatus } from "../_lib/documentAccessInfo";
import { documentAccessSelectionBar } from "../variants";

import { CompliancePortalDocumentAccessBulkDialog } from "./CompliancePortalDocumentAccessBulkDialog";

interface CompliancePortalDocumentAccessSelectionBarProps {
  selectedItems: CompliancePortalDocumentAccessInfo[];
  allSelected: boolean;
  loading: boolean;
  onClear: () => void;
  onSelectAll: () => void;
  onGrant: (items: CompliancePortalDocumentAccessInfo[]) => Promise<void>;
  onRejectOrRevoke: (items: CompliancePortalDocumentAccessInfo[]) => Promise<void>;
}

export function CompliancePortalDocumentAccessSelectionBar({
  selectedItems,
  allSelected,
  loading,
  onClear,
  onSelectAll,
  onGrant,
  onRejectOrRevoke,
}: CompliancePortalDocumentAccessSelectionBarProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const grantable = selectedItems.filter(item => item.status !== "GRANTED");
  const rejectable = selectedItems.filter(
    item => item.status !== "REJECTED" && item.status !== "REVOKED",
  );
  const revokeOnly = rejectable.length > 0 && rejectable.every(item => item.status === "GRANTED");
  const mixedReject = rejectable.some(item => item.status === "GRANTED") && !revokeOnly;
  const rejectLabel = revokeOnly
    ? t("documentAccessList.actions.revoke")
    : mixedReject
      ? t("documentAccessList.actions.rejectOrRevoke")
      : t("documentAccessList.actions.reject");
  const { bar, inner, actions } = documentAccessSelectionBar();

  if (selectedItems.length === 0) {
    return null;
  }

  return (
    <div className={bar()}>
      <div className={inner()}>
        <Text size={2} weight="medium" color="neutral" highContrast>
          {t("documentAccessList.selection.count", { count: selectedItems.length })}
        </Text>
        <div className={actions()}>
          <Button variant="ghost" color="neutral" disabled={loading} onClick={onClear}>
            {t("documentAccessList.selection.clear")}
          </Button>
          <Button
            variant="ghost"
            color="neutral"
            disabled={loading || allSelected}
            onClick={onSelectAll}
          >
            {t("documentAccessList.selection.selectAll")}
          </Button>
          {grantable.length > 0 && (
            <CompliancePortalDocumentAccessBulkDialog
              action="grant"
              count={grantable.length}
              loading={loading}
              onConfirm={() => onGrant(grantable.map(item => ({ ...item, status: "GRANTED" })))}
            >
              <Button
                type="button"
                variant="solid"
                color="neutral"
                highContrast
                disabled={loading}
                iconStart={<CheckIcon />}
              >
                {t("documentAccessList.actions.grant")}
              </Button>
            </CompliancePortalDocumentAccessBulkDialog>
          )}
          {rejectable.length > 0 && (
            <CompliancePortalDocumentAccessBulkDialog
              action="rejectOrRevoke"
              count={rejectable.length}
              loading={loading}
              onConfirm={() => onRejectOrRevoke(
                rejectable.map(item => ({
                  ...item,
                  status: rejectOrRevokeStatus(item.status),
                })),
              )}
            >
              <Button
                type="button"
                variant="solid"
                color="red"
                disabled={loading}
                iconStart={revokeOnly ? <ArrowCounterClockwiseIcon /> : <ProhibitIcon />}
              >
                {rejectLabel}
              </Button>
            </CompliancePortalDocumentAccessBulkDialog>
          )}
        </div>
      </div>
    </div>
  );
}
