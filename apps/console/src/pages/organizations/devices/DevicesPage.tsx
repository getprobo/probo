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

import {
  DotsThreeVerticalIcon,
  DownloadSimpleIcon,
  PlugsIcon,
  PlusIcon,
  TrashIcon,
} from "@phosphor-icons/react";
import { useList, usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { DevicesPageFragment$key } from "#/__generated__/core/DevicesPageFragment.graphql";
import type {
  DeviceOrder,
  DevicesPageFragment_RefetchQuery,
} from "#/__generated__/core/DevicesPageFragment_RefetchQuery.graphql";
import type { DevicesPageQuery } from "#/__generated__/core/DevicesPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { DeviceDeleteDialog } from "./_components/DeviceDeleteDialog";
import { DeviceListItem } from "./_components/DeviceListItem";
import { DeviceRevokeDialog } from "./_components/DeviceRevokeDialog";
import { DevicesListSearch } from "./_components/DevicesListSearch";
import { devicesPage } from "./_components/variants";
import {
  type DevicePlatformName,
  devicePlatforms,
  platformLabel,
} from "./_lib/deviceDisplay";
import {
  type DeviceStateFilter,
  useDevicesListFilters,
} from "./_lib/useDevicesListFilters";
import { useDownloadDeviceEvidence } from "./_lib/useDownloadDeviceEvidence";
import { CreateDeviceDialog } from "./dialogs/CreateDeviceDialog";

export const devicePageSize = 25;

export const devicesPageQuery = graphql`
  query DevicesPageQuery(
    $organizationId: ID!
    $query: String
    $state: DeviceState
    $platform: DevicePlatform
    $agentVersion: String
  ) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        canAssignDevice: permission(action: "itam:device:assign")
        canRevokeDevice: permission(action: "itam:device:revoke")
        canDeleteDevice: permission(action: "itam:device:delete")
        canCreateDevice: permission(action: "itam:device:create")
        canDownloadEvidence: permission(action: "itam:device:download")
        ...DevicesPageFragment
          @arguments(
            query: $query
            state: $state
            platform: $platform
            agentVersion: $agentVersion
          )
      }
    }
  }
`;

const devicesPageFragment = graphql`
  fragment DevicesPageFragment on Organization
  @refetchable(queryName: "DevicesPageFragment_RefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 25 }
    order: {
      type: "DeviceOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    query: { type: "String", defaultValue: null }
    state: { type: "DeviceState", defaultValue: null }
    platform: { type: "DevicePlatform", defaultValue: null }
    agentVersion: { type: "String", defaultValue: null }
  ) {
    deviceAgentVersions
    devices(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: {
        query: $query
        state: $state
        platform: $platform
        agentVersion: $agentVersion
      }
    ) @connection(key: "DevicesPage_devices", filters: ["filter"]) {
      __id
      totalCount
      edges {
        node {
          id
          owner {
            fullName
          }
          ...DeviceListItem_device
        }
      }
    }
  }
`;

interface DevicesPageProps {
  queryRef: PreloadedQuery<DevicesPageQuery>;
}

const initialOrder = {
  direction: "DESC",
  field: "CREATED_AT",
} as const satisfies DeviceOrder;

export function DevicesPage({ queryRef }: DevicesPageProps) {
  const { t } = useTranslation("organizations/devices");
  const { download: downloadEvidence, isDownloading } = useDownloadDeviceEvidence();
  const organizationId = useOrganizationId();
  const { list: selection, toggle, reset } = useList<string>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [revokeOpen, setRevokeOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [order, setOrder] = useState<DeviceOrder>(initialOrder);
  const [isPending, startTransition] = useTransition();
  const {
    ownerQuery,
    state,
    platform,
    agentVersion,
    setState,
    setPlatform,
    setAgentVersion,
  } = useDevicesListFilters();
  const skipFirstRefetch = useRef(true);

  usePageTitle(t("devices.title"));

  const { organization } = usePreloadedQuery<DevicesPageQuery>(
    devicesPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new NotFoundError("invalid type for organization node");
  }

  const pagination = usePaginationFragment<
    DevicesPageFragment_RefetchQuery,
    DevicesPageFragment$key
  >(devicesPageFragment, organization);

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    startTransition(() => {
      pagination.refetch(
        {
          query: ownerQuery,
          state,
          platform,
          agentVersion,
          order,
          first: devicePageSize,
          after: null,
          last: null,
          before: null,
        },
        { fetchPolicy: "store-and-network" },
      );
    });
  }, [ownerQuery, state, platform, agentVersion, order, pagination.refetch]);

  const devices = pagination.data.devices.edges.map(edge => edge.node);
  const agentVersions = deviceAgentVersions(
    pagination.data.deviceAgentVersions,
    agentVersion,
  );
  const hasListFilter = ownerQuery != null
    || state != null
    || platform != null
    || agentVersion != null;
  const connectionId = pagination.data.devices.__id;
  const canCreate = organization.canCreateDevice ?? false;
  const canDownloadEvidence = organization.canDownloadEvidence ?? false;
  const canRevoke = organization.canRevokeDevice ?? false;
  const canDelete = organization.canDeleteDevice ?? false;
  const canSelectDevices = canDownloadEvidence || canRevoke || canDelete;
  const showActionsMenu = canCreate || canRevoke || canDelete;
  const deviceIds = devices.map(device => device.id);
  const deviceIdSet = new Set(deviceIds);
  const selectedIds = selection.filter(id => deviceIdSet.has(id));
  const selectedDevices = devices
    .filter(device => selectedIds.includes(device.id))
    .map(device => ({ id: device.id, ownerName: device.owner?.fullName }));
  const hasSelection = selectedIds.length > 0;
  const allVisibleSelected = deviceIds.length > 0
    && deviceIds.every(id => selectedIds.includes(id));
  const {
    root,
    header,
    intro,
    title,
    actions,
    more,
    results,
    list,
    tools,
    tableFilters,
    filter,
    platformFilter: platformFilterClass,
    versionFilter,
  } = devicesPage({ pending: isPending });

  function refetchDevices() {
    startTransition(() => {
      pagination.refetch(
        {
          query: ownerQuery,
          state,
          platform,
          agentVersion,
          order,
          first: devicePageSize,
          after: null,
          last: null,
          before: null,
        },
        { fetchPolicy: "store-and-network" },
      );
    });
  }

  function changeOrder(field: "LAST_SEEN_AT") {
    setOrder({
      field,
      direction: order.field === field && order.direction === "DESC" ? "ASC" : "DESC",
    });
  }

  function sortState(field: "LAST_SEEN_AT") {
    if (order.field !== field) {
      return "none" as const;
    }
    return order.direction === "ASC" ? "ascending" as const : "descending" as const;
  }

  function downloadSelectedEvidence() {
    void downloadEvidence(selectedIds);
  }

  return (
    <div className={root()}>
      <div className={header()}>
        <div className={intro()}>
          <div className={title()}>
            <Heading level={1} size={6} weight="medium" highContrast>
              {t("devices.title")}
            </Heading>
            <Text size={2} color="faint">{pagination.data.devices.totalCount}</Text>
          </div>
          <Text size={2} color="faint">{t("devices.description")}</Text>
        </div>
        <div className={actions()}>
          {canDownloadEvidence && (
            <Button
              variant="solid"
              iconStart={<DownloadSimpleIcon />}
              disabled={!hasSelection}
              loading={isDownloading}
              onClick={downloadSelectedEvidence}
            >
              {hasSelection
                ? t("devices.actions.downloadAsEvidenceCount", { count: selectedIds.length })
                : t("devices.actions.downloadAsEvidence")}
            </Button>
          )}
          {showActionsMenu && (
            <Dropdown>
              <DropdownTrigger
                render={(
                  <IconButton
                    variant="soft"
                    color="neutral"
                    aria-label={t("devices.actions.menu")}
                  >
                    <DotsThreeVerticalIcon />
                  </IconButton>
                )}
              />
              <DropdownPopup align="end">
                {canCreate && (
                  <DropdownItem
                    iconStart={<PlusIcon />}
                    onClick={() => setCreateOpen(true)}
                  >
                    {t("devices.actions.add")}
                  </DropdownItem>
                )}
                {canRevoke && (
                  <DropdownItem
                    color="error"
                    iconStart={<PlugsIcon />}
                    disabled={!hasSelection}
                    onClick={() => setRevokeOpen(true)}
                  >
                    {t("devices.actions.revoke")}
                  </DropdownItem>
                )}
                {canDelete && (
                  <DropdownItem
                    color="error"
                    iconStart={<TrashIcon />}
                    disabled={!hasSelection}
                    onClick={() => setDeleteOpen(true)}
                  >
                    {t("devices.actions.delete")}
                  </DropdownItem>
                )}
              </DropdownPopup>
            </Dropdown>
          )}
        </div>
      </div>

      {canCreate && (
        <CreateDeviceDialog
          open={createOpen}
          onOpenChange={setCreateOpen}
          organizationId={organizationId}
          onCreated={refetchDevices}
        />
      )}

      <DeviceRevokeDialog
        open={revokeOpen}
        onOpenChange={setRevokeOpen}
        devices={selectedDevices}
        onRevoked={refetchDevices}
      />
      <DeviceDeleteDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        organizationId={organizationId}
        connectionId={connectionId}
        devices={selectedDevices}
        onDeleted={() => reset([])}
      />

      <div className={results()}>
        <div className={tools()}>
          <DevicesListSearch />
          <div className={tableFilters()}>
            <div className={filter()}>
              <Select
                value={state}
                onValueChange={(value: DeviceStateFilter | null) => {
                  setState(value);
                }}
              >
                <SelectTrigger
                  size={2}
                  placeholder={t("devices.filters.allStates")}
                  aria-label={t("devices.fields.state")}
                >
                  {(value: DeviceStateFilter | null) => deviceStateLabel(t, value)}
                </SelectTrigger>
                <SelectPopup align="end">
                  <SelectItem value={null}>{t("devices.filters.allStates")}</SelectItem>
                  <SelectItem value="PENDING">{t("devices.filters.pending")}</SelectItem>
                  <SelectItem value="ACTIVE">{t("devices.filters.active")}</SelectItem>
                  <SelectItem value="REVOKED">{t("devices.filters.revoked")}</SelectItem>
                </SelectPopup>
              </Select>
            </div>
            <div className={platformFilterClass()}>
              <Select
                value={platform}
                onValueChange={(value: DevicePlatformName | null) => {
                  setPlatform(value);
                }}
              >
                <SelectTrigger
                  size={2}
                  placeholder={t("devices.filters.allPlatforms")}
                  aria-label={t("devices.fields.platform")}
                >
                  {(value: DevicePlatformName | null) => devicePlatformLabel(t, value)}
                </SelectTrigger>
                <SelectPopup align="end">
                  <SelectItem value={null}>{t("devices.filters.allPlatforms")}</SelectItem>
                  {devicePlatforms.map(platform => (
                    <SelectItem key={platform} value={platform}>
                      {platformLabel(t, platform)}
                    </SelectItem>
                  ))}
                </SelectPopup>
              </Select>
            </div>
            <div className={versionFilter()}>
              <Select
                value={agentVersion}
                onValueChange={(value: string | null) => {
                  setAgentVersion(value);
                }}
              >
                <SelectTrigger
                  size={2}
                  placeholder={t("devices.filters.allAgentVersions")}
                  aria-label={t("devices.fields.agentVersion")}
                >
                  {(value: string | null) => value ?? t("devices.filters.allAgentVersions")}
                </SelectTrigger>
                <SelectPopup align="end">
                  <SelectItem value={null}>{t("devices.filters.allAgentVersions")}</SelectItem>
                  {agentVersions.map(version => (
                    <SelectItem key={version} value={version}>
                      {version}
                    </SelectItem>
                  ))}
                </SelectPopup>
              </Select>
            </div>
          </div>
        </div>
        <div aria-busy={isPending} className={list()}>
          <Table size={2} variant="surface">
            <TableHeader>
              <TableRow align="center">
                {canSelectDevices && (
                  <TableColumnHeaderCell width="3rem">
                    <Checkbox
                      checked={allVisibleSelected}
                      indeterminate={hasSelection && !allVisibleSelected}
                      disabled={deviceIds.length === 0}
                      onCheckedChange={(checked) => {
                        reset(checked === true ? deviceIds : []);
                      }}
                      aria-label={t("devices.actions.selectAll")}
                    />
                  </TableColumnHeaderCell>
                )}
                <TableColumnHeaderCell>{t("devices.fields.owner")}</TableColumnHeaderCell>
                <TableColumnHeaderCell>{t("devices.fields.state")}</TableColumnHeaderCell>
                <TableColumnHeaderCell>{t("devices.fields.platform")}</TableColumnHeaderCell>
                <TableColumnHeaderCell>{t("devices.fields.osVersion")}</TableColumnHeaderCell>
                <TableColumnHeaderCell>{t("devices.fields.agentVersion")}</TableColumnHeaderCell>
                <TableColumnHeaderCell
                  sort={sortState("LAST_SEEN_AT")}
                  onSort={() => changeOrder("LAST_SEEN_AT")}
                >
                  {t("devices.fields.lastSeen")}
                </TableColumnHeaderCell>
                <TableColumnHeaderCell>{t("devices.fields.postures")}</TableColumnHeaderCell>
                <TableColumnHeaderCell width="3rem" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {devices.length === 0
                ? (
                    <TableRow>
                      <TableCell colSpan={canSelectDevices ? 9 : 8} justify="center">
                        <Text size={2} color="faint">
                          {pagination.data.devices.totalCount === 0 && !hasListFilter
                            ? t("devices.empty")
                            : t("devices.filters.empty")}
                        </Text>
                      </TableCell>
                    </TableRow>
                  )
                : devices.map(device => (
                    <DeviceListItem
                      key={device.id}
                      fKey={device}
                      connectionId={connectionId}
                      canAssignDevice={organization.canAssignDevice ?? false}
                      canRevoke={canRevoke}
                      canDelete={canDelete}
                      onRevoked={refetchDevices}
                      selected={canSelectDevices ? selectedIds.includes(device.id) : undefined}
                      onSelectedChange={canSelectDevices
                        ? (checked) => {
                            const has = selection.includes(device.id);
                            if (checked !== has) {
                              toggle(device.id);
                            }
                          }
                        : undefined}
                    />
                  ))}
            </TableBody>
          </Table>
          {pagination.hasNext && (
            <div className={more()}>
              <Button
                variant="soft"
                color="neutral"
                loading={pagination.isLoadingNext}
                onClick={() => pagination.loadNext(devicePageSize)}
              >
                {t("sortableTable.actions.showMore")}
              </Button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function deviceAgentVersions(
  versions: readonly string[],
  selected: string | null,
): string[] {
  const unique = new Set(versions);
  if (selected != null) {
    unique.add(selected);
  }
  return [...unique].sort((left, right) => (
    left.localeCompare(right, undefined, { numeric: true })
  ));
}

function devicePlatformLabel(t: (key: string) => string, value: string | null) {
  return platformLabel(t, value) ?? t("devices.filters.allPlatforms");
}

function deviceStateLabel(t: (key: string) => string, value: string | null) {
  switch (value) {
    case "PENDING":
      return t("devices.filters.pending");
    case "ACTIVE":
      return t("devices.filters.active");
    case "REVOKED":
      return t("devices.filters.revoked");
    default:
      return t("devices.filters.allStates");
  }
}
