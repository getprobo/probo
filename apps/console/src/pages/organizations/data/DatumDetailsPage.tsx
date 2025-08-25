import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";
import {
  datumNodeQuery,
  useDeleteDatum,
  useUpdateDatum,
} from "/hooks/graph/DatumGraph";
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
import { useTranslate } from "@probo/i18n";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { ControlledField } from "/components/form/ControlledField";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { VendorsMultiSelectField } from "/components/form/VendorsMultiSelectField";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import z from "zod";
import { SnapshotBanner } from "/components/SnapshotBanner";
import type { DatumGraphNodeQuery } from "/hooks/graph/__generated__/DatumGraphNodeQuery.graphql";

const updateDatumSchema = z.object({
  name: z.string().min(1, "Name is required"),
  dataClassification: z.enum(["PUBLIC", "INTERNAL", "CONFIDENTIAL", "SECRET"]),
  ownerId: z.string().min(1, "Owner is required"),
  vendorIds: z.array(z.string()).optional(),
});

type Props = {
  queryRef: PreloadedQuery<DatumGraphNodeQuery>;
};

export default function DatumDetailsPage(props: Props) {
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const queryData = usePreloadedQuery<DatumGraphNodeQuery>(
    datumNodeQuery,
    props.queryRef
  );

  const datumEntry = queryData.node;

  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const deleteDatum = useDeleteDatum(
    datumEntry,
    ConnectionHandler.getConnectionID(organizationId, "DataPage_data"),
  );

  const vendors = datumEntry?.vendors?.edges.map(edge => edge.node) ?? [];
  const vendorIds = vendors.map(vendor => vendor.id);

  const { control, formState, handleSubmit, register, reset } =
    useFormWithSchema(updateDatumSchema, {
      defaultValues: {
        name: datumEntry?.name || "",
        dataClassification: datumEntry?.dataClassification || "PUBLIC",
        ownerId: datumEntry?.owner?.id || "",
        vendorIds: vendorIds,
      },
    });

  const updateDatum = useUpdateDatum();

  const onSubmit = handleSubmit(async (formData) => {
    if (!datumEntry?.id) {
      alert("id is missing from data");
      return;
    }
    try {
      await updateDatum({
        id: datumEntry.id,
        ...formData,
      });
      reset(formData);
    } catch (error) {
      console.error("Failed to update datum:", error);
    }
  });

  const breadcrumbDataUrl = isSnapshotMode
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/data`
    : `/organizations/${organizationId}/data`;

  const breadcrumbItems = [
    {
      label: __("Data"),
      to: breadcrumbDataUrl,
    },
    {
      label: datumEntry?.name || "",
    },
  ];

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <Breadcrumb items={breadcrumbItems} />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="text-2xl">{datumEntry?.name}</div>
          <Badge variant="info">{datumEntry?.dataClassification}</Badge>
        </div>
        {!isSnapshotMode && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={deleteDatum}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <form onSubmit={onSubmit} className="space-y-6 max-w-2xl">
        <Field
          label={__("Name")}
          {...register("name")}
          type="text"
          disabled={isSnapshotMode}
        />

        <ControlledField
          control={control}
          name="dataClassification"
          type="select"
          label={__("Classification")}
          disabled={isSnapshotMode}
        >
          <Option value="PUBLIC">{__("Public")}</Option>
          <Option value="INTERNAL">{__("Internal")}</Option>
          <Option value="CONFIDENTIAL">{__("Confidential")}</Option>
          <Option value="SECRET">{__("Secret")}</Option>
        </ControlledField>

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
          label={__("Vendors")}
          disabled={isSnapshotMode}
          selectedVendors={vendors}
        />

        {!isSnapshotMode && (
          <div className="flex justify-end">
            {formState.isDirty && (
              <Button type="submit" disabled={formState.isSubmitting}>
                {formState.isSubmitting ? __("Updating...") : __("Update")}
              </Button>
            )}
          </div>
        )}
      </form>
    </div>
  );
}
