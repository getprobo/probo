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

import { CheckCircleIcon, WarningIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { CardSkeleton } from "@probo/ui/src/v2/Card/CardSkeleton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeviceInstallationCardDevices_organization$key } from "#/__generated__/core/DeviceInstallationCardDevices_organization.graphql";
import type { DeviceInstallationCardDevicesRefetchQuery } from "#/__generated__/core/DeviceInstallationCardDevicesRefetchQuery.graphql";
import type { DeviceInstallationCardEmployees_organization$key } from "#/__generated__/core/DeviceInstallationCardEmployees_organization.graphql";
import type { DeviceInstallationCardEmployeesRefetchQuery } from "#/__generated__/core/DeviceInstallationCardEmployeesRefetchQuery.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { deviceInstallationCoverage } from "../_lib/deviceInstallationCoverage";
import { useSendDeviceInstallationEmails } from "../_lib/useSendDeviceInstallationEmails";

const installationPageSize = 100;

const deviceInstallationEmployeesFragment = graphql`
  fragment DeviceInstallationCardEmployees_organization on Organization
  @refetchable(queryName: "DeviceInstallationCardEmployeesRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 100 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    canSendInstallationEmail: permission(action: "itam:device:send-installation-email")
    profiles(
      first: $first
      after: $after
      orderBy: { direction: ASC, field: CREATED_AT }
      filter: { contractEnded: false, states: [ACTIVE, PENDING], kind: "EMPLOYEE" }
    ) @connection(key: "DeviceInstallationCardEmployees_profiles", filters: []) {
      edges {
        node {
          id
        }
      }
    }
  }
`;

const deviceInstallationDevicesFragment = graphql`
  fragment DeviceInstallationCardDevices_organization on Organization
  @refetchable(queryName: "DeviceInstallationCardDevicesRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 100 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    installationDevices: devices(
      first: $first
      after: $after
      orderBy: { direction: ASC, field: CREATED_AT }
    ) @connection(key: "DevicesPage_installationDevices", filters: []) {
      edges {
        node {
          id
          state
          owner {
            id
          }
        }
      }
    }
  }
`;

interface DeviceInstallationCardProps {
  organizationKey: DeviceInstallationCardEmployees_organization$key
    & DeviceInstallationCardDevices_organization$key;
  refreshToken: number;
}

export function DeviceInstallationCard({
  organizationKey,
  refreshToken,
}: DeviceInstallationCardProps) {
  const { t } = useTranslation("organizations/devices");
  const organizationId = useOrganizationId();
  const [sendInstallationEmails, isSending] = useSendDeviceInstallationEmails();
  const [isRefreshing, startRefresh] = useTransition();
  const {
    data: employeesData,
    hasNext: employeesHasNext,
    isLoadingNext: employeesLoading,
    loadNext: loadNextEmployees,
  } = usePaginationFragment<
    DeviceInstallationCardEmployeesRefetchQuery,
    DeviceInstallationCardEmployees_organization$key
  >(deviceInstallationEmployeesFragment, organizationKey);
  const {
    data: devicesData,
    hasNext: devicesHasNext,
    isLoadingNext: devicesLoading,
    loadNext: loadNextDevices,
    refetch: refetchDevices,
  } = usePaginationFragment<
    DeviceInstallationCardDevicesRefetchQuery,
    DeviceInstallationCardDevices_organization$key
  >(deviceInstallationDevicesFragment, organizationKey);
  const seenRefresh = useRef(refreshToken);

  useEffect(() => {
    if (employeesHasNext && !employeesLoading) {
      loadNextEmployees(installationPageSize);
    }
  }, [employeesHasNext, employeesLoading, loadNextEmployees]);

  useEffect(() => {
    if (isRefreshing || !devicesHasNext || devicesLoading) {
      return;
    }
    loadNextDevices(installationPageSize);
  }, [isRefreshing, devicesHasNext, devicesLoading, loadNextDevices]);

  useEffect(() => {
    if (seenRefresh.current === refreshToken) {
      return;
    }
    seenRefresh.current = refreshToken;
    startRefresh(() => {
      refetchDevices(
        { first: installationPageSize, after: null },
        { fetchPolicy: "store-and-network" },
      );
    });
  }, [refreshToken, refetchDevices]);

  const ready = !employeesHasNext
    && !devicesHasNext
    && !employeesLoading
    && !devicesLoading
    && !isRefreshing;

  if (!ready) {
    return <CardSkeleton size={2} className="h-18" />;
  }

  const coverage = deviceInstallationCoverage(
    employeesData.profiles.edges.map(edge => edge.node.id),
    devicesData.installationDevices.edges.map(edge => ({
      state: edge.node.state,
      ownerId: edge.node.owner?.id ?? null,
    })),
  );
  const complete = coverage.tone === "green";
  const showSend = !complete && employeesData.canSendInstallationEmail;

  function handleSend() {
    void sendInstallationEmails(organizationId).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <TonedCard
      tone={coverage.tone}
      fill={false}
      icon={complete
        ? <CheckCircleIcon weight="duotone" />
        : <WarningIcon weight="duotone" />}
      lead={(
        <Heading level={2} size={4} weight="medium" highContrast>
          {complete
            ? t("devices.installation.completeTitle")
            : t("devices.installation.incompleteTitle", { count: coverage.missing })}
        </Heading>
      )}
      control={showSend
        ? (
            <Button
              size={2}
              variant="solid"
              color="neutral"
              highContrast
              loading={isSending}
              onClick={handleSend}
            >
              {t("devices.installation.sendEmail")}
            </Button>
          )
        : undefined}
    >
      {!complete && (
        <Text size={2} color="neutral">
          {t("devices.installation.recommendation")}
        </Text>
      )}
    </TonedCard>
  );
}
