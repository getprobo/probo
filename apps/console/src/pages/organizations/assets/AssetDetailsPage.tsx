import {
  getAssetTypeVariant,
  validateSnapshotConsistency,
} from "@probo/helpers";
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
import { useParams } from "react-router";
import { z } from "zod";

import type { AssetGraphNodeQuery } from "#/__generated__/core/AssetGraphNodeQuery.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { VendorsMultiSelectField } from "#/components/form/VendorsMultiSelectField";
import { SnapshotBanner } from "#/components/SnapshotBanner";
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
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  validateSnapshotConsistency(assetEntry, snapshotId);

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    "AssetsPage_assets",
    { filter: { snapshotId: snapshotId || null } },
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

  const breadcrumbAssetsUrl
    = isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/assets`
      : `/organizations/${organizationId}/assets`;

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <Breadcrumb
        items={[
          {
            label: __("Assets"),
            to: breadcrumbAssetsUrl,
          },
          {
            label: assetEntry?.name ?? "",
          },
        ]}
      />

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
        {!isSnapshotMode && asset.node.canDelete && (
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
          disabled={isSnapshotMode}
        />

        <Field
          label={__("Amount")}
          {...register("amount", { valueAsNumber: true })}
          type="number"
          disabled={isSnapshotMode}
        />

        <ControlledField
          control={control}
          name="assetType"
          type="select"
          label={__("Asset Type")}
          disabled={isSnapshotMode}
        >
          <Option value="VIRTUAL">{__("Virtual")}</Option>
          <Option value="PHYSICAL">{__("Physical")}</Option>
        </ControlledField>

        <Field
          label={__("Data Types Stored")}
          {...register("dataTypesStored")}
          type="text"
          disabled={isSnapshotMode}
        />

        <PeopleSelectField
          organizationId={organizationId}
          control={control}
          name="ownerId"
          label={__("Owner")}
          disabled={isSnapshotMode}
        />

        <VendorsMultiSelectField
          organizationId={organizationId}
          control={control}
          name="vendorIds"
          selectedVendors={vendors}
          label={__("Vendors")}
          disabled={isSnapshotMode}
        />

        <div className="flex justify-end">
          {formState.isDirty && !isSnapshotMode && asset.node.canUpdate && (
            <Button type="submit" disabled={formState.isSubmitting}>
              {formState.isSubmitting ? __("Updating...") : __("Update")}
            </Button>
          )}
        </div>
      </form>
    </div>
  );
}
