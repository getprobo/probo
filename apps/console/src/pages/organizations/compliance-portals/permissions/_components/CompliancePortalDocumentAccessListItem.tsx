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

import type { CompliancePortalDocumentAccessInfo } from "@probo/helpers";
import { getCompliancePortalDocumentAccessStatusLabel } from "@probo/helpers";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import {
  documentAccessStatusColor,
  documentAccessTypeColor,
  rejectOrRevokeStatus,
} from "../_lib/documentAccessInfo";
import { documentAccessList } from "../variants";

interface CompliancePortalDocumentAccessListItemProps {
  documentAccess: CompliancePortalDocumentAccessInfo;
  canUpdate: boolean;
  disabled: boolean;
  onGrant: (documentAccess: CompliancePortalDocumentAccessInfo) => void;
  onRejectOrRevoke: (documentAccess: CompliancePortalDocumentAccessInfo) => void;
}

export function CompliancePortalDocumentAccessListItem({
  documentAccess,
  canUpdate,
  disabled,
  onGrant,
  onRejectOrRevoke,
}: CompliancePortalDocumentAccessListItemProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { t: tRoot } = useTranslation();
  const { rowActions } = documentAccessList();
  const showStatus = documentAccess.persisted || documentAccess.status !== "REQUESTED";
  const showGrant = canUpdate && documentAccess.status !== "GRANTED";
  const showRejectOrRevoke = canUpdate
    && documentAccess.status !== "REJECTED"
    && documentAccess.status !== "REVOKED";
  const revoke = documentAccess.status === "GRANTED";

  return (
    <TableRow align="center">
      <TableRowHeaderCell minWidth="12rem">
        <Text size={2} weight="medium" highContrast className="truncate">
          {documentAccess.name}
        </Text>
      </TableRowHeaderCell>
      <TableCell>
        <Badge variant="soft" color={documentAccessTypeColor(documentAccess.variant)}>
          {documentAccess.typeLabel}
        </Badge>
      </TableCell>
      <TableCell>
        <Text size={2} color="neutral">
          {documentAccess.category || "—"}
        </Text>
      </TableCell>
      <TableCell>
        {showStatus && (
          <Badge variant="soft" color={documentAccessStatusColor(documentAccess.status)}>
            {getCompliancePortalDocumentAccessStatusLabel(documentAccess.status, tRoot)}
          </Badge>
        )}
      </TableCell>
      <TableCell justify="end">
        {(showGrant || showRejectOrRevoke) && (
          <div className={rowActions()}>
            {showGrant && (
              <Button
                type="button"
                size={1}
                variant="soft"
                color="neutral"
                disabled={disabled}
                onClick={() => onGrant(documentAccess)}
              >
                {t("documentAccessList.actions.grant")}
              </Button>
            )}
            {showRejectOrRevoke && (
              <Button
                type="button"
                size={1}
                variant="soft"
                color="red"
                disabled={disabled}
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
      </TableCell>
    </TableRow>
  );
}
