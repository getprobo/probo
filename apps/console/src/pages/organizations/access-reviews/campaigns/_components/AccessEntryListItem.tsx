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

import { dateFormat } from "@probo/i18n";
import { Badge, Checkbox, IconRobot } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { AccessEntryListItem_entry$key } from "#/__generated__/core/AccessEntryListItem_entry.graphql";

import {
  BooleanStatusIcon,
  decisionBadgeVariant,
  flagBadgeVariant,
  LastLoginStatus,
  MfaStatusIcon,
} from "../../_components/accessReviewHelpers";
import { EntryDecisionActions } from "../../_components/EntryDecisionActions";
import { EntryFlagSelect } from "../../_components/EntryFlagSelect";

import { accessEntryList } from "./variants";

const accessEntryListItemFragment = graphql`
  fragment AccessEntryListItem_entry on AccessReviewEntry {
    id
    email
    fullName
    roles
    isAdmin
    active
    mfaStatus
    accountType
    lastLogin
    decision
    flags
  }
`;

interface AccessEntryListItemProps {
  entryKey: AccessEntryListItem_entry$key;
  isPendingActions: boolean;
  selected: boolean;
  onSelectedChange: (event: { shiftKey: boolean }) => void;
}

export function AccessEntryListItem({
  entryKey,
  isPendingActions,
  selected,
  onSelectedChange,
}: AccessEntryListItemProps) {
  const { i18n, t } = useTranslation();
  const entry = useFragment(accessEntryListItemFragment, entryKey);
  const { item, content, flags, trailing, status, statusLabel } = accessEntryList({
    inactive: entry.active === false,
  });

  const fullName = entry.fullName.trim();
  const email = entry.email.trim();
  const title = fullName || email;
  const role = entry.roles[0] ?? null;
  const extraRoles = entry.roles.length > 1 ? entry.roles.length - 1 : 0;

  return (
    <li className={item()}>
      {isPendingActions && (
        <div
          className="shrink-0"
          onClickCapture={(event) => {
            event.preventDefault();
            event.stopPropagation();
            onSelectedChange({ shiftKey: event.shiftKey });
          }}
        >
          <Checkbox
            checked={selected}
            onChange={() => {}}
            aria-label={t("campaignDetailPage.selectEntry", {
              name: title || email || "N/A",
            })}
          />
        </div>
      )}

      <div className={content()}>
        <div className="flex min-w-0 items-center gap-1.5">
          {entry.accountType === "SERVICE_ACCOUNT" && (
            <IconRobot size={16} className="shrink-0 text-txt-tertiary" />
          )}
          <span className="truncate text-sm font-medium text-txt-primary" title={title}>
            {title || <span className="text-txt-tertiary">N/A</span>}
          </span>
        </div>
        <div className="flex min-w-0 items-center gap-1.5 text-xs text-txt-tertiary">
          {fullName && email
            ? (
                <span className="truncate" title={email}>{email}</span>
              )
            : null}
          {fullName && email && role
            ? <span className="shrink-0">·</span>
            : null}
          {role
            ? (
                <span className="truncate" title={entry.roles.join(", ")}>
                  {role}
                </span>
              )
            : null}
          {extraRoles > 0 && (
            <span className="shrink-0">{`+${extraRoles}`}</span>
          )}
        </div>
      </div>

      {isPendingActions
        ? (
            <div className={flags()}>
              <EntryFlagSelect
                entryId={entry.id}
                currentFlags={entry.flags}
              />
            </div>
          )
        : entry.flags.length > 0
          ? (
              <div className={flags()}>
                {entry.flags.map(f => (
                  <Badge key={f} variant={flagBadgeVariant(f)}>
                    {t(`campaignDetailPage.flags.${f.toLowerCase()}`)}
                  </Badge>
                ))}
              </div>
            )
          : null}

      <div className={trailing()}>
        <div className={status()}>
          <span className={statusLabel()}>{t("campaignDetailPage.columns.admin")}</span>
          <BooleanStatusIcon
            value={entry.isAdmin}
            trueLabel={t("campaignDetailPage.values.yes")}
            falseLabel={t("campaignDetailPage.values.no")}
            unknownLabel={t("campaignDetailPage.values.unknown")}
          />
        </div>
        <div className={status()}>
          <span className={statusLabel()}>{t("campaignDetailPage.columns.mfa")}</span>
          <MfaStatusIcon
            status={entry.mfaStatus}
            label={t(`campaignDetailPage.mfaStatus.${entry.mfaStatus.toLowerCase()}`)}
          />
        </div>
        <div className={status()}>
          <span className={statusLabel()}>{t("campaignDetailPage.columns.lastLogin")}</span>
          <LastLoginStatus
            lastLogin={entry.lastLogin}
            formatted={
              entry.lastLogin
                ? dateFormat(i18n.language, entry.lastLogin)
                : ""
            }
            unknownLabel={t("campaignDetailPage.values.unknown")}
            compact
          />
        </div>

        {isPendingActions
          ? (
              <EntryDecisionActions
                entryId={entry.id}
                decision={entry.decision}
              />
            )
          : entry.decision !== "PENDING"
            ? (
                <Badge variant={decisionBadgeVariant(entry.decision)}>
                  {t(`campaignDetailPage.decisions.${entry.decision.toLowerCase()}`)}
                </Badge>
              )
            : null}
      </div>
    </li>
  );
}
