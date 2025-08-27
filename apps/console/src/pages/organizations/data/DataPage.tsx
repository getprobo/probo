import {
  Button,
  IconPlusLarge,
  PageHeader,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Badge,
  ActionDropdown,
  DropdownItem,
  IconTrashCan,
  Avatar,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import {
  graphql,
  usePaginationFragment,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { CreateDatumDialog } from "./dialogs/CreateDatumDialog";
import { useDeleteDatum, dataQuery } from "../../../hooks/graph/DatumGraph";
import type { DatumGraphListQuery } from "/hooks/graph/__generated__/DatumGraphListQuery.graphql";
import { faviconUrl } from "@probo/helpers";
import type { NodeOf } from "/types";
import type {
  DataPageFragment$key,
  DataPageFragment$data
} from "./__generated__/DataPageFragment.graphql";
import type { DataListQuery } from "./__generated__/DataListQuery.graphql";
import { SortableTable } from "/components/SortableTable";
import { SnapshotBanner } from "/components/SnapshotBanner";

const paginatedDataFragment = graphql`
  fragment DataPageFragment on Organization
  @refetchable(queryName: "DataListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    order: { type: "DatumOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    data(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: { snapshotId: $snapshotId }
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
          vendors(first: 50) {
            edges {
              node {
                id
                name
                websiteUrl
              }
            }
          }
          createdAt
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
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const queryData = usePreloadedQuery<DatumGraphListQuery>(
    dataQuery,
    props.queryRef
  );

  const data = queryData.node;

  const pagination = usePaginationFragment<DataListQuery, DataPageFragment$key>(
    paginatedDataFragment,
    data
  );

  const dataEntries = pagination.data.data.edges.map((edge) => edge.node);
  const connectionId = pagination.data.data.__id;

  const refetch = ({ order }: { order: { direction: string; field: string } }) => {
    pagination.refetch({
      snapshotId,
      order: {
        direction: order.direction as "ASC" | "DESC",
        field: order.field as "CREATED_AT" | "DATA_CLASSIFICATION" | "NAME"
      }
    });
  };

  usePageTitle(__("Data"));

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <PageHeader
        title={__("Data")}
        description={__(
          "Manage your organization's data assets and their classifications."
        )}
      >
        {!snapshotId && (
          <CreateDatumDialog
            connection={connectionId}
            organizationId={organizationId}
            onCreated={() => pagination.refetch({ snapshotId })}
          >
            <Button icon={IconPlusLarge}>{__("Add data")}</Button>
          </CreateDatumDialog>
        )}
      </PageHeader>
      <SortableTable
        {...pagination}
        refetch={refetch}
      >
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Classification")}</Th>
            <Th>{__("Owner")}</Th>
            <Th>{__("Vendors")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {dataEntries.map((entry) => (
            <DataRow key={entry.id} entry={entry} connectionId={connectionId} snapshotId={snapshotId} />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

function DataRow({
  entry,
  connectionId,
  snapshotId,
}: {
  entry: DataEntry;
  connectionId: string;
  snapshotId?: string;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const deleteDatum = useDeleteDatum(entry, connectionId);
  const vendors = entry.vendors?.edges.map((edge) => edge.node) ?? [];

  const detailUrl = snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/data/${entry.id}`
    : `/organizations/${organizationId}/data/${entry.id}`;

  return (
    <Tr to={detailUrl}>
      <Td>{entry.name}</Td>
      <Td>
        <Badge variant="info">{entry.dataClassification}</Badge>
      </Td>
      <Td>{entry.owner?.fullName ?? __("Unassigned")}</Td>
      <Td>
        {vendors.length > 0 ? (
          <div className="flex flex-wrap gap-1">
            {vendors.slice(0, 3).map((vendor) => (
              <Badge
                key={vendor.id}
                variant="neutral"
                className="flex items-center gap-1"
              >
                <Avatar
                  name={vendor.name}
                  src={faviconUrl(vendor.websiteUrl)}
                  size="s"
                />
                <span className="text-xs">{vendor.name}</span>
              </Badge>
            ))}
            {vendors.length > 3 && (
              <Badge variant="neutral" className="text-xs">
                +{vendors.length - 3}
              </Badge>
            )}
          </div>
        ) : (
          <span className="text-txt-secondary text-sm">{__("None")}</span>
        )}
      </Td>
      <Td noLink width={50} className="text-end">
      {!snapshotId && (<ActionDropdown>
            <DropdownItem
              onClick={deleteDatum}
              variant="danger"
              icon={IconTrashCan}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </Td>
    </Tr>
  );
}
