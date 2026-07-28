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

import { useCopy } from "@probo/hooks";
import { dateTimeFormat } from "@probo/i18n";
import {
  IconCheckmark1,
  IconChevronDown,
  IconChevronRight,
  IconSquareBehindSquare2,
  Td,
  Tr,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DevicePostureReportListItemFragment$key } from "#/__generated__/core/DevicePostureReportListItemFragment.graphql";

import { PostureValueBadge } from "../../_components/PostureValueBadge";
import { getPostureCheckLabel } from "../../_lib/getPostureCheckLabel";

const reportFragment = graphql`
  fragment DevicePostureReportListItemFragment on DevicePostureReport {
    id
    createdAt
    postures {
      id
      checkKey
      ...PostureValueBadge_postureFragment
    }
  }
`;

interface DevicePostureReportListItemProps {
  fKey: DevicePostureReportListItemFragment$key;
}

export function DevicePostureReportListItem({
  fKey,
}: DevicePostureReportListItemProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isCopied, copy] = useCopy();
  const { i18n, t } = useTranslation();

  const report = useFragment(reportFragment, fKey);

  return (
    <>
      <Tr
        className="cursor-pointer hover:bg-bg-hover"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <Td className="whitespace-nowrap">
          <div className="flex items-center gap-2">
            {isExpanded
              ? (
                  <IconChevronDown size={16} className="text-txt-secondary" />
                )
              : (
                  <IconChevronRight size={16} className="text-txt-secondary" />
                )}
            {dateTimeFormat(i18n.language, report.createdAt)}
          </div>
        </Td>
        <Td>
          <div className="flex items-center gap-1">
            <code className="font-mono text-xs text-txt-tertiary">
              {report.id}
            </code>
            <button
              type="button"
              className="p-1 rounded text-txt-tertiary hover:bg-bg-hover hover:text-txt-secondary transition-colors cursor-pointer"
              aria-label={t("devices.history.actions.copyCorrelationId")}
              title={
                isCopied
                  ? t("devices.history.actions.correlationIdCopied")
                  : t("devices.history.actions.copyCorrelationId")
              }
              onClick={(event) => {
                event.stopPropagation();
                copy(report.id);
              }}
            >
              {isCopied
                ? <IconCheckmark1 size={14} />
                : <IconSquareBehindSquare2 size={14} />}
            </button>
          </div>
        </Td>
        <Td>
          {t("devices.history.checkCount", {
            count: report.postures.length,
          })}
        </Td>
      </Tr>
      {isExpanded && (
        <Tr>
          <Td colSpan={3} className="bg-bg-subtle">
            <div className="py-2 pl-6 grid grid-cols-2 gap-x-8 gap-y-2">
              {report.postures.map(posture => (
                <div key={posture.id} className="flex items-center gap-4 text-sm">
                  <span className="text-txt-secondary min-w-40">
                    {getPostureCheckLabel(t, posture.checkKey)}
                  </span>
                  <PostureValueBadge postureFragmentRef={posture} />
                </div>
              ))}
            </div>
          </Td>
        </Tr>
      )}
    </>
  );
}
