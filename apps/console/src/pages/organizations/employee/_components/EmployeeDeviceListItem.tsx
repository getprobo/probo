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
import { Badge, Td, Tr } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { EmployeeDeviceListItem_device$key } from "#/__generated__/core/EmployeeDeviceListItem_device.graphql";

const employeeDeviceListItemFragment = graphql`
  fragment EmployeeDeviceListItem_device on Device {
    state
    hostname
    platform
    osVersion
    lastSeenAt
  }
`;

interface EmployeeDeviceListItemProps {
  deviceKey: EmployeeDeviceListItem_device$key;
}

function displayValue(value: string | null | undefined, pendingLabel: string) {
  return value && value.length > 0 ? value : pendingLabel;
}

function stateVariant(state: string): "success" | "warning" | "info" {
  switch (state) {
    case "ACTIVE":
      return "success";
    default:
      return "warning";
  }
}

export function EmployeeDeviceListItem({ deviceKey }: EmployeeDeviceListItemProps) {
  const { __ } = useTranslate();
  const pendingLabel = __("(pending)");

  const device = useFragment(employeeDeviceListItemFragment, deviceKey);

  return (
    <Tr>
      <Td>{displayValue(device.hostname, pendingLabel)}</Td>
      <Td>
        <Badge variant={stateVariant(device.state)}>{device.state}</Badge>
      </Td>
      <Td>{displayValue(device.platform, pendingLabel)}</Td>
      <Td>{displayValue(device.osVersion, pendingLabel)}</Td>
      <Td>{device.lastSeenAt ? formatDate(device.lastSeenAt) : __("Never")}</Td>
    </Tr>
  );
}
