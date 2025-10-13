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
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useParams } from "react-router";
import { CreateAssetDialog } from "./dialogs/CreateAssetDialog";
import { useDeleteAsset, assetsQuery } from "../../../hooks/graph/AssetGraph";
import type { AssetGraphListQuery } from "/hooks/graph/__generated__/AssetGraphListQuery.graphql";
import { faviconUrl } from "@probo/helpers";
import type { NodeOf } from "/types";
import { getAssetTypeVariant } from "@probo/helpers";
import type {
  AssetsPageFragment$data,
  AssetsPageFragment$key,
} from "./__generated__/AssetsPageFragment.graphql";
import { SortableTable } from "/components/SortableTable";
import { SnapshotBanner } from "/components/SnapshotBanner";

const paginatedAssetsFragment = graphql`
  fragment AssetsPageFragment on Organization
  @refetchable(queryName: "AssetsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    orderBy: { type: "AssetOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    assets(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $orderBy
      filter: { snapshotId: $snapshotId }
    ) @connection(key: "AssetsPage_assets", filters: ["filter"]) {
      __id
      edges {
        node {
          id
          snapshotId
          name
          amount
          assetType
          dataTypesStored
          owner {
            id
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

type AssetEntry = NodeOf<AssetsPageFragment$data["assets"]>;

type Props = {
  queryRef: PreloadedQuery<AssetGraphListQuery>;
};

export default function AssetsPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const data = usePreloadedQuery(assetsQuery, props.queryRef);
  const pagination = usePaginationFragment(
    paginatedAssetsFragment,
    data.node as AssetsPageFragment$key
  );
  const assets = pagination.data.assets?.edges.map((edge) => edge.node);
  const connectionId = pagination.data.assets.__id;

  usePageTitle(__("Assets"));

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <PageHeader
        title={__("Assets")}
        description={__(
          "Manage your organization's assets and their classifications."
        )}
      >
        {!isSnapshotMode && (
          <CreateAssetDialog
            connection={connectionId}
            organizationId={organizationId}
          >
            <Button icon={IconPlusLarge}>{__("Add asset")}</Button>
          </CreateAssetDialog>
        )}
      </PageHeader>
      <SortableTable {...pagination}>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Type")}</Th>
            <Th>{__("Amount")}</Th>
            <Th>{__("Owner")}</Th>
            <Th>{__("Vendors")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {assets.map((entry) => (
            <AssetRow
              key={entry.id}
              entry={entry}
              connectionId={connectionId}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

function AssetRow({
  entry,
  connectionId,
}: {
  entry: AssetEntry;
  connectionId: string;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const deleteAsset = useDeleteAsset(entry, connectionId);
  const vendors = entry.vendors?.edges.map((edge) => edge.node) ?? [];

  const assetUrl = isSnapshotMode && snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/assets/${entry.id}`
    : `/organizations/${organizationId}/assets/${entry.id}`;

  return (
    <Tr to={assetUrl}>
      <Td>{entry.name}</Td>
      <Td>
        <Badge variant={getAssetTypeVariant(entry.assetType)}>
          {entry.assetType === "PHYSICAL" ? __("Physical") : __("Virtual")}
        </Badge>
      </Td>
      <Td>{entry.amount}</Td>
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
        {!isSnapshotMode && (
          <ActionDropdown>
            <DropdownItem
              onClick={deleteAsset}
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
