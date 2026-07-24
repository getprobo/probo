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
import {
  ActionDropdown,
  Badge,
  DropdownItem,
  IconTrashCan,
  IconUser,
  Td,
  Tr,
  useDialogRef,
} from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeviceRowFragment$key } from "#/__generated__/core/DeviceRowFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { displayValue, stateVariant } from "../_lib/deviceDisplay";
import { useRevokeDevice } from "../_lib/useRevokeDevice";
import { ReassignDeviceDialog } from "../dialogs/ReassignDeviceDialog";

const deviceRowFragment = graphql`
  fragment DeviceRowFragment on Device {
    id
    state
    hostname
    platform
    osVersion
    lastSeenAt
    owner {
      id
      fullName
    }
    latestPostures {
      id
      status
    }
    ...ReassignDeviceDialog_device
  }
`;

interface DeviceRowProps {
  canAssignDevice: boolean;
  canRevoke: boolean;
  fKey: DeviceRowFragment$key;
}

export function DeviceRow({ canAssignDevice, canRevoke, fKey }: DeviceRowProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const reassignDialogRef = useDialogRef();
  const pendingLabel = __("(pending)");

  const device = useFragment(deviceRowFragment, fKey);

  const [confirmRevoke, isRevoking] = useRevokeDevice();

  const summary = postureSummary(device.latestPostures);
  const isRevoked = device.state === "REVOKED";
  const hasActions = !isRevoked && (canRevoke || canAssignDevice);

  return (
    <>
      <ReassignDeviceDialog
        ref={reassignDialogRef}
        deviceKey={device}
        organizationId={organizationId}
      />
      <Tr to={`/organizations/${organizationId}/devices/${device.id}`}>
        <Td>{displayValue(device.hostname, pendingLabel)}</Td>
        <Td>{device.owner?.fullName ?? __("Unassigned")}</Td>
        <Td>
          <Badge variant={stateVariant(device.state)}>{device.state}</Badge>
        </Td>
        <Td>{displayValue(device.platform, pendingLabel)}</Td>
        <Td>{displayValue(device.osVersion, pendingLabel)}</Td>
        <Td>{device.lastSeenAt ? formatDate(device.lastSeenAt) : __("Never")}</Td>
        <Td>
          <span className="text-txt-success">{summary.pass}</span>
          {" / "}
          <span className={summary.fail > 0 ? "text-txt-danger" : undefined}>
            {summary.fail}
          </span>
          {" / "}
          {summary.total}
        </Td>
        <Td noLink width={50} className="text-end">
          {hasActions && (
            <ActionDropdown>
              {canAssignDevice && (
                <DropdownItem
                  icon={IconUser}
                  onClick={() => reassignDialogRef.current?.open()}
                >
                  {__("Re-assign")}
                </DropdownItem>
              )}
              {canRevoke && (
                <DropdownItem
                  onClick={() =>
                    confirmRevoke({ id: device.id, hostname: device.hostname })}
                  disabled={isRevoking}
                  variant="danger"
                  icon={IconTrashCan}
                >
                  {__("Revoke")}
                </DropdownItem>
              )}
            </ActionDropdown>
          )}
        </Td>
      </Tr>
    </>
  );
}

function postureSummary(
  postures: readonly { status: string }[],
): { pass: number; fail: number; total: number } {
  let pass = 0;
  let fail = 0;
  for (const p of postures) {
    switch (p.status) {
      case "PASS":
        pass += 1;
        break;
      case "FAIL":
        fail += 1;
        break;
      default:
        // UNKNOWN and NOT_APPLICABLE count toward total only.
        break;
    }
  }
  return { pass, fail, total: postures.length };
}
