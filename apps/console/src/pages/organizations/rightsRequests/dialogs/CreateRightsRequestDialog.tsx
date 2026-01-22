import {
  formatDatetime,
  formatError,
  getRightsRequestStateOptions,
  getRightsRequestTypeOptions,
  type GraphQLError,
} from "@probo/helpers";
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

import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import { useCreateRightsRequest } from "../../../../hooks/graph/RightsRequestGraph";

const schema = z.object({
  requestType: z.enum(["ACCESS", "DELETION", "PORTABILITY"]),
  requestState: z.enum(["TODO", "IN_PROGRESS", "DONE"]),
  dataSubject: z.string().optional(),
  contact: z.string().optional(),
  details: z.string().optional(),
  deadline: z.string().optional(),
  actionTaken: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

interface CreateRightsRequestDialogProps {
  children: ReactNode;
  organizationId: string;
  connectionId?: string;
}

export function CreateRightsRequestDialog({
  children,
  organizationId,
  connectionId,
}: CreateRightsRequestDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const createRequest = useCreateRightsRequest(connectionId || "");

  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(schema, {
    defaultValues: {
      requestType: "ACCESS" as const,
      requestState: "TODO" as const,
      dataSubject: "",
      contact: "",
      details: "",
      deadline: "",
      actionTaken: "",
    },
  });

  const onSubmit = async (formData: FormData) => {
    try {
      await createRequest({
        organizationId,
        requestType: formData.requestType,
        requestState: formData.requestState,
        dataSubject: formData.dataSubject || undefined,
        contact: formData.contact || undefined,
        details: formData.details || undefined,
        deadline: formatDatetime(formData.deadline),
        actionTaken: formData.actionTaken || undefined,
      });

      toast({
        title: __("Success"),
        description: __("Rights request created successfully"),
        variant: "success",
      });

      reset();
      dialogRef.current?.close();
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to create rights request"), error as GraphQLError),
        variant: "error",
      });
    }
  };

  const typeOptions = getRightsRequestTypeOptions(__);
  const stateOptions = getRightsRequestStateOptions(__);

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Rights Requests"), __("Create Request")]} />}
      className="max-w-2xl"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Controller
              control={control}
              name="requestType"
              render={({ field }) => (
                <div>
                  <Label>
                    {__("Request Type")}
                    {" "}
                    *
                  </Label>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    {typeOptions.map(option => (
                      <Option key={option.value} value={option.value}>
                        {option.label}
                      </Option>
                    ))}
                  </Select>
                  {formState.errors.requestType?.message && (
                    <div className="text-red-500 text-sm mt-1">
                      {formState.errors.requestType.message}
                    </div>
                  )}
                </div>
              )}
            />

            <Controller
              control={control}
              name="requestState"
              render={({ field }) => (
                <div>
                  <Label>
                    {__("State")}
                    {" "}
                    *
                  </Label>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    {stateOptions.map(option => (
                      <Option key={option.value} value={option.value}>
                        {option.label}
                      </Option>
                    ))}
                  </Select>
                  {formState.errors.requestState?.message && (
                    <div className="text-red-500 text-sm mt-1">
                      {formState.errors.requestState.message}
                    </div>
                  )}
                </div>
              )}
            />
          </div>

          <Field
            label={__("Data Subject")}
            {...register("dataSubject")}
            placeholder={__("Enter data subject name")}
            error={formState.errors.dataSubject?.message}
          />

          <Field
            label={__("Contact")}
            {...register("contact")}
            placeholder={__("Enter contact information")}
            error={formState.errors.contact?.message}
          />

          <div>
            <Label>{__("Details")}</Label>
            <Textarea
              {...register("details")}
              placeholder={__("Enter request details")}
              rows={3}
            />
            {formState.errors.details?.message && (
              <div className="text-red-500 text-sm mt-1">
                {formState.errors.details.message}
              </div>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>{__("Deadline")}</Label>
              <Input
                type="date"
                {...register("deadline")}
              />
              {formState.errors.deadline?.message && (
                <div className="text-red-500 text-sm mt-1">
                  {formState.errors.deadline.message}
                </div>
              )}
            </div>
          </div>

          <div>
            <Label>{__("Action Taken")}</Label>
            <Textarea
              {...register("actionTaken")}
              placeholder={__("Enter action taken")}
              rows={3}
            />
            {formState.errors.actionTaken?.message && (
              <div className="text-red-500 text-sm mt-1">
                {formState.errors.actionTaken.message}
              </div>
            )}
          </div>
        </DialogContent>

        <DialogFooter>
          <Button
            type="submit"
            variant="primary"
            disabled={formState.isSubmitting}
          >
            {formState.isSubmitting ? __("Creating...") : __("Create Request")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
