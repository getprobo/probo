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

import { faviconUrl } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import {
  ActionDropdown,
  Avatar,
  Badge,
  Button,
  DropdownItem,
  IconPageTextLine,
  IconPlusLarge,
  IconTrashCan,
  IconUpload,
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link, useNavigate } from "react-router";

import type { DataListQuery } from "#/__generated__/core/DataListQuery.graphql";
import type {
  DataPageFragment$data,
  DataPageFragment$key,
} from "#/__generated__/core/DataPageFragment.graphql";
import type { DatumGraphListQuery } from "#/__generated__/core/DatumGraphListQuery.graphql";
import { SortableTable } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import { dataQuery, useDeleteDatum } from "../../../hooks/graph/DatumGraph";

import { CreateDatumDialog } from "./dialogs/CreateDatumDialog";
import { PublishDataListDialog } from "./dialogs/PublishDataListDialog";

const paginatedDataFragment = graphql`
  fragment DataPageFragment on Organization
  @refetchable(queryName: "DataListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    order: { type: "DatumOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    data(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "DataPage_data") {
      __id
      edges {
        node {
          id
          name
          dataClassification
          owner {
            fullName
          }
          thirdParties(first: 50) {
            edges {
              node {
                id
                name
                websiteUrl
              }
            }
          }
          canUpdate: permission(action: "core:datum:update")
          canDelete: permission(action: "core:datum:delete")
        }
      }
    }
  }
`;

type DataEntry = NodeOf<DataPageFragment$data["data"]>;

type Props = {
  queryRef: PreloadedQuery<DatumGraphListQuery>;
};

export default function DataPage(props: Props) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const { node: data } = usePreloadedQuery<DatumGraphListQuery>(
    dataQuery,
    props.queryRef,
  );

  const pagination = usePaginationFragment<DataListQuery, DataPageFragment$key>(
    paginatedDataFragment,
    data,
  );

  const dataEntries = pagination.data.data.edges.map(edge => edge.node);
  const connectionId = pagination.data.data.__id;
  const defaultApproverIds = (data.dataListDocument?.defaultApprovers ?? []).map(a => a.id);

  const refetch = ({
    order,
  }: {
    order: { direction: string; field: string };
  }) => {
    pagination.refetch({
      order: {
        direction: order.direction as "ASC" | "DESC",
        field: order.field as "CREATED_AT" | "DATA_CLASSIFICATION" | "NAME",
      },
    });
  };

  usePageTitle(t("dataPage.pageTitle"));

  const hasAnyAction
    = dataEntries.some(({ canDelete, canUpdate }) => canUpdate || canDelete);

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("dataPage.title")}
        description={t("dataPage.description")}
      >
        <div className="flex gap-2">
          {data.dataListDocument?.id && (
            <Button variant="secondary" asChild>
              <Link
                to={`/organizations/${organizationId}/documents/${data.dataListDocument.id}`}
              >
                <IconPageTextLine size={16} />
                {t("dataPage.actions.document")}
              </Link>
            </Button>
          )}
          {data.canPublishData && (
            <PublishDataListDialog
              organizationId={organizationId}
              defaultApproverIds={defaultApproverIds}
              onPublished={(documentId) => {
                void navigate(
                  `/organizations/${organizationId}/documents/${documentId}`,
                );
              }}
            >
              <Button variant="secondary" icon={IconUpload}>
                {t("dataPage.actions.publish")}
              </Button>
            </PublishDataListDialog>
          )}
          {data.canCreateDatum && (
            <CreateDatumDialog
              connection={connectionId}
              organizationId={organizationId}
              onCreated={() => pagination.refetch({})}
            >
              <Button icon={IconPlusLarge}>{t("dataPage.actions.add")}</Button>
            </CreateDatumDialog>
          )}
        </div>
      </PageHeader>
      <SortableTable {...pagination} refetch={refetch} pageSize={10}>
        <Thead>
          <Tr>
            <Th>{t("dataPage.columns.name")}</Th>
            <Th>{t("dataPage.columns.classification")}</Th>
            <Th>{t("dataPage.columns.owner")}</Th>
            <Th>{t("dataPage.columns.thirdParties")}</Th>
            {hasAnyAction && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {dataEntries.map(entry => (
            <DataRow
              key={entry.id}
              entry={entry}
              connectionId={connectionId}
              hasAnyAction={hasAnyAction}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

function DataRow({
  entry,
  connectionId,
  hasAnyAction,
}: {
  entry: DataEntry;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const deleteDatum = useDeleteDatum(entry, connectionId);
  const thirdParties = entry.thirdParties?.edges.map(edge => edge.node) ?? [];
  const detailUrl = `/organizations/${organizationId}/data/${entry.id}`;

  return (
    <Tr to={detailUrl}>
      <Td>{entry.name}</Td>
      <Td>
        <Badge variant="info">{entry.dataClassification}</Badge>
      </Td>
      <Td>{entry.owner?.fullName ?? t("dataPage.unassigned")}</Td>
      <Td>
        {thirdParties.length > 0
          ? (
              <div className="flex flex-wrap gap-1">
                {thirdParties.slice(0, 3).map(thirdParty => (
                  <Badge
                    key={thirdParty.id}
                    variant="neutral"
                    className="flex items-center gap-1"
                  >
                    <Avatar
                      name={thirdParty.name}
                      src={faviconUrl(thirdParty.websiteUrl)}
                      size="s"
                    />
                    <span className="text-xs">{thirdParty.name}</span>
                  </Badge>
                ))}
                {thirdParties.length > 3 && (
                  <Badge variant="neutral" className="text-xs">
                    +
                    {thirdParties.length - 3}
                  </Badge>
                )}
              </div>
            )
          : (
              <span className="text-txt-secondary text-sm">{t("dataPage.none")}</span>
            )}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {entry.canDelete && (
              <DropdownItem
                onClick={deleteDatum}
                variant="danger"
                icon={IconTrashCan}
              >
                {t("dataPage.actions.delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
