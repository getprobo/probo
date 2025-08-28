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
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { Controller } from "react-hook-form";
import { formatDatetime } from "@probo/helpers";

const schema = z.object({
  referenceId: z.string().min(1, "Reference ID is required"),
  title: z.string().min(1, "Title is required"),
  description: z.string().optional(),
  source: z.string().optional(),
  category: z.string().optional(),
  ownerId: z.string().min(1, "Owner is required"),
  incidentDate: z.string().optional(),
  status: z.enum(["OPEN", "IN_PROGRESS", "RESOLVED", "CLOSED"]),
  priority: z.enum(["LOW", "MEDIUM", "HIGH", "CRITICAL"]),
  severity: z.enum(["LOW", "MEDIUM", "HIGH", "CRITICAL"]),
});

type FormData = z.infer<typeof schema>;

interface CreateIncidentRegistryDialogProps {
  children: ReactNode;
  organizationId: string;
  connectionId?: string;
}

export function CreateIncidentRegistryDialog({
  children,
  organizationId,
}: CreateIncidentRegistryDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  // Note: We'll use mock implementation for now since we don't have GraphQL mutation yet
  const createRegistry = async (data: any) => {
    // Mock implementation - in real app this would be a GraphQL mutation
    console.log("Creating incident registry:", data);
    return Promise.resolve();
  };

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      referenceId: "",
      title: "",
      description: "",
      source: "",
      category: "",
      ownerId: "",
      incidentDate: "",
      status: "OPEN" as const,
      priority: "MEDIUM" as const,
      severity: "MEDIUM" as const,
    },
  });

  const onSubmit = handleSubmit(async (formData: FormData) => {
    try {
      await createRegistry({
        organizationId,
        referenceId: formData.referenceId,
        title: formData.title,
        description: formData.description || undefined,
        source: formData.source || undefined,
        category: formData.category || undefined,
        ownerId: formData.ownerId,
        incidentDate: formatDatetime(formData.incidentDate),
        status: formData.status,
        priority: formData.priority,
        severity: formData.severity,
      });

      toast({
        title: __("Success"),
        description: __("Incident registry entry created successfully"),
        variant: "success",
      });

      reset();
      dialogRef.current?.close();
    } catch (error) {
      toast({
        title: __("Error"),
        description: __("Failed to create incident registry entry"),
        variant: "error",
      });
    }
  });

  const statusOptions = [
    { value: "OPEN", label: __("Open") },
    { value: "IN_PROGRESS", label: __("In Progress") },
    { value: "RESOLVED", label: __("Resolved") },
    { value: "CLOSED", label: __("Closed") },
  ];

  const priorityOptions = [
    { value: "LOW", label: __("Low") },
    { value: "MEDIUM", label: __("Medium") },
    { value: "HIGH", label: __("High") },
    { value: "CRITICAL", label: __("Critical") },
  ];

  const severityOptions = [
    { value: "LOW", label: __("Low") },
    { value: "MEDIUM", label: __("Medium") },
    { value: "HIGH", label: __("High") },
    { value: "CRITICAL", label: __("Critical") },
  ];

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Registries"), __("Create Incident Entry")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Reference ID")}
            {...register("referenceId")}
            placeholder="INC-001"
            error={formState.errors.referenceId?.message}
            required
          />

          <Field
            label={__("Title")}
            {...register("title")}
            placeholder={__("Brief description of the incident")}
            error={formState.errors.title?.message}
            required
          />

          <div>
            <Label>{__("Description")}</Label>
            <Textarea
              {...register("description")}
              placeholder={__("Detailed description of the incident")}
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
              placeholder={__("How was this incident discovered?")}
              error={formState.errors.source?.message}
            />

            <Field
              label={__("Category")}
              {...register("category")}
              placeholder={__("e.g., Security, Technical, Operational")}
              error={formState.errors.category?.message}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>{__("Incident Date")}</Label>
              <Input
                type="date"
                {...register("incidentDate")}
              />
              {formState.errors.incidentDate?.message && (
                <div className="text-red-500 text-sm mt-1">
                  {formState.errors.incidentDate.message}
                </div>
              )}
            </div>

            <PeopleSelectField
              organizationId={organizationId}
              control={control}
              name="ownerId"
              label={__("Owner")}
              error={formState.errors.ownerId?.message}
              required
            />
          </div>

          <div className="grid grid-cols-3 gap-4">
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

            <Controller
              control={control}
              name="severity"
              render={({ field }) => (
                <div>
                  <Label>{__("Severity")} *</Label>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    {severityOptions.map((option) => (
                      <Option key={option.value} value={option.value}>
                        {option.label}
                      </Option>
                    ))}
                  </Select>
                  {formState.errors.severity?.message && (
                    <div className="text-red-500 text-sm mt-1">
                      {formState.errors.severity.message}
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
            {formState.isSubmitting ? __("Creating...") : __("Create Incident")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}