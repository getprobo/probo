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
} from "../_lib/auditLogEntry";
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
  const { details, actor } = auditLogEntryDialog();
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
              {dateTimeFormat(i18n.language, createdAt)}
            </Detail>
            <Detail label={t("auditLogPage.entry.recordedAt")}>
              {dateTimeFormat(i18n.language, recordedAt)}
            </Detail>
            <Detail label={t("auditLogPage.entry.actor")}>
              <span className={actor()}>
                <Badge variant="soft" size={1} color={auditLogActorColor(actorType)}>
                  {t(`auditLogPage.actorTypes.${actorType}`)}
                </Badge>
                <span>{label}</span>
              </span>
            </Detail>
            <Detail label={t("auditLogPage.entry.actorId")}>
              {actorId}
            </Detail>
            <Detail label={t("auditLogPage.entry.action")}>
              {action}
            </Detail>
            <Detail label={t("auditLogPage.entry.target")}>
              {`${resourceType} ${resourceId}`}
            </Detail>
            <Detail label={t("auditLogPage.entry.outcome")}>
              <Badge variant="soft" size={1} color={auditLogOutcomeColor(outcome)}>
                {t(`auditLogPage.outcomes.${outcome}`)}
              </Badge>
            </Detail>
            {denialReason != null && denialReason !== "" && (
              <Detail label={t("auditLogPage.entry.denialReason")}>
                {denialReason}
              </Detail>
            )}
            <Detail label={t("auditLogPage.entry.originIp")}>
              {originIp ?? ""}
            </Detail>
            <Detail label={t("auditLogPage.entry.userAgent")}>
              {userAgent ?? ""}
            </Detail>
            <Detail label={t("auditLogPage.entry.requestId")}>
              {requestId ?? ""}
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
  const text = typeof children === "string" ? children : null;

  return (
    <div className={row()}>
      <dt>
        <Text size={1} color="faint">{label}</Text>
      </dt>
      <dd className={value()}>
        {text == null
          ? children
          : text === ""
            ? <Text size={2} color="faint">—</Text>
            : <Text size={2} highContrast>{text}</Text>}
      </dd>
    </div>
  );
}
