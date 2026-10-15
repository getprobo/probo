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

import { DesktopIcon } from "@phosphor-icons/react";
import { dateFormat, dateTimeFormat } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeviceDetailsCard_deviceFragment$key } from "#/__generated__/core/DeviceDetailsCard_deviceFragment.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";

import { displayValue, stateColor } from "../_lib/deviceDisplay";

import { deviceCardFields } from "./variants";

const deviceFragment = graphql`
  fragment DeviceDetailsCard_deviceFragment on Device {
    state
    hardwareUuid
    serialNumber
    platform
    osVersion
    agentVersion
    enrolledAt
    lastSeenAt
    owner {
      fullName
    }
  }
`;

export function DeviceDetailsCard(props: {
  deviceFragmentRef: DeviceDetailsCard_deviceFragment$key;
}) {
  const { i18n, t } = useTranslation("organizations/devices");
  const pendingLabel = t("devices.values.pending");
  const device = useFragment(deviceFragment, props.deviceFragmentRef);
  const { root, property } = deviceCardFields();
  const tone = stateColor(device.state);

  return (
    <TonedCard
      tone={tone}
      fill={false}
      icon={<DesktopIcon weight="duotone" />}
      lead={(
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("devices.detailsTitle")}
        </Heading>
      )}
    >
      <div className={root()}>
        <div className={property()}>
          <Text size={1} color="faint">{t("devices.fields.state")}</Text>
          <Badge variant="soft" color={tone}>{device.state}</Badge>
        </div>
        <div className={property()}>
          <Text size={1} color="faint">{t("devices.fields.owner")}</Text>
          <Text size={2}>{device.owner?.fullName ?? t("devices.values.unassigned")}</Text>
        </div>
        <div className={property()}>
          <Text size={1} color="faint">{t("devices.fields.hardwareUuid")}</Text>
          <Text size={2}>{displayValue(device.hardwareUuid, pendingLabel)}</Text>
        </div>
        <div className={property()}>
          <Text size={1} color="faint">{t("devices.fields.serialNumber")}</Text>
          <Text size={2}>{displayValue(device.serialNumber, pendingLabel)}</Text>
        </div>
        <div className={property()}>
          <Text size={1} color="faint">{t("devices.fields.platform")}</Text>
          <Text size={2}>{displayValue(device.platform, pendingLabel)}</Text>
        </div>
        <div className={property()}>
          <Text size={1} color="faint">{t("devices.fields.osVersion")}</Text>
          <Text size={2}>{displayValue(device.osVersion, pendingLabel)}</Text>
        </div>
        <div className={property()}>
          <Text size={1} color="faint">{t("devices.fields.agentVersion")}</Text>
          <Text size={2}>{displayValue(device.agentVersion, pendingLabel)}</Text>
        </div>
        <div className={property()}>
          <Text size={1} color="faint">{t("devices.fields.enrolledAt")}</Text>
          <Text size={2}>
            {device.enrolledAt ? dateFormat(i18n.language, device.enrolledAt) : pendingLabel}
          </Text>
        </div>
        <div className={property()}>
          <Text size={1} color="faint">{t("devices.fields.lastSeen")}</Text>
          <Text size={2}>
            {device.lastSeenAt
              ? dateTimeFormat(i18n.language, device.lastSeenAt)
              : t("devices.values.never")}
          </Text>
        </div>
      </div>
    </TonedCard>
  );
}
