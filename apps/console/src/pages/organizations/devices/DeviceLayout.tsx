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
import { Button, IconEject, IconTrashCan, PageHeader } from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { Outlet, useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { DeviceLayoutQuery } from "#/__generated__/core/DeviceLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { DeviceCurrentPostures } from "./_components/DeviceCurrentPostures";
import { DeviceDetailsCard } from "./_components/DeviceDetailsCard";
import { displayValue, isDeviceDeletable } from "./_lib/deviceDisplay";
import { useDeleteDevice } from "./_lib/useDeleteDevice";
import { useRevokeDevice } from "./_lib/useRevokeDevice";

export const deviceLayoutQuery = graphql`
  query DeviceLayoutQuery($deviceId: ID!, $organizationId: ID!) {
    device: node(id: $deviceId) @required(action: THROW) {
      __typename
      ... on Device {
        id
        state
        hostname
        ...DeviceDetailsCard_deviceFragment
        ...DeviceCurrentPostures_deviceFragment
      }
    }
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canRevokeDevice: permission(action: "itam:device:revoke")
        canDeleteDevice: permission(action: "itam:device:delete")
      }
    }
  }
`;

interface DeviceLayoutProps {
  queryRef: PreloadedQuery<DeviceLayoutQuery>;
}

export function DeviceLayout({ queryRef }: DeviceLayoutProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const pendingLabel = t("devices.values.pending");

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
  const [confirmDelete, isDeleting] = useDeleteDevice({
    organizationId,
    onDeleted: () => {
      void navigate(`/organizations/${organizationId}/itam/devices`, { replace: true });
    },
  });

  const deletable = isDeviceDeletable(device.state);
  const canRevokeDevice = organization.canRevokeDevice ?? false;
  const canDeleteDevice = organization.canDeleteDevice ?? false;

  return (
    <div className="flex flex-col gap-6 h-full">
      <PageHeader title={hostnameLabel}>
        {!deletable && canRevokeDevice && (
          <Button
            variant="danger"
            icon={IconEject}
            onClick={() =>
              confirmRevoke({ id: device.id, hostname: device.hostname })}
            disabled={isRevoking}
          >
            {t("devices.actions.revoke")}
          </Button>
        )}
        {deletable && canDeleteDevice && (
          <Button
            variant="danger"
            icon={IconTrashCan}
            onClick={() =>
              confirmDelete({ id: device.id, hostname: device.hostname })}
            disabled={isDeleting}
          >
            {t("devices.actions.delete")}
          </Button>
        )}
      </PageHeader>

      <DeviceDetailsCard deviceFragmentRef={device} />

      <DeviceCurrentPostures deviceFragmentRef={device} />

      <div className="space-y-4">
        <h2 className="text-base font-medium">
          {t("devices.history.title")}
        </h2>
        <Outlet />
      </div>
    </div>
  );
}
