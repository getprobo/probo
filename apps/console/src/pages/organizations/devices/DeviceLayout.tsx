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

import { CaretLeftIcon, DownloadSimpleIcon, PlugsIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet, useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { DeviceLayoutQuery } from "#/__generated__/core/DeviceLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { DeviceCurrentPostures } from "./_components/DeviceCurrentPostures";
import { DeviceDeleteDialog } from "./_components/DeviceDeleteDialog";
import { DeviceDetailsCard } from "./_components/DeviceDetailsCard";
import { DeviceRevokeDialog } from "./_components/DeviceRevokeDialog";
import { devicesPage } from "./_components/variants";
import { isDeviceDeletable, ownerLabel } from "./_lib/deviceDisplay";
import { useDownloadDeviceEvidence } from "./_lib/useDownloadDeviceEvidence";

export const deviceLayoutQuery = graphql`
  query DeviceLayoutQuery($deviceId: ID!, $organizationId: ID!) {
    device: node(id: $deviceId) @required(action: THROW) {
      __typename
      ... on Device {
        id
        state
        owner {
          fullName
        }
        ...DeviceDetailsCard_deviceFragment
        ...DeviceCurrentPostures_deviceFragment
      }
    }
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canRevokeDevice: permission(action: "itam:device:revoke")
        canDeleteDevice: permission(action: "itam:device:delete")
        canDownloadEvidence: permission(action: "itam:device:download")
      }
    }
  }
`;

interface DeviceLayoutProps {
  queryRef: PreloadedQuery<DeviceLayoutQuery>;
}

export function DeviceLayout({ queryRef }: DeviceLayoutProps) {
  const { t } = useTranslation("organizations/devices");
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const [revokeOpen, setRevokeOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const { download: downloadEvidence, isDownloading } = useDownloadDeviceEvidence();
  const { root, back, header, intro, actions, section, sections } = devicesPage();

  const { device, organization } = usePreloadedQuery<DeviceLayoutQuery>(
    deviceLayoutQuery,
    queryRef,
  );
  usePageTitle(
    device.__typename === "Device"
      ? ownerLabel(device.owner?.fullName, t("devices.values.unassigned"))
      : t("devices.title"),
  );
  if (device.__typename !== "Device") {
    throw new NotFoundError("invalid type for device node");
  }
  if (organization.__typename !== "Organization") {
    throw new NotFoundError("invalid type for organization node");
  }

  const name = ownerLabel(device.owner?.fullName, t("devices.values.unassigned"));

  const deletable = isDeviceDeletable(device.state);
  const canRevokeDevice = organization.canRevokeDevice ?? false;
  const canDeleteDevice = organization.canDeleteDevice ?? false;
  const canDownloadEvidence = organization.canDownloadEvidence ?? false;

  return (
    <div className={root()}>
      <Link
        to={`/organizations/${organizationId}/itam/devices`}
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("devices.title")}
      </Link>
      <div className={header()}>
        <div className={intro()}>
          <Heading level={1} size={6} weight="medium" highContrast>
            {name}
          </Heading>
        </div>
        <div className={actions()}>
          {canDownloadEvidence && (
            <Button
              variant="soft"
              color="neutral"
              iconStart={<DownloadSimpleIcon />}
              loading={isDownloading}
              onClick={() => {
                void downloadEvidence([device.id]);
              }}
            >
              {t("devices.actions.downloadAsEvidence")}
            </Button>
          )}
          {!deletable && canRevokeDevice && (
            <Button
              variant="solid"
              color="red"
              iconStart={<PlugsIcon />}
              onClick={() => setRevokeOpen(true)}
            >
              {t("devices.actions.revoke")}
            </Button>
          )}
          {deletable && canDeleteDevice && (
            <Button variant="solid" color="red" onClick={() => setDeleteOpen(true)}>
              {t("devices.actions.delete")}
            </Button>
          )}
        </div>
      </div>

      <DeviceRevokeDialog
        open={revokeOpen}
        onOpenChange={setRevokeOpen}
        devices={[{ id: device.id, ownerName: device.owner?.fullName }]}
      />
      <DeviceDeleteDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        organizationId={organizationId}
        devices={[{ id: device.id, ownerName: device.owner?.fullName }]}
        onDeleted={() => {
          void navigate(`/organizations/${organizationId}/itam/devices`, { replace: true });
        }}
      />

      <div className={sections()}>
        <DeviceDetailsCard deviceFragmentRef={device} />
        <DeviceCurrentPostures deviceFragmentRef={device} />
        <div className={section()}>
          <Heading level={2} size={3} weight="medium">
            {t("devices.history.title")}
          </Heading>
          <Outlet />
        </div>
      </div>
    </div>
  );
}
