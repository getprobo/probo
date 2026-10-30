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

import { dateTimeFormat } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { TableButton } from "@probo/ui/src/v2/Table/TableButton";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { AuditLogListItem_entry$key } from "#/__generated__/iam/AuditLogListItem_entry.graphql";

import {
  auditLogActorColor,
  auditLogActorLabel,
  auditLogOutcomeColor,
} from "../_lib/auditLogEntry";
import { auditLogListItem } from "../variants";

const AUDIT_LOG_NS = "iam/organizations/audit-log";

export const auditLogListItemFragment = graphql`
  fragment AuditLogListItem_entry on AuditLogEntry {
    actorId
    actorType
    actorEmail
    actorName
    action
    resourceType
    outcome
    createdAt
  }
`;

interface AuditLogListItemProps {
  entryKey: AuditLogListItem_entry$key;
  onOpen: () => void;
}

export function AuditLogListItem({ entryKey, onOpen }: AuditLogListItemProps) {
  const { t, i18n } = useTranslation(AUDIT_LOG_NS);
  const entry = useFragment(auditLogListItemFragment, entryKey);
  const {
    actorId,
    actorType,
    actorEmail,
    actorName,
    action,
    resourceType,
    outcome,
    createdAt,
  } = entry;
  const { actor, truncate, action: actionSlot } = auditLogListItem();
  const label = auditLogActorLabel({ actorEmail, actorName, actorId });

  return (
    <TableRow align="center" interactive>
      <TableRowHeaderCell>
        <TableButton
          aria-label={t("auditLogPage.actions.openEntry")}
          onClick={onOpen}
        >
          <Text size={2}>
            {dateTimeFormat(i18n.language, createdAt)}
          </Text>
        </TableButton>
      </TableRowHeaderCell>
      <TableCell overflow="truncate">
        <div className={actor()}>
          <Badge variant="soft" size={1} color={auditLogActorColor(actorType)}>
            {t(`auditLogPage.actorTypes.${actorType}`)}
          </Badge>
          <Text size={2} highContrast className={truncate()}>
            {label}
          </Text>
        </div>
      </TableCell>
      <TableCell overflow="truncate">
        <Code size={2} variant="ghost" className={actionSlot()}>
          {action}
        </Code>
      </TableCell>
      <TableCell overflow="truncate">
        <Text size={2} highContrast className={truncate()}>
          {resourceType}
        </Text>
      </TableCell>
      <TableCell>
        <Badge variant="soft" size={1} color={auditLogOutcomeColor(outcome)}>
          {t(`auditLogPage.outcomes.${outcome}`)}
        </Badge>
      </TableCell>
    </TableRow>
  );
}
