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

import { relativeDateFormat } from "@probo/i18n";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DeviceListItem_device$key } from "#/__generated__/core/DeviceListItem_device.graphql";
import {
  formatDeviceOs,
  isDeviceConnected,
} from "#/pages/devices/_lib/deviceDisplay";

import { deviceListItem } from "./variants";

const deviceListItemFragment = graphql`
  fragment DeviceListItem_device on Device @throwOnFieldError {
    hostname
    platform
    osVersion
    lastSeenAt
    state
  }
`;

export interface DeviceListItemProps {
  deviceKey: DeviceListItem_device$key;
}

export function DeviceListItem({ deviceKey }: DeviceListItemProps) {
  const { t, i18n } = useTranslation("devices");
  const device = useFragment(deviceListItemFragment, deviceKey);
  const connected = isDeviceConnected(device.state);
  const slots = deviceListItem({ connected });
  const platformLabel = device.platform === undefined || device.platform === null
    ? null
    : t(`list.platforms.${device.platform}`);
  const os = formatDeviceOs(platformLabel, device.osVersion);
  const lastActive = device.lastSeenAt === undefined || device.lastSeenAt === null
    ? t("list.never")
    : relativeDateFormat(i18n.language, device.lastSeenAt);
  const hostname = device.hostname === undefined
    || device.hostname === null
    || device.hostname === ""
    ? t("list.pendingHostname")
    : device.hostname;

  return (
    <TableRow align="center">
      <TableRowHeaderCell className={slots.cell()} colSpan={2} style={{ padding: 0 }}>
        <div className={slots.row()}>
          <Text size={2} weight="medium" highContrast className={slots.title()}>
            {hostname}
          </Text>
          <div className={slots.meta()}>
            <span className={slots.timestamp()}>
              <Text size={1} color="current" className={slots.timestampLabel()}>
                {t("list.lastActive")}
              </Text>
              <Text size={1} color="current" className={slots.timestampValue()}>
                {lastActive}
              </Text>
            </span>
            {os === null
              ? null
              : (
                  <Text size={1} color="current" className={slots.os()}>
                    {os}
                  </Text>
                )}
            <span className={slots.status()}>
              <span className={slots.pipWrap()}>
                <span className={slots.pip()} />
              </span>
              <Text size={1} color="current" className={slots.statusLabel()}>
                {connected ? t("list.connected") : t("list.disconnected")}
              </Text>
            </span>
          </div>
        </div>
      </TableRowHeaderCell>
    </TableRow>
  );
}
