import { formatError, type GraphQLError } from "@probo/helpers";
import {
  formatDatetime,
  getStatusLabel,
  getStatusVariant,
  validateSnapshotConsistency,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  Card,
  DropdownItem,
  Field,
  Input,
  Label,
  Option,
  Select,
  Textarea,
  useToast,
} from "@probo/ui";
import { Controller } from "react-hook-form";
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";
import { z } from "zod";

import type { ContinualImprovementGraphNodeQuery } from "#/__generated__/core/ContinualImprovementGraphNodeQuery.graphql";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  continualImprovementNodeQuery,
  ContinualImprovementsConnectionKey,
  useDeleteContinualImprovement,
  useUpdateContinualImprovement,
} from "../../../hooks/graph/ContinualImprovementGraph";

const updateImprovementSchema = z.object({
  referenceId: z.string().min(1, "Reference ID is required"),
  description: z.string().optional(),
  source: z.string().optional(),
  targetDate: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED"]),
  priority: z.enum(["LOW", "MEDIUM", "HIGH"]),
  ownerId: z.string().min(1, "Owner is required"),
});

type Props = {
  queryRef: PreloadedQuery<ContinualImprovementGraphNodeQuery>;
};

export default function ContinualImprovementDetailsPage(props: Props) {
  const { node: improvement }
    = usePreloadedQuery<ContinualImprovementGraphNodeQuery>(
      continualImprovementNodeQuery,
      props.queryRef,
    );
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  validateSnapshotConsistency(improvement, snapshotId);

  const updateImprovement = useUpdateContinualImprovement();

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ContinualImprovementsConnectionKey,
    { filter: { snapshotId: snapshotId || null } },
  );

  const deleteImprovement = useDeleteContinualImprovement(
    { id: improvement.id!, referenceId: improvement.referenceId! },
    connectionId,
  );

  const { register, handleSubmit, formState, control } = useFormWithSchema(
    updateImprovementSchema,
    {
      defaultValues: {
        referenceId: improvement.referenceId || "",
        description: improvement.description || "",
        source: improvement.source || "",
        targetDate: improvement.targetDate
          ? new Date(improvement.targetDate).toISOString().split("T")[0]
          : "",
        status: improvement.status || "OPEN",
        priority: improvement.priority || "MEDIUM",
        ownerId: improvement.owner?.id || "",
      },
    },
  );

  const onSubmit = handleSubmit(async (formData) => {
    try {
      await updateImprovement({
        id: improvement.id!,
        referenceId: formData.referenceId,
        description: formData.description || undefined,
        source: formData.source || undefined,
        targetDate: formatDatetime(formData.targetDate) ?? null,
        status: formData.status,
        priority: formData.priority,
        ownerId: formData.ownerId,
      });

      toast({
        title: __("Success"),
        description: __("Continual improvement entry updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(
          __("Failed to update continual improvement"),
          error as GraphQLError,
        ),
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

  const breadcrumbImprovementsUrl = isSnapshotMode
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/continual-improvements`
    : `/organizations/${organizationId}/continual-improvements`;

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <div className="flex items-center justify-between">
        <Breadcrumb
          items={[
            {
              label: __("Continual Improvements"),
              to: breadcrumbImprovementsUrl,
            },
            { label: improvement.referenceId! },
          ]}
        />
        {!isSnapshotMode && improvement.canDelete && (
          <ActionDropdown>
            <DropdownItem onClick={deleteImprovement} variant="danger">
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <Card>
        <div className="p-6">
          <div className="mb-6">
            <div className="flex items-center gap-4">
              <h1 className="text-2xl font-bold">{improvement.referenceId}</h1>
              <Badge variant={getStatusVariant(improvement.status || "OPEN")}>
                {getStatusLabel(improvement.status || "OPEN")}
              </Badge>
              <Badge
                variant={
                  improvement.priority === "HIGH"
                    ? "danger"
                    : improvement.priority === "MEDIUM"
                      ? "warning"
                      : "success"
                }
              >
                {improvement.priority === "HIGH"
                  ? __("High")
                  : improvement.priority === "MEDIUM"
                    ? __("Medium")
                    : __("Low")}
              </Badge>
            </div>
          </div>

          <form onSubmit={e => void onSubmit(e)} className="space-y-4">
            <Field
              label={__("Reference ID")}
              {...register("referenceId")}
              error={formState.errors.referenceId?.message}
              readOnly={isSnapshotMode}
              required
            />

            <div>
              <Label>{__("Description")}</Label>
              <Textarea
                {...register("description")}
                placeholder={__("Enter description")}
                rows={3}
                readOnly={isSnapshotMode}
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
                readOnly={isSnapshotMode}
              />

              <div>
                <Label>{__("Target Date")}</Label>
                <Input
                  type="date"
                  {...register("targetDate")}
                  readOnly={isSnapshotMode}
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
              disabled={isSnapshotMode}
              required
            />

            <div className="grid grid-cols-2 gap-4">
              <Controller
                control={control}
                name="status"
                render={({ field }) => (
                  <div>
                    <Label>
                      {__("Status")}
                      {" "}
                      *
                    </Label>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      disabled={isSnapshotMode}
                    >
                      {statusOptions.map(option => (
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
                    <Label>
                      {__("Priority")}
                      {" "}
                      *
                    </Label>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      disabled={isSnapshotMode}
                    >
                      {priorityOptions.map(option => (
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

            {!isSnapshotMode && (
              <div className="flex justify-end pt-4">
                {improvement.canUpdate && (
                  <Button
                    type="submit"
                    variant="primary"
                    disabled={formState.isSubmitting}
                  >
                    {formState.isSubmitting
                      ? __("Saving...")
                      : __("Save Changes")}
                  </Button>
                )}
              </div>
            )}
          </form>
        </div>
      </Card>
    </div>
  );
}
