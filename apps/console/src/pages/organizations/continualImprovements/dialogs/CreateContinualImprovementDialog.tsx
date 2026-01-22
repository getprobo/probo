import { formatError, type GraphQLError } from "@probo/helpers";
import { formatDatetime } from "@probo/helpers";
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

import { useCreateContinualImprovement } from "../../../../hooks/graph/ContinualImprovementGraph";

const schema = z.object({
  referenceId: z.string().min(1, "Reference ID is required"),
  description: z.string().optional(),
  source: z.string().optional(),
  ownerId: z.string().min(1, "Owner is required"),
  targetDate: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "CLOSED"]),
  priority: z.enum(["LOW", "MEDIUM", "HIGH"]),
});

type FormData = z.infer<typeof schema>;

interface CreateContinualImprovementDialogProps {
  children: ReactNode;
  organizationId: string;
  connectionId?: string;
}

export function CreateContinualImprovementDialog({
  children,
  organizationId,
  connectionId,
}: CreateContinualImprovementDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const createImprovement = useCreateContinualImprovement(connectionId || "");

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      referenceId: "",
      description: "",
      source: "",
      ownerId: "",
      targetDate: "",
      status: "OPEN" as const,
      priority: "MEDIUM" as const,
    },
  });

  const onSubmit = async (formData: FormData) => {
    try {
      await createImprovement({
        organizationId,
        referenceId: formData.referenceId,
        description: formData.description || undefined,
        source: formData.source || undefined,
        ownerId: formData.ownerId,
        targetDate: formatDatetime(formData.targetDate),
        status: formData.status,
        priority: formData.priority,
      });

      toast({
        title: __("Success"),
        description: __("Continual improvement entry created successfully"),
        variant: "success",
      });

      reset();
      dialogRef.current?.close();
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to create continual improvement"), error as GraphQLError),
        variant: "error",
      });
    }
  };

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
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Continual Improvements"), __("Create Entry")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Reference ID")}
            {...register("referenceId")}
            placeholder="CI-001"
            error={formState.errors.referenceId?.message}
            required
          />

          <div>
            <Label>{__("Description")}</Label>
            <Textarea
              {...register("description")}
              placeholder={__("Enter description of the continual improvement item")}
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
              placeholder={__("Enter source")}
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
                  <Label>
                    {__("Status")}
                    {" "}
                    *
                  </Label>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
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
        </DialogContent>

        <DialogFooter>
          <Button
            type="submit"
            variant="primary"
            disabled={formState.isSubmitting}
          >
            {formState.isSubmitting ? __("Creating...") : __("Create Entry")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
