import { formatError, type GraphQLError } from "@probo/helpers";
import { formatDatetime, getObligationStatusOptions, getObligationTypeOptions } from "@probo/helpers";
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

import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import { useCreateObligation } from "../../../../hooks/graph/ObligationGraph";

const schema = z.object({
  area: z.string().optional(),
  source: z.string().optional(),
  requirement: z.string().optional(),
  actionsToBeImplemented: z.string().optional(),
  regulator: z.string().optional(),
  type: z.enum(["LEGAL", "CONTRACTUAL"]),
  ownerId: z.string().min(1, "Owner is required"),
  lastReviewDate: z.string().optional(),
  dueDate: z.string().optional(),
  status: z.enum(["NON_COMPLIANT", "PARTIALLY_COMPLIANT", "COMPLIANT"]),
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
  const statusOptions = getObligationStatusOptions(__);
  const typeOptions = getObligationTypeOptions(__);

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      area: "",
      source: "",
      requirement: "",
      actionsToBeImplemented: "",
      regulator: "",
      type: "LEGAL" as const,
      ownerId: "",
      lastReviewDate: "",
      dueDate: "",
      status: "NON_COMPLIANT" as const,
    },
  });

  const onSubmit = async (formData: FormData) => {
    try {
      await createObligation({
        organizationId,
        area: formData.area || undefined,
        source: formData.source || undefined,
        requirement: formData.requirement || undefined,
        actionsToBeImplemented: formData.actionsToBeImplemented || undefined,
        regulator: formData.regulator || undefined,
        type: formData.type,
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
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to create obligation"), error as GraphQLError),
        variant: "error",
      });
    }
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Obligations"), __("Create Obligation")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
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
            <Field
              label={__("Regulator")}
              {...register("regulator")}
              placeholder={__("Enter regulator")}
              error={formState.errors.regulator?.message}
            />

            <Field label={__("Type")}>
              <Controller
                control={control}
                name="type"
                render={({ field }) => (
                  <Select
                    variant="editor"
                    placeholder={__("Select type")}
                    onValueChange={field.onChange}
                    value={field.value}
                    className="w-full"
                  >
                    {typeOptions.map(option => (
                      <Option key={option.value} value={option.value}>
                        {option.label}
                      </Option>
                    ))}
                  </Select>
                )}
              />
              {formState.errors.type && (
                <p className="text-sm text-red-500 mt-1">{formState.errors.type.message}</p>
              )}
            </Field>
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
