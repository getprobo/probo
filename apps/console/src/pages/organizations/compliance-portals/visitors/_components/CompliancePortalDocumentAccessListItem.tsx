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
import { getCompliancePortalDocumentAccessStatusLabel } from "@probo/helpers";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import {
  documentAccessStatusColor,
  rejectOrRevokeStatus,
} from "../_lib/documentAccessInfo";
import { documentAccessList } from "../variants";

interface CompliancePortalDocumentAccessListItemProps {
  documentAccess: CompliancePortalDocumentAccessInfo;
  canUpdate: boolean;
  disabled: boolean;
  selected?: boolean;
  onSelectedChange?: () => void;
  onGrant: (documentAccess: CompliancePortalDocumentAccessInfo) => void;
  onRejectOrRevoke: (documentAccess: CompliancePortalDocumentAccessInfo) => void;
}

export function CompliancePortalDocumentAccessListItem({
  documentAccess,
  canUpdate,
  disabled,
  selected,
  onSelectedChange,
  onGrant,
  onRejectOrRevoke,
}: CompliancePortalDocumentAccessListItemProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { t: tRoot } = useTranslation();
  const { titleRow, title, meta, badge, trailing } = documentAccessList();
  const showGrant = canUpdate && documentAccess.status !== "GRANTED";
  const showRejectOrRevoke = canUpdate
    && documentAccess.status !== "REJECTED"
    && documentAccess.status !== "REVOKED";
  const revoke = documentAccess.status === "GRANTED";
  const kind = documentAccess.type === "document" && documentAccess.category !== ""
    ? tRoot(`documentTypeOptions.types.${documentAccess.category}`)
    : documentAccess.category;

  return (
    <ListItem>
      {onSelectedChange != null && (
        <Checkbox
          checked={selected ?? false}
          disabled={disabled}
          onCheckedChange={onSelectedChange}
          aria-label={t("documentAccessList.selection.selectRow", { title: documentAccess.name })}
        />
      )}
      <ListItemContent>
        <div className={titleRow()}>
          <Text size={2} weight="medium" color="neutral" highContrast className={title()}>
            {documentAccess.name}
          </Text>
          {documentAccess.status != null && (
            <Badge
              size={1}
              variant="soft"
              color={documentAccessStatusColor(documentAccess.status)}
              className={badge()}
            >
              {getCompliancePortalDocumentAccessStatusLabel(documentAccess.status, tRoot)}
            </Badge>
          )}
        </div>
        {kind !== "" && (
          <Text size={1} color="gold" className={meta()}>
            {kind}
          </Text>
        )}
      </ListItemContent>
      {(showGrant || showRejectOrRevoke) && (
        <div className={trailing()}>
          {showGrant && (
            <Button
              type="button"
              size={2}
              variant="ghost"
              color="neutral"
              disabled={disabled}
              iconStart={<CheckIcon />}
              onClick={() => onGrant(documentAccess)}
            >
              {t("documentAccessList.actions.grant")}
            </Button>
          )}
          {showRejectOrRevoke && (
            <Button
              type="button"
              size={2}
              variant="ghost"
              color="red"
              disabled={disabled}
              iconStart={revoke ? <ArrowCounterClockwiseIcon /> : <ProhibitIcon />}
              onClick={() => onRejectOrRevoke({
                ...documentAccess,
                status: rejectOrRevokeStatus(documentAccess.status),
              })}
            >
              {revoke
                ? t("documentAccessList.actions.revoke")
                : t("documentAccessList.actions.reject")}
            </Button>
          )}
        </div>
      )}
    </ListItem>
  );
}
