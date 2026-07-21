// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
import { IconCircleCheck, IconRadioUnchecked } from "@probo/ui";
import { clsx } from "clsx";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { VersionRowFragment$key } from "#/__generated__/core/VersionRowFragment.graphql";

const fragment = graphql`
  fragment VersionRowFragment on EmployeeDocumentVersion {
    # eslint-disable-next-line relay/unused-fields
    id
    major
    minor
    signed
    publishedAt
  }
`;

export function VersionRow({
  fKey,
  isSelected,
  onSelect,
}: {
  fKey: VersionRowFragment$key;
  isSelected: boolean;
  onSelect: () => void;
}) {
  const { t, i18n } = useTranslation();
  const versionData = useFragment<VersionRowFragment$key>(fragment, fKey);
  const isVersionSigned = versionData.signed;

  return (
    <div
      onClick={onSelect}
      className={clsx(
        "flex items-center gap-3 py-3 px-4 transition-colors cursor-pointer",
        isSelected
          ? "bg-blue-50 border-l-4 border-blue-500"
          : "bg-transparent hover:bg-level-1",
      )}
    >
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-level-2 flex-shrink-0">
        {isVersionSigned
          ? (
              <IconCircleCheck size={20} className="text-txt-success" />
            )
          : (
              <IconRadioUnchecked size={20} className="text-txt-tertiary" />
            )}
      </div>
      <div className="flex-1 min-w-0">
        <p
          className={clsx(
            "text-sm font-medium truncate",
            isVersionSigned ? "text-txt-tertiary" : "text-txt-primary",
          )}
        >
          {versionData.publishedAt
            ? t("versionRow.versionWithDate", {
                major: versionData.major,
                minor: versionData.minor,
                date: dateFormat(i18n.language, versionData.publishedAt),
              })
            : t("versionRow.version", {
                major: versionData.major,
                minor: versionData.minor,
              })}
        </p>
      </div>
      <div className="flex-shrink-0">
        <span
          className={clsx(
            "inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium",
            isVersionSigned
              ? "bg-green-100 text-green-800"
              : isSelected
                ? "bg-blue-100 text-blue-800"
                : "bg-gray-100 text-gray-700",
          )}
        >
          {isVersionSigned
            ? t("versionRow.status.signed")
            : isSelected
              ? t("versionRow.status.inReview")
              : t("versionRow.status.waitingSignature")}
        </span>
      </div>
    </div>
  );
}
