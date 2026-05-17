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

import { formatDate } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Button,
  DropdownItem,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import {
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";

import type { DeviceGraphListQuery } from "#/__generated__/core/DeviceGraphListQuery.graphql";
import type {
  DeviceGraphPaginatedFragment$data,
  DeviceGraphPaginatedFragment$key,
} from "#/__generated__/core/DeviceGraphPaginatedFragment.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import {
  devicesQuery,
  paginatedDevicesFragment,
  useRevokeDevice,
} from "#/hooks/graph/DeviceGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

type Device = NodeOf<DeviceGraphPaginatedFragment$data["devices"]>;

type Props = {
  queryRef: PreloadedQuery<DeviceGraphListQuery>;
};

export default function DevicesPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const data = usePreloadedQuery(devicesQuery, props.queryRef);
  // eslint-disable-next-line relay/generated-typescript-types
  const pagination = usePaginationFragment(
    paginatedDevicesFragment,
    data.node as DeviceGraphPaginatedFragment$key,
  );

  const devices = pagination.data.devices?.edges.map(edge => edge.node);

  usePageTitle(__("Devices"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Devices")}
        description={__(
          "Computers running the Probo posture agent. The agent runs as a managed OS service and reports configuration evidence such as disk encryption and screen lock.",
        )}
      >
        {data.node?.canCreateEnrollmentToken && (
          <Button
            icon={IconPlusLarge}
            onClick={() => void navigate(
              `/organizations/${organizationId}/devices/enroll`,
            )}
          >
            {__("Enroll device")}
          </Button>
        )}
      </PageHeader>
      <SortableTable {...pagination}>
        <Thead>
          <Tr>
            <SortableTh field="HOSTNAME">{__("Hostname")}</SortableTh>
            <Th>{__("Platform")}</Th>
            <Th>{__("OS version")}</Th>
            <Th>{__("Last seen")}</Th>
            <Th>{__("Posture")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {devices?.map(device => (
            <DeviceRow
              key={device.id}
              device={device}
              organizationId={organizationId}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

function DeviceRow({
  device,
  organizationId,
}: {
  device: Device;
  organizationId: string;
}) {
  const { __ } = useTranslate();
  const revoke = useRevokeDevice(device);

  const url = `/organizations/${organizationId}/devices/${device.id}`;
  const summary = postureSummary(device.latestPostures);

  return (
    <Tr to={url}>
      <Td>{device.hostname}</Td>
      <Td>{device.platform}</Td>
      <Td>{device.osVersion}</Td>
      <Td>
        {device.lastSeenAt ? formatDate(device.lastSeenAt) : __("Never")}
      </Td>
      <Td>
        <span className="text-green-700">{summary.pass}</span>
        {" / "}
        <span className="text-red-700">{summary.fail}</span>
        {" / "}
        <span className="text-tertiary">{summary.unknown}</span>
      </Td>
      <Td noLink width={50} className="text-end">
        <ActionDropdown>
          <DropdownItem
            onClick={revoke}
            variant="danger"
            icon={IconTrashCan}
          >
            {__("Revoke")}
          </DropdownItem>
        </ActionDropdown>
      </Td>
    </Tr>
  );
}

function postureSummary(
  postures?: readonly { status: string }[] | null,
): { pass: number; fail: number; unknown: number } {
  const acc = { pass: 0, fail: 0, unknown: 0 };
  for (const p of postures ?? []) {
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
