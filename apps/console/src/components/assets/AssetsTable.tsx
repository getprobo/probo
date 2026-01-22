import {
  getAssetTypeVariant,
  promisifyMutation,
  sprintf,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  SelectCell,
  TextCell,
  useConfirm,
} from "@probo/ui";
import { useMutation } from "react-relay";
import type { usePaginationFragmentHookType } from "react-relay/relay-hooks/usePaginationFragment";
import { Link } from "react-router";
import type { OperationType } from "relay-runtime";
import { z } from "zod";

import type { AssetGraphDeleteMutation } from "#/__generated__/core/AssetGraphDeleteMutation.graphql";
import type {
  AssetsPageFragment$data,
  AssetsPageFragment$key,
} from "#/__generated__/core/AssetsPageFragment.graphql";
import {
  createAssetMutation,
  deleteAssetMutation,
  updateAssetMutation,
} from "#/hooks/graph/AssetGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { EditableTable } from "../table/EditableTable";
import { PeopleCell } from "../table/PeopleCell";
import { VendorsCell } from "../table/VendorsCell";

type Props = {
  connectionId: string;
  pagination: usePaginationFragmentHookType<
    OperationType,
    AssetsPageFragment$key,
    AssetsPageFragment$data
  >;
  assets: AssetsPageFragment$data["assets"]["edges"][0]["node"][];
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

export function AssetsTable(props: Props) {
  const { connectionId, pagination, assets } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const deleteAsset = useDeleteAsset(connectionId);

  return (
    <EditableTable
      pageSize={10}
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
          <DropdownItem asChild>
            <Link to={`/organizations/${organizationId}/assets/${item.id}`}>
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
      )}
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
            defaultValue={item?.dataTypesStored ?? defaultValue.dataTypesStored}
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
            defaultValue={item?.vendors?.edges?.map(edge => edge.node) ?? []}
          />
        </>
      )}
    />
  );
}

const useDeleteAsset = (connectionId: string) => {
  const [mutate] = useMutation<AssetGraphDeleteMutation>(deleteAssetMutation);
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
              assetId: asset.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete \"%s\". This action cannot be undone.",
          ),
          asset.name,
        ),
      },
    );
  };
};
