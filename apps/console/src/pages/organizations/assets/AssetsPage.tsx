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
  EditableCell,
  Select,
  Option,
  DataTable,
  CellHead,
  Cell,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import {
  graphql,
  usePaginationFragment,
  usePreloadedQuery,
  type PreloadedQuery,
  useMutation,
} from "react-relay";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useParams } from "react-router";
import { CreateAssetDialog } from "./dialogs/CreateAssetDialog";
import {
  useDeleteAsset,
  assetsQuery,
  updateAssetMutation,
} from "../../../hooks/graph/AssetGraph";
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
import { Authorized } from "/permissions";
import { isAuthorized } from "/permissions";
import { PeopleSelectOptions } from "/components/form/PeopleSelectField.tsx";

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
    data.node as AssetsPageFragment$key,
  );
  const assets = pagination.data.assets?.edges.map((edge) => edge.node);
  const connectionId = pagination.data.assets.__id;

  usePageTitle(__("Assets"));

  const hasAnyAction = !isSnapshotMode && (
    isAuthorized(organizationId, "Asset", "updateAsset") ||
    isAuthorized(organizationId, "Asset", "deleteAsset")
  );

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <PageHeader
        title={__("Assets")}
        description={__(
          "Manage your organization's assets and their classifications.",
        )}
      >
        {!isSnapshotMode && (
          <Authorized entity="Organization" action="createAsset">
            <CreateAssetDialog
              connection={connectionId}
              organizationId={organizationId}
            >
              <Button icon={IconPlusLarge}>{__("Add asset")}</Button>
            </CreateAssetDialog>
          </Authorized>
        )}
      </PageHeader>
      <DataTable columns={6}>
        <CellHead>{__("Name")}</CellHead>
        <CellHead>{__("Type")}</CellHead>
        <CellHead>{__("Amount")}</CellHead>
        <CellHead>{__("Owner")}</CellHead>
        <CellHead>{__("Vendors")}</CellHead>
        <CellHead></CellHead>
        {assets.map((entry) => (
          <AssetRow key={entry.id} entry={entry} connectionId={connectionId} />
        ))}
      </DataTable>
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

  const assetUrl =
    isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/assets/${entry.id}`
      : `/organizations/${organizationId}/assets/${entry.id}`;

  const [mutate, isLoading] = useMutation(updateAssetMutation);
  const updater = (fieldName: keyof typeof entry) => (value: string) => {
    // Only send an update if the value changed
    if (entry[fieldName] === value) {
      return;
    }
    mutate({
      variables: {
        input: {
          id: entry.id,
          [fieldName]: value,
        },
      },
    });
  };

  return (
    <>
      <EditableCell
        type="text"
        defaultValue={entry.name}
        onValueChange={updater("name")}
      />
      <EditableCell
        type="select"
        isLoading={isLoading}
        onValueChange={updater("assetType")}
        options={
          <>
            <Option value="VIRTUAL">
              <Badge variant={getAssetTypeVariant("VIRTUAL")}>
                {__("Virtual")}
              </Badge>
            </Option>
            <Option value="PHYSICAL">
              <Badge variant={getAssetTypeVariant("PHYSICAL")}>
                {__("Physical")}
              </Badge>
            </Option>
          </>
        }
      >
        <Badge variant={getAssetTypeVariant(entry.assetType)}>
          {entry.assetType === "PHYSICAL" ? __("Physical") : __("Virtual")}
        </Badge>
      </EditableCell>
      <EditableCell
        type="text"
        defaultValue={entry.amount}
        onValueChange={updater("amount")}
      />
      <EditableCell
        type="select"
        isLoading={isLoading}
        onValueChange={updater("owner")}
        options={<PeopleSelectOptions organizationId={organizationId} />}
      >
        {entry.owner?.fullName ?? __("Unassigned")}
      </EditableCell>
      <Cell>
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
      </Cell>
      <Cell className="text-end">
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
      </Cell>
    </>
  );
}
