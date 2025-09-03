import { type ReactNode } from "react";
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
import { useCreateObligation } from "../../../../hooks/graph/ObligationGraph";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { Controller } from "react-hook-form";
import { formatDatetime, getStatusOptions } from "@probo/helpers";

const schema = z.object({
  referenceId: z.string().min(1, "Reference ID is required"),
  area: z.string().optional(),
  source: z.string().optional(),
  requirement: z.string().optional(),
  actionsToBeImplemented: z.string().optional(),
  regulator: z.string().optional(),
  ownerId: z.string().min(1, "Owner is required"),
  lastReviewDate: z.string().optional(),
  dueDate: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED"]),
});

type FormData = z.infer<typeof schema>;

interface CreateObligationDialogProps {
  children: ReactNode;
  organizationId: string;
  connection?: string;
}

export function CreateObligationDialog({
  children,
  organizationId,
  connection,
}: CreateObligationDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const createObligation = useCreateObligation(connection || "");
  const statusOptions = getStatusOptions(__);

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      referenceId: "",
      area: "",
      source: "",
      requirement: "",
      actionsToBeImplemented: "",
      regulator: "",
      ownerId: "",
      lastReviewDate: "",
      dueDate: "",
      status: "OPEN" as const,
    },
  });

  const onSubmit = handleSubmit(async (formData: FormData) => {
    try {
      await createObligation({
        organizationId,
        referenceId: formData.referenceId,
        area: formData.area || undefined,
        source: formData.source || undefined,
        requirement: formData.requirement || undefined,
        actionsToBeImplemented: formData.actionsToBeImplemented || undefined,
        regulator: formData.regulator || undefined,
        ownerId: formData.ownerId,
        lastReviewDate: formatDatetime(formData.lastReviewDate),
        dueDate: formatDatetime(formData.dueDate),
        status: formData.status,
      });

      toast({
        title: __("Success"),
        description: __("Obligation created successfully"),
        variant: "success",
      });

      reset();
      dialogRef.current?.close();
    } catch {
      toast({
        title: __("Error"),
        description: __("Failed to create obligation"),
        variant: "error",
      });
    }
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Obligations"), __("Create Obligation")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Reference ID")}
            {...register("referenceId")}
            placeholder="CR-001"
            error={formState.errors.referenceId?.message}
            required
          />

          <div className="grid grid-cols-2 gap-4">
            <Field
              label={__("Area")}
              {...register("area")}
              placeholder={__("Enter area")}
              error={formState.errors.area?.message}
            />

            <Field
              label={__("Source")}
              {...register("source")}
              placeholder={__("Enter source")}
              error={formState.errors.source?.message}
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
            <Field
              label={__("Regulator")}
              {...register("regulator")}
              placeholder={__("Enter regulator")}
              error={formState.errors.regulator?.message}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="lastReviewDate">{__("Last Review Date")}</Label>
              <Input
                id="lastReviewDate"
                type="date"
                {...register("lastReviewDate")}
              />
              {formState.errors.lastReviewDate && (
                <p className="text-sm text-red-500">{formState.errors.lastReviewDate.message}</p>
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
            <Label htmlFor="requirement">{__("Requirement")}</Label>
            <Textarea
              id="requirement"
              {...register("requirement")}
              placeholder={__("Enter requirement details...")}
              rows={3}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="actionsToBeImplemented">{__("Actions to be Implemented")}</Label>
            <Textarea
              id="actionsToBeImplemented"
              {...register("actionsToBeImplemented")}
              placeholder={__("Enter actions to be implemented...")}
              rows={3}
            />
          </div>
        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={formState.isSubmitting}>
            {formState.isSubmitting ? __("Creating...") : __("Create Obligation")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
