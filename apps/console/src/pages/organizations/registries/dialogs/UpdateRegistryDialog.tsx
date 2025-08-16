
import {
  Button,
  Field,
  useToast,
  Dialog,
  DialogContent,
  DialogFooter,
  useDialogRef,
  Textarea,
  Breadcrumb,
  Label,
  Select,
  Option,
  Input,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useUpdateNonconformityRegistry } from "../../../../hooks/graph/NonconformityRegistryGraph";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { Controller } from "react-hook-form";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { formatDatetime } from "@probo/helpers";

const schema = z.object({
  referenceId: z.string().min(1, "Reference ID is required"),
  description: z.string().optional(),
  dateIdentified: z.string().optional(),
  rootCause: z.string().min(1, "Root cause is required"),
  correctiveAction: z.string().optional(),
  ownerId: z.string().min(1, "Owner is required"),
  dueDate: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED"]),
  effectivenessCheck: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

interface UpdateRegistryDialogProps {
  registry: any; // TODO: Type properly when GraphQL types are generated
  onClose: () => void;
}

export function UpdateRegistryDialog({
  registry,
  onClose,
}: UpdateRegistryDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const updateRegistry = useUpdateNonconformityRegistry();
  const organizationId = useOrganizationId();

  // Helper function to format date for input field
  const formatDateForInput = (dateString: string | null | undefined) => {
    if (!dateString) return "";
    return dateString.split('T')[0]; // Extract YYYY-MM-DD from datetime
  };

  const { register, handleSubmit, formState, control } = useFormWithSchema(schema, {
    defaultValues: {
      referenceId: registry?.referenceId || "",
      description: registry?.description || "",
      dateIdentified: formatDateForInput(registry?.dateIdentified),
      rootCause: registry?.rootCause || "",
      correctiveAction: registry?.correctiveAction || "",
      ownerId: registry?.owner?.id || "",
      dueDate: formatDateForInput(registry?.dueDate),
      status: registry?.status || "OPEN",
      effectivenessCheck: registry?.effectivenessCheck || "",
    },
  });

  const onSubmit = handleSubmit(async (formData: FormData) => {
    try {
      await updateRegistry({
        id: registry.id,
        referenceId: formData.referenceId,
        description: formData.description || undefined,
        dateIdentified: formatDatetime(formData.dateIdentified),
        rootCause: formData.rootCause,
        correctiveAction: formData.correctiveAction || undefined,
        ownerId: formData.ownerId,
        dueDate: formatDatetime(formData.dueDate),
        status: formData.status,
        effectivenessCheck: formData.effectivenessCheck || undefined,
      });

      toast({
        title: __("Success"),
        description: __("Registry entry updated successfully"),
        variant: "success",
      });

      onClose();
    } catch (error) {
      toast({
        title: __("Error"),
        description: __("Failed to update registry entry"),
        variant: "error",
      });
    }
  });

  return (
    <Dialog
      ref={dialogRef}
      defaultOpen
      title={<Breadcrumb items={[__("Registries"), __("Edit Entry")]} />}
      className="max-w-2xl"
      onClose={onClose}
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-4">
          <div className="text-sm text-gray-600 mb-4">
            {__("ID")}: {registry?.referenceId}
          </div>

          <Field
            label={__("Reference ID")}
            {...register("referenceId")}
            placeholder="NC-001"
            error={formState.errors.referenceId?.message}
            required
          />

          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="ownerId"
            label={__("Owner")}
            error={formState.errors.ownerId?.message}
            required
          />

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="dateIdentified">{__("Date Identified")}</Label>
              <Input
                id="dateIdentified"
                type="date"
                {...register("dateIdentified")}
              />
              {formState.errors.dateIdentified && (
                <p className="text-sm text-red-500">{formState.errors.dateIdentified.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="dueDate">{__("Due Date")}</Label>
              <Input
                id="dueDate"
                type="date"
                {...register("dueDate")}
              />
              {formState.errors.dueDate && (
                <p className="text-sm text-red-500">{formState.errors.dueDate.message}</p>
              )}
            </div>
          </div>

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
                >
                  <Option value="OPEN">{__("Open")}</Option>
                  <Option value="IN_PROGRESS">{__("In Progress")}</Option>
                  <Option value="CLOSED">{__("Closed")}</Option>
                </Select>
              )}
            />
            {formState.errors.status && (
              <p className="text-sm text-red-500 mt-1">{formState.errors.status.message}</p>
            )}
          </Field>

          <div className="space-y-2">
            <Label htmlFor="description">{__("Description")}</Label>
            <Textarea
              id="description"
              {...register("description")}
              placeholder={__("Brief description of the nonconformity...")}
              rows={2}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="rootCause">{__("Root Cause")} *</Label>
            <Textarea
              id="rootCause"
              {...register("rootCause")}
              placeholder={__("Detailed analysis of the root cause...")}
              rows={3}
            />
            {formState.errors.rootCause && (
              <p className="text-sm text-red-500">{formState.errors.rootCause.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="correctiveAction">{__("Corrective Action")}</Label>
            <Textarea
              id="correctiveAction"
              {...register("correctiveAction")}
              placeholder={__("Proposed corrective actions...")}
              rows={3}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="effectivenessCheck">{__("Effectiveness Check")}</Label>
            <Textarea
              id="effectivenessCheck"
              {...register("effectivenessCheck")}
              placeholder={__("Assessment of corrective action effectiveness...")}
              rows={2}
            />
          </div>
        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={formState.isSubmitting}>
            {formState.isSubmitting ? __("Updating...") : __("Update Registry Entry")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
