// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { ArrowRightIcon, DevicesIcon, TrayIcon } from "@phosphor-icons/react";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { useParams } from "react-router";

import type { DeviceCard_organization$key } from "#/__generated__/core/DeviceCard_organization.graphql";
import type { DeviceCard_viewer$key } from "#/__generated__/core/DeviceCard_viewer.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { dashboardCard, deviceCard } from "./variants";

const deviceCardViewerFragment = graphql`
  fragment DeviceCard_viewer on Viewer
  @argumentDefinitions(organizationId: { type: "ID!" })
  @throwOnFieldError {
    enrolledDevices(organizationId: $organizationId, first: 20) {
      edges {
        node {
          state
        }
      }
    }
  }
`;

const deviceCardOrganizationFragment = graphql`
  fragment DeviceCard_organization on Organization @throwOnFieldError {
    canEnrollDevice: permission(action: "itam:device:enroll")
  }
`;

export interface DeviceCardProps {
  viewerKey: DeviceCard_viewer$key;
  organizationKey: DeviceCard_organization$key;
}

export function DeviceCard({
  viewerKey,
  organizationKey,
}: DeviceCardProps) {
  const { t } = useTranslation();
  const { organizationId } = useParams();
  const slots = dashboardCard();
  const status = deviceCard();
  const viewer = useFragment(deviceCardViewerFragment, viewerKey);
  const organization = useFragment(deviceCardOrganizationFragment, organizationKey);

  if (organizationId == null) {
    throw new NotFoundError("organizationId is required");
  }

  const connected = viewer.enrolledDevices.edges.some(({ node }) => {
    return node.state === "ACTIVE" || node.state === "PENDING";
  });
  const canEnroll = organization.canEnrollDevice;

  return (
    <Card variant="soft" padding="none" size={3} className={slots.frame()}>
      <div className={slots.header()}>
        <DevicesIcon className={slots.icon()} />
        <div className={slots.copy()}>
          <Heading level={2} size={2} weight="medium" highContrast>
            {t("homePage.dashboard.devices.title")}
          </Heading>
          <Text size={1} color="current" className={slots.description()}>
            {t("homePage.dashboard.devices.description")}
          </Text>
        </div>
        <Link
          to={`/${organizationId}/devices`}
          size={2}
          color="neutral"
          underline={false}
          className={slots.view()}
        >
          {t("homePage.dashboard.view")}
        </Link>
      </div>
      <div className={slots.body()}>
        {connected
          ? (
              <div className={slots.empty()}>
                <span className={status.status()}>
                  <span className={status.pip()} />
                </span>
                <Text size={2} weight="medium" color="current" className={slots.emptyLabel()}>
                  {t("homePage.dashboard.devices.connected")}
                </Text>
              </div>
            )
          : (
              <div className={slots.empty()}>
                <TrayIcon className={slots.emptyIcon()} />
                {canEnroll
                  ? (
                      <ButtonLink
                        to={`/${organizationId}/devices/register`}
                        size={2}
                        variant="soft"
                        color="neutral"
                        iconEnd={<ArrowRightIcon />}
                      >
                        {t("homePage.dashboard.devices.action")}
                      </ButtonLink>
                    )
                  : null}
              </div>
            )}
      </div>
    </Card>
  );
}
