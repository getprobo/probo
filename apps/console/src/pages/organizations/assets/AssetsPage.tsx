import {
  ActionDropdown,
  Avatar,
  Badge,
  Button,
  DropdownItem,
  EditableCell,
  IconCrossLargeX,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
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
import { useParams } from "react-router";
import { CreateAssetDialog } from "./dialogs/CreateAssetDialog";
import {
  assetsQuery,
  createAssetMutation,
  deleteAssetMutation,
  updateAssetMutation,
} from "../../../hooks/graph/AssetGraph";
import type { AssetGraphListQuery } from "/hooks/graph/__generated__/AssetGraphListQuery.graphql";
import {
  faviconUrl,
  getAssetTypeVariant,
  promisifyMutation,
  sprintf,
} from "@probo/helpers";
import type { AssetsPageFragment$key } from "./__generated__/AssetsPageFragment.graphql";
import { SnapshotBanner } from "/components/SnapshotBanner";
import z from "zod";
import { usePeople } from "/hooks/graph/PeopleGraph.ts";
import { useVendors } from "/hooks/graph/VendorGraph.ts";
import { EditableTable } from "/components/table/EditableTable.tsx";
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
  amount: z.coerce.number().min(0, "Amount is required"),
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
        action={({ item }) => (
          <ActionDropdown>
            <DropdownItem
              onClick={() => deleteAsset(item)}
              variant="danger"
              icon={IconTrashCan}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
        row={({ item, onUpdate, errors }) => (
          <>
            <EditableCell
              type="text"
              value={item?.name ?? ""}
              onValueChange={(v) => onUpdate("name", v)}
              blink={Boolean(errors?.name)}
            />
            <EditableCell
              type="select"
              items={["VIRTUAL", "PHYSICAL"]}
              value={item?.assetType ?? "VIRTUAL"}
              itemRenderer={({ item }) => (
                <Badge variant={getAssetTypeVariant(item ?? "VIRTUAL")}>
                  {item === "PHYSICAL" ? __("Physical") : __("Virtual")}
                </Badge>
              )}
              onValueChange={(v) => onUpdate("assetType", v)}
              blink={Boolean(errors?.assetType)}
            />
            <EditableCell
              type="text"
              value={item?.dataTypesStored ?? ""}
              onValueChange={(v) => onUpdate("dataTypesStored", v)}
              blink={Boolean(errors?.dataTypesStored)}
            />
            <EditableCell
              type="text"
              value={item?.amount.toString() ?? "0"}
              onValueChange={(v) => onUpdate("amount", v)}
              blink={Boolean(errors?.amount)}
            />
            <EditableCell
              type="select"
              items={() =>
                usePeople(organizationId, { excludeContractEnded: true })
              }
              value={item?.owner}
              itemRenderer={({ item }) => (
                <div className="flex gap-2">
                  <Avatar name={item.fullName} />
                  {item.fullName}
                </div>
              )}
              onValueChange={(v) => onUpdate("ownerId", v.id)}
              blink={Boolean(errors?.ownerId)}
            />
            <EditableCell
              type="multiple"
              items={() => useVendors(organizationId)}
              value={item?.vendors.edges.map((edge) => edge.node)}
              itemRenderer={({ item, onRemove }) => (
                <VendorBadge key={item.id} vendor={item} onRemove={onRemove} />
              )}
              onValueChange={(v) =>
                onUpdate(
                  "vendorIds",
                  v.map((v) => v.id),
                )
              }
              blink={Boolean(errors?.vendorIds)}
            />
          </>
        )}
      />
    </div>
  );
}

type Vendor = {
  id: string;
  name: string;
  websiteUrl: string | null | undefined;
};

function VendorBadge({
  vendor,
  onRemove,
}: {
  vendor: Vendor;
  onRemove?: () => void;
}) {
  return (
    <Badge variant="neutral" className="flex items-center gap-1">
      <Avatar name={vendor.name} src={faviconUrl(vendor.websiteUrl)} size="s" />
      <span className="max-w-[100px] text-ellipsis overflow-hidden min-w-0 block">
        {vendor.name}
      </span>
      {onRemove && (
        <button
          onClick={onRemove}
          className="size-4 hover:text-txt-primary cursor-pointer"
        >
          <IconCrossLargeX size={14} />
        </button>
      )}
    </Badge>
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
