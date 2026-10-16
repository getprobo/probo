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

import { ChecksIcon, DotsThreeIcon, PlugsIcon, QuestionIcon, TrashIcon, UserIcon, WarningIcon, XIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableLink } from "@probo/ui/src/v2/Table/TableLink";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeviceListItem_device$key } from "#/__generated__/core/DeviceListItem_device.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  displayValue,
  isDeviceDeletable,
  ownerLabel,
  platformLabel,
  postureStatusOutcome,
  stateColor,
} from "../_lib/deviceDisplay";
import { getPostureCheckLabel } from "../_lib/getPostureCheckLabel";
import { ReassignDeviceDialog } from "../dialogs/ReassignDeviceDialog";

import { DeviceDeleteDialog } from "./DeviceDeleteDialog";
import { DeviceRevokeDialog } from "./DeviceRevokeDialog";
import { postureChecksPopover } from "./variants";

const deviceListItemFragment = graphql`
  fragment DeviceListItem_device on Device {
    id
    state
    platform
    osVersion
    agentVersion
    lastSeenAt
    owner {
      id
      fullName
    }
    latestPostures {
      checkKey
      status
    }
    ...ReassignDeviceDialog_device
  }
`;

interface DeviceListItemProps {
  canAssignDevice: boolean;
  canDelete: boolean;
  canRevoke: boolean;
  connectionId: string;
  fKey: DeviceListItem_device$key;
  selected?: boolean;
  onSelectedChange?: (selected: boolean) => void;
  onRevoked?: () => void;
}

export function DeviceListItem({
  canAssignDevice,
  canDelete,
  canRevoke,
  connectionId,
  fKey,
  selected,
  onSelectedChange,
  onRevoked,
}: DeviceListItemProps) {
  const { i18n, t } = useTranslation("organizations/devices");
  const organizationId = useOrganizationId();
  const [reassignOpen, setReassignOpen] = useState(false);
  const [revokeOpen, setRevokeOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const pendingLabel = t("devices.values.pending");
  const device = useFragment(deviceListItemFragment, fKey);
  const name = ownerLabel(device.owner?.fullName, t("devices.values.unassigned"));
  const passes = device.latestPostures.filter(
    posture => posture.status === "PASS",
  ).length;
  const warnings: { key: string; label: string; outcome: "warning" | "unknown" }[] = [];
  for (const posture of device.latestPostures) {
    const outcome = postureStatusOutcome(posture.status);
    if (outcome === "pass" || outcome === "notApplicable") {
      continue;
    }
    warnings.push({
      key: posture.checkKey,
      label: getPostureCheckLabel(t, posture.checkKey),
      outcome,
    });
  }
  const platform = platformLabel(t, device.platform) ?? pendingLabel;
  const deletable = isDeviceDeletable(device.state);
  const showAssign = canAssignDevice && !deletable;
  const showRevoke = canRevoke && !deletable;
  const showDelete = canDelete && deletable;
  const hasActions = showAssign || showRevoke || showDelete;
  const { trigger, popup, list, item } = postureChecksPopover();

  return (
    <>
      <ReassignDeviceDialog
        open={reassignOpen}
        onOpenChange={setReassignOpen}
        deviceKey={device}
        organizationId={organizationId}
      />
      <DeviceRevokeDialog
        open={revokeOpen}
        onOpenChange={setRevokeOpen}
        devices={[{ id: device.id, ownerName: device.owner?.fullName }]}
        onRevoked={onRevoked}
      />
      <DeviceDeleteDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        organizationId={organizationId}
        connectionId={connectionId}
        devices={[{ id: device.id, ownerName: device.owner?.fullName }]}
      />
      <TableRow align="center" interactive>
        {onSelectedChange && (
          <TableCell interactive width="3rem">
            <Checkbox
              checked={selected ?? false}
              onCheckedChange={(checked) => {
                if (checked === true || checked === false) {
                  onSelectedChange(checked);
                }
              }}
              aria-label={t("devices.actions.selectDevice", { owner: name })}
            />
          </TableCell>
        )}
        <TableRowHeaderCell>
          <TableLink to={`/organizations/${organizationId}/itam/devices/${device.id}`}>
            {name}
          </TableLink>
        </TableRowHeaderCell>
        <TableCell>
          <Badge variant="soft" color={stateColor(device.state)}>{device.state}</Badge>
        </TableCell>
        <TableCell>{platform}</TableCell>
        <TableCell>{displayValue(device.osVersion, pendingLabel)}</TableCell>
        <TableCell>{displayValue(device.agentVersion, pendingLabel)}</TableCell>
        <TableCell>
          {device.lastSeenAt
            ? dateTimeFormat(i18n.language, device.lastSeenAt)
            : t("devices.values.never")}
        </TableCell>
        <TableCell interactive>
          {passes > 0 && warnings.length === 0 && (
            <Badge size={1} variant="soft" color="green" iconStart={<ChecksIcon />}>
              {t("devices.postures.checks", { count: passes })}
            </Badge>
          )}
          {warnings.length > 0 && (
            <Popover>
              <PopoverTrigger
                render={(
                  <button
                    type="button"
                    className={trigger()}
                    aria-label={t("devices.postures.showWarnings", { count: warnings.length })}
                  />
                )}
              >
                <Badge size={1} variant="soft" color="amber" iconStart={<WarningIcon />}>
                  {t("devices.postures.warnings", { count: warnings.length })}
                </Badge>
              </PopoverTrigger>
              <PopoverPopup className={popup()} align="end">
                <ul className={list()}>
                  {warnings.map(warning => (
                    <li key={warning.key} className={item()}>
                      {warning.outcome === "unknown"
                        ? <QuestionIcon className="size-3" aria-hidden />
                        : <XIcon className="size-3 text-red-11" aria-hidden />}
                      {warning.label}
                    </li>
                  ))}
                </ul>
              </PopoverPopup>
            </Popover>
          )}
        </TableCell>
        <TableCell interactive justify="end" width="3rem">
          {hasActions && (
            <Dropdown>
              <DropdownTrigger
                render={(
                  <IconButton variant="ghost" color="neutral" aria-label={t("devices.actions.menu")}>
                    <DotsThreeIcon />
                  </IconButton>
                )}
              />
              <DropdownPopup>
                {showAssign && (
                  <DropdownItem iconStart={<UserIcon />} onClick={() => setReassignOpen(true)}>
                    {t("devices.actions.reassign")}
                  </DropdownItem>
                )}
                {showRevoke && (
                  <DropdownItem color="error" iconStart={<PlugsIcon />} onClick={() => setRevokeOpen(true)}>
                    {t("devices.actions.revoke")}
                  </DropdownItem>
                )}
                {showDelete && (
                  <DropdownItem
                    color="error"
                    iconStart={<TrashIcon />}
                    onClick={() => setDeleteOpen(true)}
                  >
                    {t("devices.actions.delete")}
                  </DropdownItem>
                )}
              </DropdownPopup>
            </Dropdown>
          )}
        </TableCell>
      </TableRow>
    </>
  );
}
