// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatDate, formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  DropdownItem,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeviceRowFragment$key } from "#/__generated__/core/DeviceRowFragment.graphql";
import type { DeviceRowRevokeMutation } from "#/__generated__/core/DeviceRowRevokeMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const deviceRowFragment = graphql`
  fragment DeviceRowFragment on Device {
    id
    hostname
    platform
    osVersion
    lastSeenAt
    revokedAt
    latestPostures {
      id
      checkKey
      status
      observedAt
    }
  }
`;

const revokeDeviceMutation = graphql`
  mutation DeviceRowRevokeMutation($input: RevokeDeviceInput!) {
    revokeDevice(input: $input) {
      device {
        id
        revokedAt
      }
    }
  }
`;

interface DeviceRowProps {
  canRevoke: boolean;
  fKey: DeviceRowFragment$key;
}

export function DeviceRow({ canRevoke, fKey }: DeviceRowProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { toast } = useToast();
  const confirm = useConfirm();

  const device = useFragment(deviceRowFragment, fKey);

  const [revokeDevice] = useMutation<DeviceRowRevokeMutation>(
    revokeDeviceMutation,
  );

  const handleRevoke = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          revokeDevice({
            variables: { input: { deviceId: device.id } },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: __("Error"),
                  description: errors[0].message,
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Success"),
                  description: __("Device revoked"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to revoke device"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            "Revoke device \"%s\"? The agent on the device will stop reporting and must be re-enrolled.",
          ),
          device.hostname,
        ),
        variant: "danger",
        label: __("Revoke"),
      },
    );
  };

  const summary = postureSummary(device.latestPostures);
  const isRevoked = Boolean(device.revokedAt);

  return (
    <Tr to={`/organizations/${organizationId}/devices/${device.id}`}>
      <Td>{device.hostname}</Td>
      <Td>{device.platform}</Td>
      <Td>{device.osVersion}</Td>
      <Td>{device.lastSeenAt ? formatDate(device.lastSeenAt) : __("Never")}</Td>
      <Td>
        <span className="text-green-700">{summary.pass}</span>
        {" / "}
        <span className="text-red-700">{summary.fail}</span>
        {" / "}
        <span className="text-tertiary">{summary.unknown}</span>
      </Td>
      <Td noLink width={50} className="text-end">
        {canRevoke && !isRevoked && (
          <ActionDropdown>
            <DropdownItem
              onClick={handleRevoke}
              variant="danger"
              icon={IconTrashCan}
            >
              {__("Revoke")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </Td>
    </Tr>
  );
}

function postureSummary(
  postures: readonly { status: string }[],
): { pass: number; fail: number; unknown: number } {
  const acc = { pass: 0, fail: 0, unknown: 0 };
  for (const p of postures) {
    switch (p.status) {
      case "PASS":
        acc.pass += 1;
        break;
      case "FAIL":
        acc.fail += 1;
        break;
      default:
        acc.unknown += 1;
    }
  }
  return acc;
}
