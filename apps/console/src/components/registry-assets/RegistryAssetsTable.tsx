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
  getRegistryAssetTypeVariant,
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

import type { RegistryAssetGraphDeleteMutation } from "#/__generated__/core/RegistryAssetGraphDeleteMutation.graphql";
import type {
  RegistryAssetsPageFragment$data,
  RegistryAssetsPageFragment$key,
} from "#/__generated__/core/RegistryAssetsPageFragment.graphql";
import {
  createRegistryAssetMutation,
  deleteRegistryAssetMutation,
  updateRegistryAssetMutation,
} from "#/hooks/graph/RegistryAssetGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { z } from "#/lib/zod";

import { EditableTable } from "../table/EditableTable";
import { PeopleCell } from "../table/PeopleCell";
import { ThirdPartiesCell } from "../table/ThirdPartiesCell";

type Props = {
  connectionId: string;
  pagination: usePaginationFragmentHookType<
    OperationType,
    RegistryAssetsPageFragment$key,
    RegistryAssetsPageFragment$data
  >;
  assets: RegistryAssetsPageFragment$data["registryAssets"]["edges"][0]["node"][];
};

const schema = z.object({
  name: z.string().trim().min(1, "Name is required"),
  amount: z.coerce.number().min(1, "Amount is required"),
  registryAssetType: z.enum(["PHYSICAL", "VIRTUAL"]),
  ownerId: z.string().trim().min(1, "Owner is required"),
  thirdPartyIds: z.array(z.string()).optional(),
  dataTypesStored: z.string().trim().min(1, "Data types stored is required"),
  organizationId: z.string().trim().min(1, "Organization is required"),
});

const defaultValue = {
  name: "",
  amount: 0,
  registryAssetType: "VIRTUAL",
  ownerId: "",
  thirdPartyIds: [],
  dataTypesStored: "",
  organizationId: "",
} satisfies z.infer<typeof schema>;

export function RegistryAssetsTable(props: Props) {
  const { connectionId, pagination, assets } = props;

  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const deleteRegistryAsset = useDeleteRegistryAsset(connectionId);

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
      updateMutation={updateRegistryAssetMutation}
      createMutation={createRegistryAssetMutation}
      addLabel={t("assetsTable.actions.add")}
      defaultValue={{
        ...defaultValue,
        organizationId,
      }}
      action={({ item }) => (
        <ActionDropdown>
          <DropdownItem asChild>
            <Link to={`/organizations/${organizationId}/registries/assets/${item.id}`}>
              <IconPencil size={16} />
              {t("assetsTable.actions.edit")}
            </Link>
          </DropdownItem>
          <DropdownItem
            onClick={() => deleteRegistryAsset(item)}
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
            name="registryAssetType"
            items={["VIRTUAL", "PHYSICAL"]}
            itemRenderer={({ item }) => (
              <Badge variant={getRegistryAssetTypeVariant(item ?? "VIRTUAL")}>
                {item === "PHYSICAL"
                  ? t("assetsTable.assetTypes.physical")
                  : t("assetsTable.assetTypes.virtual")}
              </Badge>
            )}
            defaultValue={item?.registryAssetType ?? defaultValue.registryAssetType}
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

const useDeleteRegistryAsset = (connectionId: string) => {
  const [mutate] = useMutation<RegistryAssetGraphDeleteMutation>(deleteRegistryAssetMutation);
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
              registryAssetId: asset.id,
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
