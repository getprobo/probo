import {
  formatDatetime,
  formatError,
  getStatusLabel,
  getStatusOptions,
  getStatusVariant,
  type GraphQLError,
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
  IconTrashCan,
  Input,
  Option,
  Textarea,
  useToast,
} from "@probo/ui";
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";
import { z } from "zod";

import type { NonconformityGraphNodeQuery } from "#/__generated__/core/NonconformityGraphNodeQuery.graphql";
import { AuditSelectField } from "#/components/form/AuditSelectField";
import { ControlledField } from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  NonconformitiesConnectionKey,
  nonconformityNodeQuery,
  useDeleteNonconformity,
  useUpdateNonconformity,
} from "../../../hooks/graph/NonconformityGraph";

const updateNonconformitySchema = z.object({
  referenceId: z.string().min(1, "Reference ID is required"),
  description: z.string().optional(),
  dateIdentified: z.string().optional(),
  dueDate: z.string().optional(),
  rootCause: z.string().min(1, "Root cause is required"),
  correctiveAction: z.string().optional(),
  effectivenessCheck: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED"]),
  ownerId: z.string().min(1, "Owner is required"),
  auditId: z.string().optional(),
});

type Props = {
  queryRef: PreloadedQuery<NonconformityGraphNodeQuery>;
};

export default function NonconformityDetailsPage(props: Props) {
  const { node: nonconformity }
    = usePreloadedQuery<NonconformityGraphNodeQuery>(
      nonconformityNodeQuery,
      props.queryRef,
    );
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  validateSnapshotConsistency(nonconformity, snapshotId);

  const deleteNonconformity = useDeleteNonconformity(
    { id: nonconformity.id!, referenceId: nonconformity.referenceId! },
    ConnectionHandler.getConnectionID(
      organizationId,
      NonconformitiesConnectionKey,
      { filter: { snapshotId: snapshotId || null } },
    ),
  );

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(updateNonconformitySchema, {
      defaultValues: {
        referenceId: nonconformity.referenceId || "",
        description: nonconformity.description || "",
        dateIdentified: nonconformity.dateIdentified?.split("T")[0] || "",
        dueDate: nonconformity.dueDate?.split("T")[0] || "",
        rootCause: nonconformity.rootCause || "",
        correctiveAction: nonconformity.correctiveAction || "",
        effectivenessCheck: nonconformity.effectivenessCheck || "",
        status: nonconformity.status || "OPEN",
        ownerId: nonconformity.owner?.id || "",
        auditId: nonconformity.audit?.id || "",
      },
    });

  const updateNonconformity = useUpdateNonconformity();
  const { toast } = useToast();

  const onSubmit = handleSubmit(async (formData) => {
    if (!nonconformity.id) return;

    try {
      await updateNonconformity({
        id: nonconformity.id,
        referenceId: formData.referenceId,
        description: formData.description,
        dateIdentified: formatDatetime(formData.dateIdentified) ?? null,
        dueDate: formatDatetime(formData.dueDate) ?? null,
        rootCause: formData.rootCause,
        correctiveAction: formData.correctiveAction,
        effectivenessCheck: formData.effectivenessCheck,
        status: formData.status,
        ownerId: formData.ownerId,
        auditId: formData.auditId,
      });
      reset(formData);
      toast({
        title: __("Success"),
        description: __("Nonconformity updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(
          __("Failed to update nonconformity"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  const statusOptions = getStatusOptions(__);

  const breadcrumbNonconformitiesUrl = isSnapshotMode
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/nonconformities`
    : `/organizations/${organizationId}/nonconformities`;

  return (
    <div className="space-y-6">
      {isSnapshotMode && <SnapshotBanner snapshotId={snapshotId!} />}
      <Breadcrumb
        items={[
          {
            label: __("Nonconformities"),
            to: breadcrumbNonconformitiesUrl,
          },
          {
            label: nonconformity.referenceId || __("Unknown Nonconformity"),
          },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="text-2xl font-semibold">
            {nonconformity.referenceId}
          </div>
          <Badge variant={getStatusVariant(nonconformity.status || "OPEN")}>
            {getStatusLabel(nonconformity.status || "OPEN")}
          </Badge>
        </div>
        {!isSnapshotMode && (
          <ActionDropdown variant="secondary">
            {nonconformity.canDelete && (
              <DropdownItem
                variant="danger"
                icon={IconTrashCan}
                onClick={deleteNonconformity}
              >
                {__("Delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        )}
      </div>

      <div className="max-w-4xl">
        <Card padded>
          <form onSubmit={e => void onSubmit(e)} className="space-y-6">
            <Field
              label={__("Reference ID")}
              required
              error={formState.errors.referenceId?.message}
            >
              <Input
                {...register("referenceId")}
                placeholder={__("Enter reference ID")}
                disabled={isSnapshotMode}
              />
            </Field>

            <AuditSelectField
              organizationId={organizationId}
              control={control}
              name="auditId"
              label={__("Audit")}
              error={formState.errors.auditId?.message}
              disabled={isSnapshotMode}
            />

            <Field label={__("Description")}>
              <Textarea
                {...register("description")}
                placeholder={__("Enter description")}
                rows={3}
                disabled={isSnapshotMode}
              />
            </Field>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <ControlledField
                control={control}
                name="status"
                type="select"
                label={__("Status")}
                required
                disabled={isSnapshotMode}
              >
                {statusOptions.map(option => (
                  <Option key={option.value} value={option.value}>
                    {option.label}
                  </Option>
                ))}
              </ControlledField>

              <PeopleSelectField
                organizationId={organizationId}
                control={control}
                name="ownerId"
                label={__("Owner")}
                error={formState.errors.ownerId?.message}
                required
                disabled={isSnapshotMode}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field label={__("Date Identified")}>
                <Input
                  {...register("dateIdentified")}
                  type="date"
                  disabled={isSnapshotMode}
                />
              </Field>

              <Field label={__("Due Date")}>
                <Input
                  {...register("dueDate")}
                  type="date"
                  disabled={isSnapshotMode}
                />
              </Field>
            </div>

            <Field
              label={__("Root Cause")}
              required
              error={formState.errors.rootCause?.message}
            >
              <Textarea
                {...register("rootCause")}
                placeholder={__("Enter root cause")}
                rows={3}
                disabled={isSnapshotMode}
              />
            </Field>

            <Field label={__("Corrective Action")}>
              <Textarea
                {...register("correctiveAction")}
                placeholder={__("Enter corrective action")}
                rows={3}
                disabled={isSnapshotMode}
              />
            </Field>

            <Field label={__("Effectiveness Check")}>
              <Textarea
                {...register("effectivenessCheck")}
                placeholder={__("Enter effectiveness check details")}
                rows={3}
                disabled={isSnapshotMode}
              />
            </Field>

            <div className="flex justify-end">
              {formState.isDirty
                && !isSnapshotMode
                && nonconformity.canUpdate && (
                <Button type="submit" disabled={formState.isSubmitting}>
                  {formState.isSubmitting ? __("Updating...") : __("Update")}
                </Button>
              )}
            </div>
          </form>
        </Card>
      </div>
    </div>
  );
}
