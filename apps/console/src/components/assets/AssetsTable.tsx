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
  getAssetTypeVariant,
  promisifyMutation,
} from "@probo/helpers";
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
import { useTranslation } from "react-i18next";
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
import { ThirdPartiesCell } from "../table/ThirdPartiesCell";

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
  thirdPartyIds: z.array(z.string()).optional(),
  dataTypesStored: z.string().trim().min(1, "Data types stored is required"),
  organizationId: z.string().trim().min(1, "Organization is required"),
});

const defaultValue = {
  name: "",
  amount: 0,
  assetType: "VIRTUAL",
  ownerId: "",
  thirdPartyIds: [],
  dataTypesStored: "",
  organizationId: "",
} satisfies z.infer<typeof schema>;

export function AssetsTable(props: Props) {
  const { connectionId, pagination, assets } = props;

  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const deleteAsset = useDeleteAsset(connectionId);

  return (
    <EditableTable
      pageSize={10}
      connectionId={connectionId}
      pagination={pagination}
      items={assets}
      columns={[
        t("assetsTable.columns.name"),
        t("assetsTable.columns.type"),
        t("assetsTable.columns.dataTypesStored"),
        t("assetsTable.columns.amount"),
        t("assetsTable.columns.owner"),
        t("assetsTable.columns.thirdParties"),
      ]}
      schema={schema}
      updateMutation={updateAssetMutation}
      createMutation={createAssetMutation}
      addLabel={t("assetsTable.actions.add")}
      defaultValue={{
        ...defaultValue,
        organizationId,
      }}
      action={({ item }) => (
        <ActionDropdown>
          <DropdownItem asChild>
            <Link to={`/organizations/${organizationId}/assets/${item.id}`}>
              <IconPencil size={16} />
              {t("assetsTable.actions.edit")}
            </Link>
          </DropdownItem>
          <DropdownItem
            onClick={() => deleteAsset(item)}
            variant="danger"
            icon={IconTrashCan}
          >
            {t("assetsTable.actions.delete")}
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
                {item === "PHYSICAL"
                  ? t("assetsTable.assetTypes.physical")
                  : t("assetsTable.assetTypes.virtual")}
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
          <ThirdPartiesCell
            name="thirdPartyIds"
            organizationId={organizationId}
            defaultValue={item?.thirdParties?.edges?.map(edge => edge.node) ?? []}
          />
        </>
      )}
    />
  );
}

const useDeleteAsset = (connectionId: string) => {
  const [mutate] = useMutation<AssetGraphDeleteMutation>(deleteAssetMutation);
  const confirm = useConfirm();
  const { t } = useTranslation();

  return (asset: { id: string; name: string }) => {
    if (!asset.id || !asset.name) {
      return alert(t("assetsTable.delete.missingIdOrName"));
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
        message: t("assetsTable.delete.confirmation", { name: asset.name }),
      },
    );
  };
};
