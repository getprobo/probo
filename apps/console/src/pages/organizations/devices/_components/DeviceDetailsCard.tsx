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

import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Card } from "@probo/ui";
import type { ReactNode } from "react";
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
  const { __ } = useTranslate();
  const pendingLabel = __("(pending)");
  const device = useFragment(deviceFragment, props.deviceFragmentRef);

  return (
    <Card className="space-y-4" padded>
      <div className="grid grid-cols-3 gap-4">
        <DetailField
          label={__("State")}
          value={
            <Badge variant={stateVariant(device.state)}>{device.state}</Badge>
          }
        />
        <DetailField
          label={__("Owner")}
          value={device.owner?.fullName ?? __("Unassigned")}
        />
        <DetailField
          label={__("Hardware UUID")}
          value={displayValue(device.hardwareUuid, pendingLabel)}
        />
        <DetailField
          label={__("Serial number")}
          value={displayValue(device.serialNumber, pendingLabel)}
        />
        <DetailField
          label={__("Platform")}
          value={displayValue(device.platform, pendingLabel)}
        />
        <DetailField
          label={__("OS version")}
          value={displayValue(device.osVersion, pendingLabel)}
        />
        <DetailField
          label={__("Agent version")}
          value={displayValue(device.agentVersion, pendingLabel)}
        />
        <DetailField
          label={__("Enrolled at")}
          value={
            device.enrolledAt ? formatDate(device.enrolledAt) : pendingLabel
          }
        />
        <DetailField
          label={__("Last seen")}
          value={device.lastSeenAt ? formatDate(device.lastSeenAt) : __("Never")}
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
