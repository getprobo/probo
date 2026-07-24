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
import {
  Breadcrumb,
  Button,
  PageHeader,
  TabLink,
  Tabs,
} from "@probo/ui";
import {
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { DeviceLayoutQuery } from "#/__generated__/core/DeviceLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { DeviceDetailsCard } from "./_components/DeviceDetailsCard";
import { displayValue } from "./_lib/deviceDisplay";
import { useRevokeDevice } from "./_lib/useRevokeDevice";

export const deviceLayoutQuery = graphql`
  query DeviceLayoutQuery($deviceId: ID!, $organizationId: ID!) {
    device: node(id: $deviceId) @required(action: THROW) {
      __typename
      ... on Device {
        id
        state
        hostname
        platform
        ...DeviceDetailsCard_deviceFragment
      }
    }
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canRevokeDevice: permission(action: "itam:device:revoke")
      }
    }
  }
`;

interface DeviceLayoutProps {
  queryRef: PreloadedQuery<DeviceLayoutQuery>;
}

export function DeviceLayout({ queryRef }: DeviceLayoutProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const pendingLabel = __("(pending)");

  const { device, organization } = usePreloadedQuery<DeviceLayoutQuery>(
    deviceLayoutQuery,
    queryRef,
  );
  if (device.__typename !== "Device") {
    throw new Error("invalid type for device node");
  }
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  usePageTitle(displayValue(device.hostname, pendingLabel));

  const hostnameLabel = displayValue(device.hostname, pendingLabel);

  const [confirmRevoke, isRevoking] = useRevokeDevice();

  const isRevoked = device.state === "REVOKED";
  const canRevokeDevice = organization.canRevokeDevice ?? false;

  return (
    <div className="flex flex-col gap-6 h-full">
      <Breadcrumb
        items={[
          {
            label: __("Devices"),
            to: `/organizations/${organizationId}/devices`,
          },
          { label: hostnameLabel },
        ]}
      />
      <PageHeader
        title={hostnameLabel}
        description={displayValue(device.platform, pendingLabel)}
      >
        {!isRevoked && canRevokeDevice && (
          <Button
            variant="danger"
            onClick={() =>
              confirmRevoke({ id: device.id, hostname: device.hostname })}
            disabled={isRevoking}
          >
            {__("Revoke")}
          </Button>
        )}
      </PageHeader>

      <DeviceDetailsCard deviceFragmentRef={device} />

      <Tabs>
        <TabLink
          to={`/organizations/${organizationId}/devices/${device.id}/postures`}
        >
          {__("Postures")}
        </TabLink>
      </Tabs>

      <Outlet />
    </div>
  );
}
