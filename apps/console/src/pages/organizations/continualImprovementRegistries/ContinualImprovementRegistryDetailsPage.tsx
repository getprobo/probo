import {
  ConnectionHandler,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import {
  continualImprovementRegistryNodeQuery,
  useDeleteContinualImprovementRegistry,
  useUpdateContinualImprovementRegistry,
  ContinualImprovementRegistriesConnectionKey,
} from "../../../hooks/graph/ContinualImprovementRegistryGraph";
import {
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  DropdownItem,
  Field,
  Option,
  Input,
  Card,
  Textarea,
  useToast,
  Select,
  Label,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { AuditSelectField } from "/components/form/AuditSelectField";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { Controller } from "react-hook-form";
import z from "zod";
import { getStatusVariant, getStatusLabel, formatDatetime } from "@probo/helpers";
import type { ContinualImprovementRegistryGraphNodeQuery } from "/hooks/graph/__generated__/ContinualImprovementRegistryGraphNodeQuery.graphql";

const updateRegistrySchema = z.object({
  referenceId: z.string().min(1, "Reference ID is required"),
  description: z.string().optional(),
  source: z.string().optional(),
  targetDate: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED"]),
  priority: z.enum(["LOW", "MEDIUM", "HIGH"]),
  ownerId: z.string().min(1, "Owner is required"),
  auditId: z.string().min(1, "Audit is required"),
});

type Props = {
  queryRef: PreloadedQuery<ContinualImprovementRegistryGraphNodeQuery>;
};

export default function ContinualImprovementRegistryDetailsPage(props: Props) {
  const data = usePreloadedQuery<ContinualImprovementRegistryGraphNodeQuery>(continualImprovementRegistryNodeQuery, props.queryRef);
  const registry = data.node;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();

  if (!registry) {
    return <div>{__("Continual improvement registry entry not found")}</div>;
  }

  const updateRegistry = useUpdateContinualImprovementRegistry();

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ContinualImprovementRegistriesConnectionKey
  );

  const deleteRegistry = useDeleteContinualImprovementRegistry({ id: registry.id!, referenceId: registry.referenceId! }, connectionId);

  const { register, handleSubmit, formState, control } = useFormWithSchema(
    updateRegistrySchema,
    {
      defaultValues: {
        referenceId: registry.referenceId || "",
        description: registry.description || "",
        source: registry.source || "",
        targetDate: registry.targetDate
          ? new Date(registry.targetDate).toISOString().split("T")[0]
          : "",
        status: registry.status || "OPEN",
        priority: registry.priority || "MEDIUM",
        ownerId: registry.owner?.id || "",
        auditId: registry.audit?.id || "",
      },
    }
  );

  const onSubmit = handleSubmit(async (formData) => {
    try {
      await updateRegistry({
        id: registry.id!,
        referenceId: formData.referenceId,
        description: formData.description || undefined,
        source: formData.source || undefined,
        targetDate: formatDatetime(formData.targetDate),
        status: formData.status,
        priority: formData.priority,
        ownerId: formData.ownerId,
        auditId: formData.auditId,
      });

      toast({
        title: __("Success"),
        description: __("Continual improvement registry entry updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: __("Failed to update continual improvement registry entry"),
        variant: "error",
      });
    }
  });

  const statusOptions = [
    { value: "OPEN", label: __("Open") },
    { value: "IN_PROGRESS", label: __("In Progress") },
    { value: "CLOSED", label: __("Closed") },
  ];

  const priorityOptions = [
    { value: "LOW", label: __("Low") },
    { value: "MEDIUM", label: __("Medium") },
    { value: "HIGH", label: __("High") },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Breadcrumb
          items={[
            { label: __("Continual Improvement Registries"), to: "../continual-improvement-registries" },
            { label: registry.referenceId! },
          ]}
        />
        <ActionDropdown>
          <DropdownItem onClick={deleteRegistry} variant="danger">
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
      </div>

      <Card>
        <div className="p-6">
          <div className="mb-6">
            <div className="flex items-center gap-4">
              <h1 className="text-2xl font-bold">{registry.referenceId}</h1>
              <Badge variant={getStatusVariant(registry.status || "OPEN")}>
                {getStatusLabel(registry.status || "OPEN")}
              </Badge>
              <Badge variant={registry.priority === "HIGH" ? "danger" : registry.priority === "MEDIUM" ? "warning" : "success"}>
                {registry.priority === "HIGH" ? __("High") : registry.priority === "MEDIUM" ? __("Medium") : __("Low")}
              </Badge>
            </div>
          </div>

          <form onSubmit={onSubmit} className="space-y-4">
            <Field
              label={__("Reference ID")}
              {...register("referenceId")}
              error={formState.errors.referenceId?.message}
              required
            />

            <AuditSelectField
              organizationId={organizationId}
              control={control}
              name="auditId"
              label={__("Audit")}
              error={formState.errors.auditId?.message}
              required
            />

            <div>
              <Label>{__("Description")}</Label>
              <Textarea
                {...register("description")}
                placeholder={__("Enter description")}
                rows={3}
              />
              {formState.errors.description?.message && (
                <div className="text-red-500 text-sm mt-1">
                  {formState.errors.description.message}
                </div>
              )}
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Field
                label={__("Source")}
                {...register("source")}
                error={formState.errors.source?.message}
              />

              <div>
                <Label>{__("Target Date")}</Label>
                <Input
                  type="date"
                  {...register("targetDate")}
                />
                {formState.errors.targetDate?.message && (
                  <div className="text-red-500 text-sm mt-1">
                    {formState.errors.targetDate.message}
                  </div>
                )}
              </div>
            </div>

            <PeopleSelectField
              organizationId={organizationId}
              control={control}
              name="ownerId"
              label={__("Owner")}
              error={formState.errors.ownerId?.message}
              required
            />

            <div className="grid grid-cols-2 gap-4">
              <Controller
                control={control}
                name="status"
                render={({ field }) => (
                  <div>
                    <Label>{__("Status")} *</Label>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                    >
                      {statusOptions.map((option) => (
                        <Option key={option.value} value={option.value}>
                          {option.label}
                        </Option>
                      ))}
                    </Select>
                    {formState.errors.status?.message && (
                      <div className="text-red-500 text-sm mt-1">
                        {formState.errors.status.message}
                      </div>
                    )}
                  </div>
                )}
              />

              <Controller
                control={control}
                name="priority"
                render={({ field }) => (
                  <div>
                    <Label>{__("Priority")} *</Label>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                    >
                      {priorityOptions.map((option) => (
                        <Option key={option.value} value={option.value}>
                          {option.label}
                        </Option>
                      ))}
                    </Select>
                    {formState.errors.priority?.message && (
                      <div className="text-red-500 text-sm mt-1">
                        {formState.errors.priority.message}
                      </div>
                    )}
                  </div>
                )}
              />
            </div>

            <div className="flex justify-end pt-4">
              <Button
                type="submit"
                variant="primary"
                disabled={formState.isSubmitting}
              >
                {formState.isSubmitting ? __("Saving...") : __("Save Changes")}
              </Button>
            </div>
          </form>
        </div>
      </Card>
    </div>
  );
}
