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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { PageHeader, Tbody, Th, Thead, Tr } from "@probo/ui";
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

import { DeviceRow } from "./_components/DeviceRow";

export const devicesPageQuery = graphql`
  query DevicesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        canRevokeDevice: permission(action: "core:device:revoke")
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

  usePageTitle(__("Devices"));

  const { organization } = usePreloadedQuery(devicesPageQuery, queryRef);
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
          "Computers running the Probo posture agent. The agent runs as a managed OS service and reports configuration evidence such as disk encryption and screen lock.",
        )}
      />

      <SortableTable
        {...pagination}
        refetch={
          pagination.refetch as ComponentProps<typeof SortableTable>["refetch"]
        }
      >
        <Thead>
          <Tr>
            <SortableTh field="HOSTNAME">{__("Hostname")}</SortableTh>
            <Th>{__("Platform")}</Th>
            <Th>{__("OS version")}</Th>
            <SortableTh field="LAST_SEEN_AT">{__("Last seen")}</SortableTh>
            <Th>{__("Posture")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {devices.map(device => (
            <DeviceRow
              key={device.id}
              fKey={device}
              canRevoke={organization.canRevokeDevice ?? false}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
