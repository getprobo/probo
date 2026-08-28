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

import { PlusIcon } from "@phosphor-icons/react";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useRefetchableFragment } from "react-relay";
import { useParams } from "react-router";

import type { DevicesList_organization$key } from "#/__generated__/core/DevicesList_organization.graphql";
import type { DevicesList_viewer$key } from "#/__generated__/core/DevicesList_viewer.graphql";
import type { DevicesListRefetchQuery } from "#/__generated__/core/DevicesListRefetchQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";
import { PageHeader } from "#/pages/_components/PageHeader";
import { DOCUMENT_LIST_PAGE_SIZE } from "#/pages/_lib/documentList";

import { DeviceListItem } from "./DeviceListItem";
import { DevicesEmpty } from "./DevicesEmpty";
import { devicesList } from "./variants";

const devicesListViewerFragment = graphql`
  fragment DevicesList_viewer on Viewer
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    first: { type: "Int" }
    after: { type: "CursorKey" }
    last: { type: "Int" }
    before: { type: "CursorKey" }
  )
  @refetchable(queryName: "DevicesListRefetchQuery")
  @throwOnFieldError {
    enrolledDevices(
      organizationId: $organizationId
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { field: LAST_SEEN_AT, direction: DESC }
    ) {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          ...DeviceListItem_device
        }
      }
    }
  }
`;

const devicesListOrganizationFragment = graphql`
  fragment DevicesList_organization on Organization @throwOnFieldError {
    canEnrollDevice: permission(action: "itam:device:enroll")
  }
`;

export interface DevicesListProps {
  viewerKey: DevicesList_viewer$key;
  organizationKey: DevicesList_organization$key;
}

export function DevicesList({
  viewerKey,
  organizationKey,
}: DevicesListProps) {
  const { t } = useTranslation("devices");
  const { t: tApp } = useTranslation();
  const { organizationId } = useParams();
  const organization = useFragment(devicesListOrganizationFragment, organizationKey);
  const [data, refetch] = useRefetchableFragment<
    DevicesListRefetchQuery,
    DevicesList_viewer$key
  >(devicesListViewerFragment, viewerKey);

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);

  const { enrolledDevices } = data;
  const { isPending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    enrolledDevices.pageInfo,
    DOCUMENT_LIST_PAGE_SIZE,
  );
  const slots = devicesList({ busy: isPending });

  if (organizationId === undefined) {
    throw new NotFoundError("organizationId is required");
  }

  const empty = enrolledDevices.edges.length === 0
    && !enrolledDevices.pageInfo.hasPreviousPage;
  const canEnroll = organization.canEnrollDevice;
  const registerTo = `/${organizationId}/devices/register`;

  return (
    <>
      <PageHeader
        homeLabel={tApp("homePage.breadcrumb")}
        currentLabel={t("breadcrumb")}
        title={t("list.title")}
        actions={!empty && canEnroll
          ? (
              <ButtonLink
                to={registerTo}
                size={2}
                variant="solid"
                color="neutral"
                highContrast
                iconStart={<PlusIcon />}
              >
                {t("list.register")}
              </ButtonLink>
            )
          : undefined}
      />
      {empty
        ? (
            <DevicesEmpty
              action={canEnroll
                ? (
                    <ButtonLink
                      to={registerTo}
                      size={2}
                      variant="solid"
                      color="neutral"
                      highContrast
                      iconStart={<PlusIcon />}
                    >
                      {t("empty.register")}
                    </ButtonLink>
                  )
                : undefined}
            />
          )
        : (
            <div className={slots.body()}>
              <div className={slots.frame()} aria-busy={isPending || undefined}>
                <Table variant="ghost" className={slots.table()}>
                  <TableHeader className={slots.header()}>
                    <TableRow>
                      <TableColumnHeaderCell>
                        {t("list.columns.hostname")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell>
                        {t("list.columns.details")}
                      </TableColumnHeaderCell>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {enrolledDevices.edges.map(({ node }) => (
                      <DeviceListItem key={node.id} deviceKey={node} />
                    ))}
                  </TableBody>
                </Table>
              </div>
              <Pagination
                className={slots.pager()}
                variant="surface"
                showLabels
                hasPrevious={enrolledDevices.pageInfo.hasPreviousPage}
                hasNext={enrolledDevices.pageInfo.hasNextPage}
                previousLabel={tApp("pagination.previous")}
                nextLabel={tApp("pagination.next")}
                disabled={isPending}
                onPrevious={goPrevious}
                onNext={goNext}
              />
            </div>
          )}
    </>
  );
}
