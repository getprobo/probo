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
import { Badge, Card } from "@probo/ui";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeviceDetailsCard_deviceFragment$key } from "#/__generated__/core/DeviceDetailsCard_deviceFragment.graphql";

import { displayValue, stateVariant } from "../_lib/deviceDisplay";

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
  const { i18n, t } = useTranslation();
  const pendingLabel = t("devices.values.pending");
  const device = useFragment(deviceFragment, props.deviceFragmentRef);

  return (
    <Card className="space-y-4" padded>
      <div className="grid grid-cols-3 gap-4">
        <DetailField
          label={t("devices.fields.state")}
          value={
            <Badge variant={stateVariant(device.state)}>{device.state}</Badge>
          }
        />
        <DetailField
          label={t("devices.fields.owner")}
          value={device.owner?.fullName ?? t("devices.values.unassigned")}
        />
        <DetailField
          label={t("devices.fields.hardwareUuid")}
          value={displayValue(device.hardwareUuid, pendingLabel)}
        />
        <DetailField
          label={t("devices.fields.serialNumber")}
          value={displayValue(device.serialNumber, pendingLabel)}
        />
        <DetailField
          label={t("devices.fields.platform")}
          value={displayValue(device.platform, pendingLabel)}
        />
        <DetailField
          label={t("devices.fields.osVersion")}
          value={displayValue(device.osVersion, pendingLabel)}
        />
        <DetailField
          label={t("devices.fields.agentVersion")}
          value={displayValue(device.agentVersion, pendingLabel)}
        />
        <DetailField
          label={t("devices.fields.enrolledAt")}
          value={
            device.enrolledAt
              ? dateFormat(i18n.language, device.enrolledAt)
              : pendingLabel
          }
        />
        <DetailField
          label={t("devices.fields.lastSeen")}
          value={
            device.lastSeenAt
              ? dateFormat(i18n.language, device.lastSeenAt)
              : t("devices.values.never")
          }
        />
      </div>
    </Card>
  );
}

function DetailField(props: { label: string; value: ReactNode }) {
  return (
    <div>
      <div className="text-xs text-txt-tertiary font-semibold mb-1">
        {props.label}
      </div>
      <div className="text-sm text-txt-primary">{props.value}</div>
    </div>
  );
}
