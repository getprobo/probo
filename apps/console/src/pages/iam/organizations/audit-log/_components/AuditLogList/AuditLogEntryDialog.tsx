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
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { AuditLogEntryDialog_entry$key } from "#/__generated__/iam/AuditLogEntryDialog_entry.graphql";

import {
  auditLogActorColor,
  auditLogActorLabel,
  auditLogOutcomeColor,
} from "../../_lib/auditLogEntry";
import { auditLogEntryDialog } from "../variants";

const auditLogEntryDialogFragment = graphql`
  fragment AuditLogEntryDialog_entry on AuditLogEntry {
    actorId
    actorType
    actorEmail
    actorName
    action
    resourceType
    resourceId
    outcome
    denialReason
    originIp
    userAgent
    requestId
    createdAt
    recordedAt
  }
`;

const AUDIT_LOG_NS = "iam/organizations/audit-log";

interface AuditLogEntryDialogProps {
  entryKey: AuditLogEntryDialog_entry$key;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AuditLogEntryDialog({
  entryKey,
  open,
  onOpenChange,
}: AuditLogEntryDialogProps) {
  const { t, i18n } = useTranslation(AUDIT_LOG_NS);
  const entry = useFragment(auditLogEntryDialogFragment, entryKey);
  const {
    actorId,
    actorType,
    actorEmail,
    actorName,
    action,
    resourceType,
    resourceId,
    outcome,
    denialReason,
    originIp,
    userAgent,
    requestId,
    createdAt,
    recordedAt,
  } = entry;
  const { details, actor, target } = auditLogEntryDialog();
  const label = auditLogActorLabel({ actorEmail, actorName, actorId });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("auditLogPage.entry.title")}</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <dl className={details()}>
            <Detail label={t("auditLogPage.entry.when")}>
              <Text size={2} highContrast>
                {dateTimeFormat(i18n.language, createdAt)}
              </Text>
            </Detail>
            <Detail label={t("auditLogPage.entry.recordedAt")}>
              <Text size={2} highContrast>
                {dateTimeFormat(i18n.language, recordedAt)}
              </Text>
            </Detail>
            <Detail label={t("auditLogPage.entry.actor")}>
              <span className={actor()}>
                <Badge variant="soft" size={1} color={auditLogActorColor(actorType)}>
                  {t(`auditLogPage.actorTypes.${actorType}`)}
                </Badge>
                <Text size={2} highContrast>{label}</Text>
              </span>
            </Detail>
            <Detail label={t("auditLogPage.entry.actorId")}>
              <Code size={2} variant="ghost">{actorId}</Code>
            </Detail>
            <Detail label={t("auditLogPage.entry.action")}>
              <Code size={2} variant="ghost">{action}</Code>
            </Detail>
            <Detail label={t("auditLogPage.entry.target")}>
              <span className={target()}>
                <Text size={2} highContrast>{resourceType}</Text>
                <Code size={1} variant="ghost">{resourceId}</Code>
              </span>
            </Detail>
            <Detail label={t("auditLogPage.entry.outcome")}>
              <Badge variant="soft" size={1} color={auditLogOutcomeColor(outcome)}>
                {t(`auditLogPage.outcomes.${outcome}`)}
              </Badge>
            </Detail>
            {denialReason != null && denialReason !== "" && (
              <Detail label={t("auditLogPage.entry.denialReason")}>
                {textValue(denialReason)}
              </Detail>
            )}
            <Detail label={t("auditLogPage.entry.originIp")}>
              {textValue(originIp)}
            </Detail>
            <Detail label={t("auditLogPage.entry.userAgent")}>
              {textValue(userAgent)}
            </Detail>
            <Detail label={t("auditLogPage.entry.requestId")}>
              {requestId == null || requestId === ""
                ? <Text size={2} color="faint">—</Text>
                : <Code size={2} variant="ghost">{requestId}</Code>}
            </Detail>
          </dl>
        </DialogBody>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("auditLogPage.entry.close")}
              </Button>
            )}
          />
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}

function Detail({ label, children }: { label: string; children: ReactNode }) {
  const { row, value } = auditLogEntryDialog();

  return (
    <div className={row()}>
      <dt>
        <Text size={1} color="faint">{label}</Text>
      </dt>
      <dd className={value()}>{children}</dd>
    </div>
  );
}

function textValue(value: string | null | undefined) {
  if (value == null || value === "") {
    return <Text size={2} color="faint">—</Text>;
  }

  return <Text size={2} highContrast>{value}</Text>;
}
