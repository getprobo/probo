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

import { CaretDownIcon, CaretRightIcon, CopyIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DevicePostureReportListItemFragment$key } from "#/__generated__/core/DevicePostureReportListItemFragment.graphql";

import { PostureValueBadge } from "../../_components/PostureValueBadge";
import { postureReport } from "../../_components/variants";
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
  const { i18n, t } = useTranslation("organizations/devices");
  const toast = useToast();
  const report = useFragment(reportFragment, fKey);
  const { time, correlation, checks, check, checkLabel } = postureReport();

  async function copyCorrelationId() {
    try {
      await navigator.clipboard.writeText(report.id);
    } catch {
      toast.add({
        title: t("common.error"),
        description: t("devices.history.errors.copyCorrelationId"),
        type: "error",
      });
      return;
    }
    toast.add({
      title: t("devices.history.actions.correlationIdCopied"),
      type: "success",
    });
  }

  return (
    <>
      <TableRow align="center">
        <TableCell>
          <button
            type="button"
            className={time()}
            aria-expanded={isExpanded}
            onClick={() => setIsExpanded(!isExpanded)}
          >
            {isExpanded ? <CaretDownIcon /> : <CaretRightIcon />}
            {dateTimeFormat(i18n.language, report.createdAt)}
          </button>
        </TableCell>
        <TableCell>
          <div className={correlation()}>
            <Code size={1}>{report.id}</Code>
            <IconButton
              variant="ghost"
              color="neutral"
              size={1}
              aria-label={t("devices.history.actions.copyCorrelationId")}
              onClick={() => {
                void copyCorrelationId();
              }}
            >
              <CopyIcon />
            </IconButton>
          </div>
        </TableCell>
        <TableCell>
          {t("devices.history.checkCount", { count: report.postures.length })}
        </TableCell>
      </TableRow>
      {isExpanded && (
        <TableRow>
          <TableCell colSpan={3}>
            <div className={checks()}>
              {report.postures.map(posture => (
                <div key={posture.id} className={check()}>
                  <Text size={2} color="faint" className={checkLabel()}>
                    {getPostureCheckLabel(t, posture.checkKey)}
                  </Text>
                  <PostureValueBadge postureFragmentRef={posture} />
                </div>
              ))}
            </div>
          </TableCell>
        </TableRow>
      )}
    </>
  );
}
