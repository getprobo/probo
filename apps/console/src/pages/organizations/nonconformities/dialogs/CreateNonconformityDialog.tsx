import { formatError, type GraphQLError } from "@probo/helpers";
import { formatDatetime, getStatusOptions } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Label,
  Option,
  Select,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode } from "react";
import { Controller } from "react-hook-form";
import { z } from "zod";

import { AuditSelectField } from "#/components/form/AuditSelectField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import { useCreateNonconformity } from "../../../../hooks/graph/NonconformityGraph";

const schema = z.object({
  referenceId: z.string().min(1, "Reference ID is required"),
  description: z.string().optional(),
  auditId: z.string().optional(),
  dateIdentified: z.string().optional(),
  rootCause: z.string().min(1, "Root cause is required"),
  correctiveAction: z.string().optional(),
  ownerId: z.string().min(1, "Owner is required"),
  dueDate: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED"]),
  effectivenessCheck: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

interface CreateNonconformityDialogProps {
  children: ReactNode;
  connection?: string;
  organizationId: string;
}

export function CreateNonconformityDialog({
  children,
  organizationId,
  connection,
}: CreateNonconformityDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const createNonconformity = useCreateNonconformity(connection || "");
  const statusOptions = getStatusOptions(__);

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      referenceId: "",
      description: "",
      auditId: "",
      dateIdentified: "",
      rootCause: "",
      correctiveAction: "",
      ownerId: "",
      dueDate: "",
      status: "OPEN" as const,
      effectivenessCheck: "",
    },
  });

  const onSubmit = async (formData: FormData) => {
    try {
      await createNonconformity({
        organizationId,
        referenceId: formData.referenceId,
        description: formData.description || undefined,
        auditId: formData.auditId || undefined,
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
        description: __("Nonconformity created successfully"),
        variant: "success",
      });

      reset();
      dialogRef.current?.close();
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to create nonconformity"), error as GraphQLError),
        variant: "error",
      });
    }
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Nonconformities"), __("Create")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Reference ID")}
            {...register("referenceId")}
            placeholder="NC-001"
            error={formState.errors.referenceId?.message}
            required
          />

          <AuditSelectField
            organizationId={organizationId}
            control={control}
            name="auditId"
            label={__("Audit")}
            error={formState.errors.auditId?.message}
          />

          <div className="space-y-2">
            <Label htmlFor="description">{__("Description")}</Label>
            <Textarea
              id="description"
              {...register("description")}
              placeholder={__("Brief description of the nonconformity...")}
              rows={2}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
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
                    {statusOptions.map(option => (
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

            <PeopleSelectField
              organizationId={organizationId}
              control={control}
              name="ownerId"
              label={__("Owner")}
              error={formState.errors.ownerId?.message}
              required
            />
          </div>

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

          <div className="space-y-2">
            <Label htmlFor="rootCause">
              {__("Root Cause")}
              {" "}
              *
            </Label>
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
            {formState.isSubmitting ? __("Creating...") : __("Create Nonconformity")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
