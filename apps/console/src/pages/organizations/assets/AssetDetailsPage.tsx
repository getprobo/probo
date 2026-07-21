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

import { getAssetTypeVariant } from "@probo/helpers";
import {
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  DropdownItem,
  Field,
  IconTrashCan,
  Option,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { z } from "zod";

import type { AssetGraphNodeQuery } from "#/__generated__/core/AssetGraphNodeQuery.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { ThirdPartiesMultiSelectField } from "#/components/form/ThirdPartiesMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  assetNodeQuery,
  useDeleteAsset,
  useUpdateAsset,
} from "../../../hooks/graph/AssetGraph";

type Props = {
  queryRef: PreloadedQuery<AssetGraphNodeQuery>;
};

export default function AssetDetailsPage(props: Props) {
  const asset = usePreloadedQuery<AssetGraphNodeQuery>(
    assetNodeQuery,
    props.queryRef,
  );
  const assetEntry = asset.node;
  const { t } = useTranslation();
  const organizationId = useOrganizationId();

  const updateAssetSchema = z.object({
    name: z.string().min(1, t("assetDetailsPage.validation.nameRequired")),
    amount: z.number().min(1, t("assetDetailsPage.validation.amountRequired")),
    assetType: z.enum(["PHYSICAL", "VIRTUAL"]),
    dataTypesStored: z.string().min(
      1,
      t("assetDetailsPage.validation.dataTypesStoredRequired"),
    ),
    ownerId: z.string().min(1, t("assetDetailsPage.validation.ownerRequired")),
    thirdPartyIds: z.array(z.string()).optional(),
  });

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    "AssetsPage_assets",
  );
  const deleteAsset = useDeleteAsset(assetEntry, connectionId);

  const thirdParties = assetEntry.thirdParties?.edges.map(edge => edge.node) ?? [];
  const thirdPartyIds = thirdParties.map(thirdParty => thirdParty.id);

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(updateAssetSchema, {
      defaultValues: {
        name: assetEntry.name || "",
        amount: assetEntry.amount || 0,
        assetType: assetEntry.assetType || "VIRTUAL",
        dataTypesStored: assetEntry.dataTypesStored || "",
        ownerId: assetEntry.owner?.id || "",
        thirdPartyIds: thirdPartyIds,
      },
    });

  const updateAsset = useUpdateAsset();

  const onSubmit = handleSubmit(async (formData) => {
    await updateAsset({
      id: assetEntry.id!,
      ...formData,
    });
    reset(formData);
  });

  const breadcrumbItems = [
    {
      label: t("assetDetailsPage.breadcrumb.assets"),
      to: `/organizations/${organizationId}/assets`,
    },
    {
      label: assetEntry?.name ?? "",
    },
  ];

  return (
    <div className="space-y-6">
      <Breadcrumb items={breadcrumbItems} />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="text-2xl">{assetEntry?.name}</div>
          <Badge
            variant={getAssetTypeVariant(assetEntry?.assetType ?? "VIRTUAL")}
          >
            {assetEntry?.assetType === "PHYSICAL"
              ? t("assetDetailsPage.assetTypes.physical")
              : t("assetDetailsPage.assetTypes.virtual")}
          </Badge>
        </div>
        {asset.node.canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={deleteAsset}
            >
              {t("assetDetailsPage.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <form onSubmit={e => void onSubmit(e)} className="space-y-6 max-w-2xl">
        <Field
          label={t("assetDetailsPage.fields.name")}
          {...register("name")}
          type="text"
          disabled={!assetEntry.canUpdate}
        />

        <Field
          label={t("assetDetailsPage.fields.amount")}
          {...register("amount", { valueAsNumber: true })}
          type="number"
          disabled={!assetEntry.canUpdate}
        />

        <ControlledField
          control={control}
          name="assetType"
          type="select"
          label={t("assetDetailsPage.fields.assetType")}
          disabled={!assetEntry.canUpdate}
        >
          <Option value="VIRTUAL">
            {t("assetDetailsPage.assetTypes.virtual")}
          </Option>
          <Option value="PHYSICAL">
            {t("assetDetailsPage.assetTypes.physical")}
          </Option>
        </ControlledField>

        <Field
          label={t("assetDetailsPage.fields.dataTypesStored")}
          {...register("dataTypesStored")}
          type="text"
          disabled={!assetEntry.canUpdate}
        />

        <PeopleSelectField
          organizationId={organizationId}
          control={control}
          name="ownerId"
          label={t("assetDetailsPage.fields.owner")}
          disabled={!assetEntry.canUpdate}
        />

        <ThirdPartiesMultiSelectField
          organizationId={organizationId}
          control={control}
          name="thirdPartyIds"
          selectedThirdParties={thirdParties}
          label={t("assetDetailsPage.fields.thirdParties")}
          disabled={!assetEntry.canUpdate}
        />

        <div className="flex justify-end">
          {formState.isDirty && assetEntry.canUpdate && (
            <Button type="submit" disabled={formState.isSubmitting}>
              {formState.isSubmitting
                ? t("assetDetailsPage.actions.updating")
                : t("assetDetailsPage.actions.update")}
            </Button>
          )}
        </div>
      </form>
    </div>
  );
}
