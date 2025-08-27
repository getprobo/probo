import {
  ConnectionHandler,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import {
  nonconformityRegistryNodeQuery,
  useDeleteNonconformityRegistry,
  useUpdateNonconformityRegistry,
  RegistriesConnectionKey,
} from "../../../hooks/graph/NonconformityRegistryGraph";
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
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { ControlledField } from "/components/form/ControlledField";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { AuditSelectField } from "/components/form/AuditSelectField";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import z from "zod";
import { getStatusVariant, getStatusLabel, formatDatetime, getNonconformityRegistryStatusOptions } from "@probo/helpers";
import type { NonconformityRegistryGraphNodeQuery } from "/hooks/graph/__generated__/NonconformityRegistryGraphNodeQuery.graphql";

const updateRegistrySchema = z.object({
  referenceId: z.string().min(1, "Reference ID is required"),
  description: z.string().optional(),
  dateIdentified: z.string().optional(),
  dueDate: z.string().optional(),
  rootCause: z.string().min(1, "Root cause is required"),
  correctiveAction: z.string().optional(),
  effectivenessCheck: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED"]),
  ownerId: z.string().min(1, "Owner is required"),
  auditId: z.string().min(1, "Audit is required"),
});

type Props = {
  queryRef: PreloadedQuery<NonconformityRegistryGraphNodeQuery>;
};

export default function NonconformityRegistryDetailsPage(props: Props) {
  const data = usePreloadedQuery<NonconformityRegistryGraphNodeQuery>(
    nonconformityRegistryNodeQuery,
    props.queryRef
  );
  const registry = data.node;
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const deleteRegistry = useDeleteNonconformityRegistry(
    { id: registry.id!, referenceId: registry.referenceId! },
    ConnectionHandler.getConnectionID(organizationId, RegistriesConnectionKey)
  );

  const { control, formState, handleSubmit, register, reset } = useFormWithSchema(
    updateRegistrySchema,
    {
      defaultValues: {
        referenceId: registry.referenceId || "",
        description: registry.description || "",
        dateIdentified: registry.dateIdentified?.split('T')[0] || "",
        dueDate: registry.dueDate?.split('T')[0] || "",
        rootCause: registry.rootCause || "",
        correctiveAction: registry.correctiveAction || "",
        effectivenessCheck: registry.effectivenessCheck || "",
        status: registry.status || "OPEN",
        ownerId: registry.owner?.id || "",
        auditId: registry.audit?.id || "",
      },
    }
  );

  const updateRegistry = useUpdateNonconformityRegistry();
  const { toast } = useToast();

  const onSubmit = handleSubmit(async (formData) => {
    if (!registry.id) return;

    try {
      await updateRegistry({
        id: registry.id,
        referenceId: formData.referenceId,
        description: formData.description,
        dateIdentified: formatDatetime(formData.dateIdentified),
        dueDate: formatDatetime(formData.dueDate),
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
        description: __("Non conformity registry entry updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: error instanceof Error ? error.message : __("Failed to update non conformity registry entry"),
        variant: "error",
      });
    }
  });

  const statusOptions = getNonconformityRegistryStatusOptions(__);

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("Nonconformity Registries"),
            to: `/organizations/${organizationId}/nonconformity-registries`,
          },
          {
            label: registry.referenceId || __("Unknown Nonconformity Registry"),
          },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="text-2xl font-semibold">
            {registry.referenceId}
          </div>
          <Badge variant={getStatusVariant(registry.status || "OPEN")}>
            {getStatusLabel(registry.status || "OPEN")}
          </Badge>
        </div>
        <ActionDropdown variant="secondary">
          <DropdownItem
            variant="danger"
            icon={IconTrashCan}
            onClick={deleteRegistry}
          >
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
      </div>

      <div className="max-w-4xl">
        <Card padded>
          <form onSubmit={onSubmit} className="space-y-6">
            <Field
              label={__("Reference ID")}
              required
              error={formState.errors.referenceId?.message}
            >
              <Input
                {...register("referenceId")}
                placeholder={__("Enter reference ID")}
              />
            </Field>

            <AuditSelectField
              organizationId={organizationId}
              control={control}
              name="auditId"
              label={__("Audit")}
              error={formState.errors.auditId?.message}
              required
            />

            <Field label={__("Description")}>
              <Textarea
                {...register("description")}
                placeholder={__("Enter description")}
                rows={3}
              />
            </Field>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <ControlledField
                control={control}
                name="status"
                type="select"
                label={__("Status")}
                required
              >
                {statusOptions.map((option) => (
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
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field label={__("Date Identified")}>
                <Input {...register("dateIdentified")} type="date" />
              </Field>

              <Field label={__("Due Date")}>
                <Input {...register("dueDate")} type="date" />
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
              />
            </Field>

            <Field label={__("Corrective Action")}>
              <Textarea
                {...register("correctiveAction")}
                placeholder={__("Enter corrective action")}
                rows={3}
              />
            </Field>

            <Field label={__("Effectiveness Check")}>
              <Textarea
                {...register("effectivenessCheck")}
                placeholder={__("Enter effectiveness check details")}
                rows={3}
              />
            </Field>

            <div className="flex justify-end">
              {formState.isDirty && (
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
