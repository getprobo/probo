import {
  ActionDropdown,
  Avatar,
  Badge,
  Button,
  Cell,
  CellHead,
  DataTable,
  DropdownItem,
  EditableCell,
  IconCheckmark1,
  IconCrossLargeX,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  Row,
  RowButton,
  Spinner,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { usePageTitle, useToggle } from "@probo/hooks";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useParams } from "react-router";
import { CreateAssetDialog } from "./dialogs/CreateAssetDialog";
import {
  assetsQuery,
  updateAssetMutation,
  useCreateAsset,
  useDeleteAsset,
} from "../../../hooks/graph/AssetGraph";
import type { AssetGraphListQuery } from "/hooks/graph/__generated__/AssetGraphListQuery.graphql";
import { faviconUrl, getAssetTypeVariant } from "@probo/helpers";
import type { NodeOf } from "/types";
import type {
  AssetsPageFragment$data,
  AssetsPageFragment$key,
} from "./__generated__/AssetsPageFragment.graphql";
import { SnapshotBanner } from "/components/SnapshotBanner";
import {
  type MutationFieldUpdate,
  useMutateField,
} from "/hooks/useMutateField.tsx";
import type { UpdateAssetInput } from "/hooks/graph/__generated__/AssetGraphUpdateMutation.graphql.ts";
import z from "zod";
import { useStateWithSchema } from "/hooks/useStateWithSchema.ts";
import { usePeople } from "/hooks/graph/PeopleGraph.ts";
import { useVendors } from "/hooks/graph/VendorGraph.ts";
import clsx from "clsx";
import { Authorized } from "/permissions";
import { isAuthorized } from "/permissions";

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
  const { update } = useMutateField<UpdateAssetInput>(updateAssetMutation);
  const [showAdd, toggleAdd] = useToggle(false);

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
      <DataTable
        columns={[...Array.from({ length: 6 }).map(() => "1fr"), "56px"]}
      >
        <Row>
          <CellHead>{__("Name")}</CellHead>
          <CellHead>{__("Type")}</CellHead>
          <CellHead>{__("Data Types stored")}</CellHead>
          <CellHead>{__("Amount")}</CellHead>
          <CellHead>{__("Owner")}</CellHead>
          <CellHead>{__("Vendors")}</CellHead>
          <CellHead></CellHead>
        </Row>
        {assets.map((entry) => (
          <AssetRow
            key={entry.id}
            entry={entry}
            connectionId={connectionId}
            onUpdate={(field, value) => update(entry.id, field, value)}
          />
        ))}
        {showAdd ? (
          <AssetAddRow
            organizationId={organizationId}
            onSuccess={toggleAdd}
            connection={connectionId}
          />
        ) : (
          <RowButton onClick={toggleAdd}>{__("Add a new asset")}</RowButton>
        )}
      </DataTable>
    </div>
  );
}

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  amount: z.coerce.number().min(1, "Amount is required"),
  assetType: z.enum(["PHYSICAL", "VIRTUAL"]),
  ownerId: z.string().min(1, "Owner is required"),
  vendorIds: z.array(z.string()).optional(),
  dataTypesStored: z.string().min(1, "Data types stored is required"),
});

function AssetAddRow({
  organizationId,
  onSuccess,
  connection,
}: {
  organizationId: string;
  onSuccess: () => void;
  connection: string;
}) {
  const [value, setValue, errors] = useStateWithSchema(schema, {
    name: "",
    amount: 0,
    assetType: "VIRTUAL",
    ownerId: "",
    vendorIds: [],
    dataTypesStored: "",
  });

  const [createAsset, isMutating] = useCreateAsset(connection);

  const onSubmit = async () => {
    await createAsset({
      ...value,
      organizationId,
    });
    onSuccess();
  };

  return (
    <AssetRow
      // @ts-expect-error - TS doesn't know form value match schema
      onUpdate={setValue}
      onSubmit={onSubmit}
      errors={errors}
      loading={isMutating}
    />
  );
}

function AssetRow({
  entry,
  connectionId,
  onUpdate,
  onSubmit,
  errors,
  loading,
}: {
  entry?: AssetEntry;
  connectionId?: string;
  onUpdate: MutationFieldUpdate<UpdateAssetInput>;
  onSubmit?: () => void;
  errors?: Record<string, string>;
  loading?: boolean;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const deleteAsset = useDeleteAsset(entry, connectionId);
  const isOk = Object.keys(errors ?? {}).length === 0;
  return (
    <Row>
      <EditableCell
        type="text"
        value={entry?.name ?? ""}
        onValueChange={(v) => onUpdate("name", v)}
        blink={Boolean(errors?.name)}
      />
      <EditableCell
        type="select"
        items={["VIRTUAL", "PHYSICAL"]}
        value={entry?.assetType ?? "VIRTUAL"}
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
        value={entry?.dataTypesStored ?? ""}
        onValueChange={(v) => onUpdate("dataTypesStored", v)}
        blink={Boolean(errors?.dataTypeStored)}
      />
      <EditableCell
        type="text"
        value={entry?.amount.toString() ?? ""}
        onValueChange={(v) => onUpdate("amount", v)}
        blink={Boolean(errors?.amount)}
      />
      <EditableCell
        type="select"
        items={() => usePeople(organizationId, { excludeContractEnded: true })}
        value={entry?.owner}
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
        value={entry?.vendors.edges.map((edge) => edge.node)}
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
      <Cell className="text-end">
        {loading && (
          <Button
            disabled={true}
            variant="tertiary"
            className="text-txt-secondary"
          >
            <Spinner size={16} />
          </Button>
        )}
        {onSubmit && !loading && (
          <Button
            disabled={!isOk}
            variant="tertiary"
            className={clsx(isOk ? "text-txt-success" : "text-txt-secondary")}
            onClick={onSubmit}
          >
            <IconCheckmark1 size={16} />
          </Button>
        )}
        {!isSnapshotMode && entry && (
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
    </Row>
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
