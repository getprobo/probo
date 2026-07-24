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
import { Button, Card, IconPlusLarge, Tbody, Th, Thead, Tr } from "@probo/ui";
import { useTransition } from "react";
import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
  useRefetchableFragment,
} from "react-relay";

import type { EmployeeDevicesPage_viewer$key } from "#/__generated__/core/EmployeeDevicesPage_viewer.graphql";
import type { EmployeeDevicesPageQuery } from "#/__generated__/core/EmployeeDevicesPageQuery.graphql";
import type { EmployeeDevicesPageRefetchQuery } from "#/__generated__/core/EmployeeDevicesPageRefetchQuery.graphql";

import { CreateDeviceForm } from "./_components/CreateDeviceForm";
import { EmployeeDeviceListItem } from "./_components/EmployeeDeviceListItem";

const employeeDevicesPageViewerFragment = graphql`
  fragment EmployeeDevicesPage_viewer on Viewer
  @refetchable(queryName: "EmployeeDevicesPageRefetchQuery")
  @argumentDefinitions(organizationId: { type: "ID!" }) {
    enrolledDevices(
      organizationId: $organizationId
      first: 100
      orderBy: { field: CREATED_AT, direction: DESC }
    ) {
      edges {
        node {
          id
          ...EmployeeDeviceListItem_device
        }
      }
    }
  }
`;

export const employeeDevicesPageQuery = graphql`
  query EmployeeDevicesPageQuery($organizationId: ID!) {
    viewer @required(action: THROW) {
      ...EmployeeDevicesPage_viewer @arguments(organizationId: $organizationId)
    }
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canEnrollDevice: permission(action: "itam:device:enroll")
      }
    }
  }
`;

interface EmployeeDevicesPageProps {
  queryRef: PreloadedQuery<EmployeeDevicesPageQuery>;
}

export function EmployeeDevicesPage({ queryRef }: EmployeeDevicesPageProps) {
  const { __ } = useTranslate();

  usePageTitle(__("Devices"));

  const { viewer, organization } = usePreloadedQuery<EmployeeDevicesPageQuery>(
    employeeDevicesPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  const [, startTransition] = useTransition();

  const [viewerData, refetchDevices] = useRefetchableFragment<
    EmployeeDevicesPageRefetchQuery,
    EmployeeDevicesPage_viewer$key
  >(employeeDevicesPageViewerFragment, viewer);

  const devices = viewerData.enrolledDevices.edges.map(edge => edge.node);
  const canEnrollDevice = organization.canEnrollDevice ?? false;

  const handleDeviceCreated = () => {
    startTransition(() => {
      refetchDevices({}, { fetchPolicy: "store-and-network" });
    });
  };

  return (
    <div className="space-y-6">
      <header className="flex items-center justify-between gap-4">
        <h1 className="text-2xl font-semibold">{__("Your devices")}</h1>
        {canEnrollDevice && (
          <Button to="/enroll" icon={IconPlusLarge}>
            {__("Enroll new device")}
          </Button>
        )}
      </header>

      <Card>
        {devices.length > 0
          ? (
              <table className="w-full table-fixed">
                <Thead>
                  <Tr>
                    <Th className="text-left">{__("Hostname")}</Th>
                    <Th className="w-32 text-left">{__("State")}</Th>
                    <Th className="w-32 text-left">{__("Platform")}</Th>
                    <Th className="w-40 text-left">{__("OS version")}</Th>
                    <Th className="w-40 text-left">{__("Last seen")}</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {devices.map(device => (
                    <EmployeeDeviceListItem
                      key={device.id}
                      deviceKey={device}
                    />
                  ))}
                </Tbody>
              </table>
            )
          : (
              <div className="px-4 py-12 text-center">
                <h3 className="text-lg font-semibold">
                  {__("No devices enrolled yet")}
                </h3>
              </div>
            )}

      </Card>

      {canEnrollDevice && (
        <CreateDeviceForm onDeviceCreated={handleDeviceCreated} />
      )}
    </div>
  );
}
