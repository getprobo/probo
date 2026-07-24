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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge, PageHeader, Tbody, Th, Thead, Tr } from "@probo/ui";
import type { ComponentProps } from "react";
import {
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { DevicesPageFragment$key } from "#/__generated__/core/DevicesPageFragment.graphql";
import type { DevicesPageFragment_RefetchQuery } from "#/__generated__/core/DevicesPageFragment_RefetchQuery.graphql";
import type { DevicesPageQuery } from "#/__generated__/core/DevicesPageQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { DeviceRow } from "./_components/DeviceRow";
import { PostureColumnHeader } from "./_components/PostureColumnHeader";
import { CreateDeviceDialog } from "./dialogs/CreateDeviceDialog";

export const devicesPageQuery = graphql`
  query DevicesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        canAssignDevice: permission(action: "itam:device:assign")
        canRevokeDevice: permission(action: "itam:device:revoke")
        canCreateDevice: permission(action: "itam:device:create")
        ...DevicesPageFragment
      }
    }
  }
`;

const devicesPageFragment = graphql`
  fragment DevicesPageFragment on Organization
  @refetchable(queryName: "DevicesPageFragment_RefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "DeviceOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    devices(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "DevicesPage_devices", filters: ["orderBy"]) {
      edges {
        node {
          id
          ...DeviceRowFragment
        }
      }
    }
  }
`;

interface DevicesPageProps {
  queryRef: PreloadedQuery<DevicesPageQuery>;
}

export function DevicesPage({ queryRef }: DevicesPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  usePageTitle(__("Devices"));

  const { organization } = usePreloadedQuery<DevicesPageQuery>(
    devicesPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  const pagination = usePaginationFragment<
    DevicesPageFragment_RefetchQuery,
    DevicesPageFragment$key
  >(devicesPageFragment, organization);

  const devices = pagination.data.devices.edges.map(edge => edge.node);

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Devices")}
        description={__(
          "Manage computers enrolled with the Probo posture agent.",
        )}
      >
        {organization.canCreateDevice && (
          <CreateDeviceDialog
            organizationId={organizationId}
            onCreated={() => {
              pagination.refetch({}, { fetchPolicy: "store-and-network" });
            }}
          >
            <Button icon={IconPlusLarge}>{__("Add device")}</Button>
          </CreateDeviceDialog>
        )}
      </PageHeader>

      <SortableTable
        {...pagination}
        refetch={
          pagination.refetch as ComponentProps<typeof SortableTable>["refetch"]
        }
      >
        <Thead>
          <Tr>
            <SortableTh field="HOSTNAME">{__("Hostname")}</SortableTh>
            <Th>{__("Owner")}</Th>
            <Th>{__("State")}</Th>
            <Th>{__("Platform")}</Th>
            <Th>{__("OS version")}</Th>
            <SortableTh field="LAST_SEEN_AT">{__("Last seen")}</SortableTh>
            <Th>
              <PostureColumnHeader />
            </Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {devices.map(device => (
            <DeviceRow
              key={device.id}
              fKey={device}
              canAssignDevice={organization.canAssignDevice ?? false}
              canRevoke={organization.canRevokeDevice ?? false}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
