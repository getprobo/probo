import {
  ConnectionHandler,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import {
  obligationNodeQuery,
  useDeleteObligation,
  useUpdateObligation,
  ObligationsConnectionKey,
} from "../../../hooks/graph/ObligationGraph";
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
import { formatError, type GraphQLError } from "@probo/helpers";
import z from "zod";
import { getObligationStatusVariant, getObligationStatusLabel, formatDatetime, getObligationStatusOptions, validateSnapshotConsistency } from "@probo/helpers";
import { SnapshotBanner } from "/components/SnapshotBanner";
import type { ObligationGraphNodeQuery } from "/hooks/graph/__generated__/ObligationGraphNodeQuery.graphql";
import { Authorized } from "/permissions";

const updateObligationSchema = z.object({
  area: z.string().optional(),
  source: z.string().optional(),
  requirement: z.string().optional(),
  actionsToBeImplemented: z.string().optional(),
  regulator: z.string().optional(),
  lastReviewDate: z.string().optional(),
  dueDate: z.string().optional(),
  status: z.enum(["NON_COMPLIANT", "PARTIALLY_COMPLIANT", "COMPLIANT"]),
  ownerId: z.string().min(1, "Owner is required"),
});

type Props = {
  queryRef: PreloadedQuery<ObligationGraphNodeQuery>;
};

export default function ObligationDetailsPage(props: Props) {
  const data = usePreloadedQuery<ObligationGraphNodeQuery>(obligationNodeQuery, props.queryRef);
  const obligation = data.node;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  if (!obligation) {
    return <div>{__("Obligation not found")}</div>;
  }

  validateSnapshotConsistency(obligation, snapshotId);

  const updateObligation = useUpdateObligation();
  const statusOptions = getObligationStatusOptions(__);

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    ObligationsConnectionKey
  );

  const deleteObligation = useDeleteObligation(
    { id: obligation?.id! },
    connectionId
  );

  const { register, handleSubmit, formState, control } = useFormWithSchema(
    updateObligationSchema,
    {
      defaultValues: {
        area: obligation?.area || "",
        source: obligation?.source || "",
        requirement: obligation?.requirement || "",
        actionsToBeImplemented: obligation?.actionsToBeImplemented || "",
        regulator: obligation?.regulator || "",
        lastReviewDate: obligation?.lastReviewDate
          ? new Date(obligation.lastReviewDate).toISOString().split("T")[0]
          : "",
        dueDate: obligation?.dueDate
          ? new Date(obligation.dueDate).toISOString().split("T")[0]
          : "",
        status: obligation?.status ?? "NON_COMPLIANT",
        ownerId: obligation?.owner?.id || "",
      },
    }
  );

  const onSubmit = handleSubmit(async (formData) => {
    try {
      await updateObligation({
        id: obligation.id!,
        area: formData.area || undefined,
        source: formData.source || undefined,
        requirement: formData.requirement || undefined,
        actionsToBeImplemented: formData.actionsToBeImplemented || undefined,
        regulator: formData.regulator || undefined,
        lastReviewDate: formatDatetime(formData.lastReviewDate) ?? null,
        dueDate: formatDatetime(formData.dueDate) ?? null,
        status: formData.status,
        ownerId: formData.ownerId,
      });

      toast({
        title: __("Success"),
        description: __("Obligation updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to update obligation"), error as GraphQLError),
        variant: "error",
      });
    }
  });

    const breadcrumbObligationsUrl = isSnapshotMode
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/obligations`
    : `/organizations/${organizationId}/obligations`;

  return (
    <div className="space-y-6">
      {isSnapshotMode && snapshotId && (
        <SnapshotBanner snapshotId={snapshotId} />
      )}
      <div className="flex justify-between items-start">
        <div>
                   <Breadcrumb
           items={[
             { label: __("Obligations"), to: breadcrumbObligationsUrl },
             { label: __("Obligation Details") },
           ]}
         />
        <div className="flex items-center gap-3 mt-2">
          <h1 className="text-2xl font-bold">{__("Obligation")}</h1>
          <Badge variant={getObligationStatusVariant(obligation.status ?? "NON_COMPLIANT")}>
            {getObligationStatusLabel(obligation.status ?? "NON_COMPLIANT")}
          </Badge>
        </div>
      </div>

        {!isSnapshotMode && (
          <Authorized entity="Obligation" action="deleteObligation">
            <ActionDropdown>
              <DropdownItem icon={IconTrashCan} onClick={deleteObligation}>
                {__("Delete")}
              </DropdownItem>
            </ActionDropdown>
          </Authorized>
        )}
      </div>

      <Card padded>
        <form onSubmit={onSubmit} className="space-y-6">

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
              <Authorized entity="Obligation" action="updateObligation">
                <Button
                  type="submit"
                  disabled={formState.isSubmitting}
                >
                  {formState.isSubmitting ? __("Saving...") : __("Save Changes")}
                </Button>
              </Authorized>
            </div>
          )}
        </form>
      </Card>
    </div>
  );
}
