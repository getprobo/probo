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

import { ChecksIcon, QuestionIcon, WarningIcon, XIcon } from "@phosphor-icons/react";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DeviceCardListItem_device$key } from "#/__generated__/core/DeviceCardListItem_device.graphql";
import {
  getPostureCheckLabel,
  postureStatusOutcome,
} from "#/pages/_lib/postureChecks";

import { devicePostureChecks } from "./variants";

const deviceCardListItemFragment = graphql`
  fragment DeviceCardListItem_device on Device @throwOnFieldError {
    hostname
    latestPostures {
      checkKey
      status
    }
  }
`;

export interface DeviceCardListItemProps {
  deviceKey: DeviceCardListItem_device$key;
  showHostname: boolean;
}

export function DeviceCardListItem({
  deviceKey,
  showHostname,
}: DeviceCardListItemProps) {
  const { t } = useTranslation();
  const device = useFragment(deviceCardListItemFragment, deviceKey);
  const slots = devicePostureChecks();
  const warnings: {
    key: string;
    label: string;
    outcome: "warning" | "unknown";
  }[] = [];
  let passCount = 0;
  for (const posture of device.latestPostures) {
    const outcome = postureStatusOutcome(posture.status);
    if (outcome === "pass") {
      passCount += 1;
      continue;
    }
    if (outcome === "notApplicable") {
      continue;
    }
    warnings.push({
      key: posture.checkKey,
      label: getPostureCheckLabel(t, posture.checkKey),
      outcome,
    });
  }
  const hostname = device.hostname === undefined
    || device.hostname === null
    || device.hostname === ""
    ? t("homePage.dashboard.devices.pendingHostname")
    : device.hostname;

  return (
    <div className={slots.root()}>
      {showHostname
        ? (
            <Text size={1} weight="medium" highContrast className={slots.hostname()}>
              {hostname}
            </Text>
          )
        : null}
      {warnings.length > 0
        ? (
            <>
              <Badge size={1} variant="soft" color="amber" iconStart={<WarningIcon />}>
                {t("homePage.dashboard.devices.warnings", { count: warnings.length })}
              </Badge>
              <ul className={slots.list()}>
                {warnings.map(warning => (
                  <li key={warning.key} className={slots.item()}>
                    {warning.outcome === "unknown"
                      ? <QuestionIcon className={slots.unknownIcon()} aria-hidden />
                      : <XIcon className={slots.failIcon()} aria-hidden />}
                    <Text size={1}>
                      {warning.label}
                    </Text>
                  </li>
                ))}
              </ul>
            </>
          )
        : passCount > 0
          ? (
              <Badge size={1} variant="soft" color="green" iconStart={<ChecksIcon />}>
                {t("homePage.dashboard.devices.checks", { count: passCount })}
              </Badge>
            )
          : (
              <Text size={2} weight="medium" className={slots.empty()}>
                {t("homePage.dashboard.devices.emptyPostures")}
              </Text>
            )}
    </div>
  );
}
