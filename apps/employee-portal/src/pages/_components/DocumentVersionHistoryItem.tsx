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

import { parseDate } from "@probo/helpers";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentVersionHistoryItem_version$key } from "#/__generated__/core/DocumentVersionHistoryItem_version.graphql";
import type { DocumentVersionHistoryKind } from "#/pages/_lib/documentVersion";

import { documentVersionHistoryItem } from "./variants";

const documentVersionHistoryItemFragment = graphql`
  fragment DocumentVersionHistoryItem_version on EmployeeDocumentVersion {
    major
    minor
    signed
    publishedAt
    createdAt
    approvalDecision {
      state
    }
  }
`;

export interface DocumentVersionHistoryItemProps {
  versionKey: DocumentVersionHistoryItem_version$key;
  kind: DocumentVersionHistoryKind;
  selected: boolean;
  current: boolean;
  onSelect: () => void;
}

export function DocumentVersionHistoryItem({
  versionKey,
  kind,
  selected,
  current,
  onSelect,
}: DocumentVersionHistoryItemProps) {
  const { t, i18n } = useTranslation();
  const slots = documentVersionHistoryItem();
  const version = useFragment(documentVersionHistoryItemFragment, versionKey);
  const dated = version.publishedAt ?? version.createdAt;
  const date = new Intl.DateTimeFormat(i18n.language, {
    dateStyle: "medium",
  }).format(parseDate(dated));

  return (
    <button
      type="button"
      className={slots.row()}
      aria-pressed={selected}
      aria-label={t("documents.versions.select", {
        major: version.major,
        minor: version.minor,
      })}
      onClick={onSelect}
    >
      <span className={slots.radio()} aria-hidden>
        {selected ? <span className={slots.radioDot()} /> : null}
      </span>
      <span className={slots.copy()}>
        <Text size={2} weight="medium" highContrast>
          {t("documents.versions.label", {
            major: version.major,
            minor: version.minor,
          })}
        </Text>
        <span className={slots.meta()}>
          <Text size={1} color="faint">
            {date}
          </Text>
          <Text size={1} color="faint">
            ·
          </Text>
          <Text size={1} color="faint">
            {versionStatusLabel(kind, version.signed, version.approvalDecision?.state, t)}
          </Text>
        </span>
      </span>
      {current
        ? (
            <Text size={2} color="faint" className={slots.current()}>
              {t("documents.versions.current")}
            </Text>
          )
        : null}
    </button>
  );
}

function versionStatusLabel(
  kind: DocumentVersionHistoryKind,
  signed: boolean,
  approvalState: string | null | undefined,
  t: (key: string) => string,
): string {
  if (kind === "signatures") {
    return signed
      ? t("documents.versions.status.signed")
      : t("documents.versions.status.inReview");
  }
  if (approvalState === "APPROVED") {
    return t("documents.versions.status.approved");
  }
  if (approvalState === "REJECTED") {
    return t("documents.versions.status.rejected");
  }
  return t("documents.versions.status.inReview");
}
