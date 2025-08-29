import {
  ConnectionHandler,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import {
  complianceRegistryNodeQuery,
  useDeleteComplianceRegistry,
  useUpdateComplianceRegistry,
  ComplianceRegistriesConnectionKey,
} from "../../../hooks/graph/ComplianceRegistryGraph";
import {
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  DropdownItem,
  Field,
  IconTrashCan,
  Option,
  Input,
  Card,
  Textarea,
  useToast,
  Select,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useParams } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { Controller } from "react-hook-form";
import z from "zod";
import { getStatusVariant, getStatusLabel, formatDatetime, getComplianceRegistryStatusOptions, validateSnapshotConsistency } from "@probo/helpers";
import { SnapshotBanner } from "/components/SnapshotBanner";
import type { ComplianceRegistryGraphNodeQuery } from "/hooks/graph/__generated__/ComplianceRegistryGraphNodeQuery.graphql";

const updateRegistrySchema = z.object({
  referenceId: z.string().min(1, "Reference ID is required"),
  area: z.string().optional(),
  source: z.string().optional(),
  requirement: z.string().optional(),
  actionsToBeImplemented: z.string().optional(),
  regulator: z.string().optional(),
  lastReviewDate: z.string().optional(),
  dueDate: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED"]),
  ownerId: z.string().min(1, "Owner is required"),
});

type Props = {
  queryRef: PreloadedQuery<ComplianceRegistryGraphNodeQuery>;
};

export default function ComplianceRegistryDetailsPage(props: Props) {
  const data = usePreloadedQuery<ComplianceRegistryGraphNodeQuery>(complianceRegistryNodeQuery, props.queryRef);
  const registry = data.node;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  if (!registry) {
    return <div>{__("Compliance registry entry not found")}</div>;
  }

  validateSnapshotConsistency(registry, snapshotId);

  const updateRegistry = useUpdateComplianceRegistry();
  const statusOptions = getComplianceRegistryStatusOptions(__);

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ComplianceRegistriesConnectionKey
  );

  const deleteRegistry = useDeleteComplianceRegistry({ id: registry.id!, referenceId: registry.referenceId! }, connectionId);

  const { register, handleSubmit, formState, control } = useFormWithSchema(
    updateRegistrySchema,
    {
      defaultValues: {
        referenceId: registry.referenceId || "",
        area: registry.area || "",
        source: registry.source || "",
        requirement: registry.requirement || "",
        actionsToBeImplemented: registry.actionsToBeImplemented || "",
        regulator: registry.regulator || "",
        lastReviewDate: registry.lastReviewDate
          ? new Date(registry.lastReviewDate).toISOString().split("T")[0]
          : "",
        dueDate: registry.dueDate
          ? new Date(registry.dueDate).toISOString().split("T")[0]
          : "",
        status: registry.status || "OPEN",
        ownerId: registry.owner?.id || "",
      },
    }
  );

  const onSubmit = handleSubmit(async (formData) => {
    try {
      await updateRegistry({
        id: registry.id!,
        referenceId: formData.referenceId,
        area: formData.area || undefined,
        source: formData.source || undefined,
        requirement: formData.requirement || undefined,
        actionsToBeImplemented: formData.actionsToBeImplemented || undefined,
        regulator: formData.regulator || undefined,
        lastReviewDate: formatDatetime(formData.lastReviewDate),
        dueDate: formatDatetime(formData.dueDate),
        status: formData.status,
        ownerId: formData.ownerId,

      });

      toast({
        title: __("Success"),
        description: __("Compliance registry entry updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: __("Failed to update compliance registry entry"),
        variant: "error",
      });
    }
  });

    const breadcrumbComplianceRegistriesUrl = isSnapshotMode
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/compliance-registries`
    : `/organizations/${organizationId}/compliance-registries`;

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <div className="flex justify-between items-start">
        <div>
                   <Breadcrumb
           items={[
             { label: __("Compliance Registries"), to: breadcrumbComplianceRegistriesUrl },
             { label: registry.referenceId! },
           ]}
         />
        <div className="flex items-center gap-3 mt-2">
          <h1 className="text-2xl font-bold">{registry.referenceId}</h1>
          <Badge variant={getStatusVariant(registry.status || "OPEN")}>
            {getStatusLabel(registry.status || "OPEN")}
          </Badge>
        </div>
      </div>

        {!isSnapshotMode && (
          <ActionDropdown>
            <DropdownItem icon={IconTrashCan} onClick={deleteRegistry}>
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <Card padded>
        <form onSubmit={onSubmit} className="space-y-6">
          <Field
            label={__("Reference ID")}
            error={formState.errors.referenceId?.message}
            required
          >
            <Input
              {...register("referenceId")}
              placeholder={__("Enter reference ID")}
              disabled={isSnapshotMode}
            />
          </Field>



          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field
              label={__("Area")}
              error={formState.errors.area?.message}
            >
              <Input
                {...register("area")}
                placeholder={__("Enter area")}
                disabled={isSnapshotMode}
              />
            </Field>

            <Field
              label={__("Source")}
              error={formState.errors.source?.message}
            >
              <Input
                {...register("source")}
                placeholder={__("Enter source")}
                disabled={isSnapshotMode}
              />
            </Field>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field label={__("Status")}>
              <Controller
                control={control}
                name="status"
                render={({ field }) => (
                  <Select
                    variant="editor"
                    placeholder={__("Select status")}
                    onValueChange={field.onChange}
                    value={field.value}
                    className="w-full"
                    disabled={isSnapshotMode}
                  >
                    {statusOptions.map((option) => (
                      <Option key={option.value} value={option.value}>
                        {option.label}
                      </Option>
                    ))}
                  </Select>
                )}
              />
              {formState.errors.status && (
                <p className="text-sm text-red-500 mt-1">{formState.errors.status.message}</p>
              )}
            </Field>

            <Controller
              name="ownerId"
              control={control}
              render={() => (
                <PeopleSelectField
                  organizationId={organizationId}
                  control={control}
                  name="ownerId"
                  label={__("Owner")}
                  error={formState.errors.ownerId?.message}
                  required
                  disabled={isSnapshotMode}
                />
              )}
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field
              label={__("Regulator")}
              error={formState.errors.regulator?.message}
            >
              <Input
                {...register("regulator")}
                placeholder={__("Enter regulator")}
                disabled={isSnapshotMode}
              />
            </Field>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field
              label={__("Last Review Date")}
              error={formState.errors.lastReviewDate?.message}
            >
              <Input
                {...register("lastReviewDate")}
                type="date"
                disabled={isSnapshotMode}
              />
            </Field>

            <Field
              label={__("Due Date")}
              error={formState.errors.dueDate?.message}
            >
              <Input
                {...register("dueDate")}
                type="date"
                disabled={isSnapshotMode}
              />
            </Field>
          </div>

          <Field
            label={__("Requirement")}
            error={formState.errors.requirement?.message}
          >
            <Textarea
              {...register("requirement")}
              placeholder={__("Enter requirement")}
              rows={4}
              disabled={isSnapshotMode}
            />
          </Field>

          <Field
            label={__("Actions to be Implemented")}
            error={formState.errors.actionsToBeImplemented?.message}
          >
            <Textarea
              {...register("actionsToBeImplemented")}
              placeholder={__("Enter actions to be implemented")}
              rows={4}
              disabled={isSnapshotMode}
            />
          </Field>

          {!isSnapshotMode && (
            <div className="flex justify-end">
              <Button
                type="submit"
                disabled={formState.isSubmitting}
              >
                {formState.isSubmitting ? __("Saving...") : __("Save Changes")}
              </Button>
            </div>
          )}
        </form>
      </Card>
    </div>
  );
}
