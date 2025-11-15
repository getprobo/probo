import {
  ActionDropdown,
  Badge,
  Button,
  DropdownItem,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  SelectCell,
  TextCell,
  useConfirm,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import {
  graphql,
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { Link, useParams } from "react-router";
import {
  assetsQuery,
  createAssetMutation,
  deleteAssetMutation,
  updateAssetMutation,
} from "/hooks/graph/AssetGraph";
import type { AssetGraphListQuery } from "/hooks/graph/__generated__/AssetGraphListQuery.graphql";
import {
  getAssetTypeVariant,
  promisifyMutation,
  sprintf,
} from "@probo/helpers";
import type { AssetsPageFragment$key } from "./__generated__/AssetsPageFragment.graphql";
import { SnapshotBanner } from "/components/SnapshotBanner";
import z from "zod";
import { EditableTable } from "/components/table/EditableTable.tsx";
import { PeopleCell } from "/components/table/PeopleCell.tsx";
import { VendorsCell } from "/components/table/VendorsCell.tsx";
import { CreateAssetDialog } from "./dialogs/CreateAssetDialog";
import { Authorized, isAuthorized } from "/permissions";

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

type Props = {
  queryRef: PreloadedQuery<AssetGraphListQuery>;
};

const schema = z.object({
  name: z.string().trim().min(1, "Name is required"),
  amount: z.coerce.number().min(1, "Amount is required"),
  assetType: z.enum(["PHYSICAL", "VIRTUAL"]),
  ownerId: z.string().trim().min(1, "Owner is required"),
  vendorIds: z.array(z.string()).optional(),
  dataTypesStored: z.string().trim().min(1, "Data types stored is required"),
  organizationId: z.string().trim().min(1, "Organization is required"),
});

const defaultValue = {
  name: "",
  amount: 0,
  assetType: "VIRTUAL",
  ownerId: "",
  vendorIds: [],
  dataTypesStored: "",
  organizationId: "",
} satisfies z.infer<typeof schema>;

export default function AssetsPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const assetUrl = (entry: { id: string }) =>
    isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/assets/${entry.id}`
      : `/organizations/${organizationId}/assets/${entry.id}`;

  const data = usePreloadedQuery(assetsQuery, props.queryRef);
  const pagination = usePaginationFragment(
    paginatedAssetsFragment,
    data.node as AssetsPageFragment$key,
  );
  const assets = pagination.data.assets?.edges.map((edge) => edge.node);
  const connectionId = pagination.data.assets.__id;
  const deleteAsset = useDeleteAsset(connectionId);

  const hasAnyAction =
    !isSnapshotMode &&
    (isAuthorized(organizationId, "Asset", "updateAsset") ||
      isAuthorized(organizationId, "Asset", "deleteAsset"));
  usePageTitle(__("Assets"));

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
      <EditableTable
        connectionId={connectionId}
        pagination={pagination}
        items={assets}
        columns={[
          __("Name"),
          __("Type"),
          __("Data Types stored"),
          __("Amount"),
          __("Owner"),
          __("Vendors"),
        ]}
        schema={schema}
        updateMutation={updateAssetMutation}
        createMutation={createAssetMutation}
        addLabel={__("Add a new asset")}
        defaultValue={{
          ...defaultValue,
          organizationId,
        }}
        action={({ item }) =>
          hasAnyAction ? (
            <ActionDropdown>
              <DropdownItem asChild>
                <Link to={assetUrl(item)}>
                  <IconPencil size={16} />
                  {__("Edit")}
                </Link>
              </DropdownItem>
              <DropdownItem
                onClick={() => deleteAsset(item)}
                variant="danger"
                icon={IconTrashCan}
              >
                {__("Delete")}
              </DropdownItem>
            </ActionDropdown>
          ) : null
        }
        row={({ item }) => (
          <>
            <TextCell name="name" defaultValue={item?.name ?? ""} required />
            <SelectCell
              name="assetType"
              items={["VIRTUAL", "PHYSICAL"]}
              itemRenderer={({ item }) => (
                <Badge variant={getAssetTypeVariant(item ?? "VIRTUAL")}>
                  {item === "PHYSICAL" ? __("Physical") : __("Virtual")}
                </Badge>
              )}
              defaultValue={item?.assetType ?? defaultValue.assetType}
            />
            <TextCell
              name="dataTypesStored"
              defaultValue={
                item?.dataTypesStored ?? defaultValue.dataTypesStored
              }
              required
            />
            <TextCell
              name="amount"
              defaultValue={(item?.amount ?? defaultValue.amount).toString()}
              required
            />
            <PeopleCell
              name="ownerId"
              defaultValue={item?.owner}
              organizationId={organizationId}
            />
            <VendorsCell
              name="vendorIds"
              organizationId={organizationId}
              defaultValue={item?.vendors.edges.map((edge) => edge.node) ?? []}
            />
          </>
        )}
      />
    </div>
  );
}

const useDeleteAsset = (connectionId: string) => {
  const [mutate] = useMutation(deleteAssetMutation);
  const confirm = useConfirm();
  const { __ } = useTranslate();

  return (asset: { id: string; name: string }) => {
    if (!asset.id || !asset.name) {
      return alert(__("Failed to delete asset: missing id or name"));
    }
    confirm(
      () =>
        promisifyMutation(mutate)({
          variables: {
            input: {
              assetId: asset.id!,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete "%s". This action cannot be undone.',
          ),
          asset.name,
        ),
      },
    );
  };
};
