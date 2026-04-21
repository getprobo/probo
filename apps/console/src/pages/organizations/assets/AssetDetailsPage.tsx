// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { getAssetTypeVariant } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { z } from "zod";

import type { AssetGraphNodeQuery } from "#/__generated__/core/AssetGraphNodeQuery.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { VendorsMultiSelectField } from "#/components/form/VendorsMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  assetNodeQuery,
  useDeleteAsset,
  useUpdateAsset,
} from "../../../hooks/graph/AssetGraph";

const updateAssetSchema = z.object({
  name: z.string().min(1, "Name is required"),
  amount: z.number().min(1, "Amount is required"),
  assetType: z.enum(["PHYSICAL", "VIRTUAL"]),
  dataTypesStored: z.string().min(1, "Data types stored is required"),
  ownerId: z.string().min(1, "Owner is required"),
  vendorIds: z.array(z.string()).optional(),
});

type Props = {
  queryRef: PreloadedQuery<AssetGraphNodeQuery>;
};

export default function AssetDetailsPage(props: Props) {
  const asset = usePreloadedQuery<AssetGraphNodeQuery>(
    assetNodeQuery,
    props.queryRef,
  );
  const assetEntry = asset.node;
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    "AssetsPage_assets",
  );
  const deleteAsset = useDeleteAsset(assetEntry, connectionId);

  const vendors = assetEntry.vendors?.edges.map(edge => edge.node) ?? [];
  const vendorIds = vendors.map(vendor => vendor.id);

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(updateAssetSchema, {
      defaultValues: {
        name: assetEntry.name || "",
        amount: assetEntry.amount || 0,
        assetType: assetEntry.assetType || "VIRTUAL",
        dataTypesStored: assetEntry.dataTypesStored || "",
        ownerId: assetEntry.owner?.id || "",
        vendorIds: vendorIds,
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
      label: __("Assets"),
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
              ? __("Physical")
              : __("Virtual")}
          </Badge>
        </div>
        {asset.node.canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={deleteAsset}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <form onSubmit={e => void onSubmit(e)} className="space-y-6 max-w-2xl">
        <Field
          label={__("Name")}
          {...register("name")}
          type="text"
          disabled={!assetEntry.canUpdate}
        />

        <Field
          label={__("Amount")}
          {...register("amount", { valueAsNumber: true })}
          type="number"
          disabled={!assetEntry.canUpdate}
        />

        <ControlledField
          control={control}
          name="assetType"
          type="select"
          label={__("Asset Type")}
          disabled={!assetEntry.canUpdate}
        >
          <Option value="VIRTUAL">{__("Virtual")}</Option>
          <Option value="PHYSICAL">{__("Physical")}</Option>
        </ControlledField>

        <Field
          label={__("Data Types Stored")}
          {...register("dataTypesStored")}
          type="text"
          disabled={!assetEntry.canUpdate}
        />

        <PeopleSelectField
          organizationId={organizationId}
          control={control}
          name="ownerId"
          label={__("Owner")}
          disabled={!assetEntry.canUpdate}
        />

        <VendorsMultiSelectField
          organizationId={organizationId}
          control={control}
          name="vendorIds"
          selectedVendors={vendors}
          label={__("Vendors")}
          disabled={!assetEntry.canUpdate}
        />

        <div className="flex justify-end">
          {formState.isDirty && assetEntry.canUpdate && (
            <Button type="submit" disabled={formState.isSubmitting}>
              {formState.isSubmitting ? __("Updating...") : __("Update")}
            </Button>
          )}
        </div>
      </form>
    </div>
  );
}
