import { cleanFormData } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
} from "@probo/ui";
import { useEffect } from "react";
import { graphql } from "relay-runtime";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

type Props = {
  serviceId: string;
  service: {
    name: string;
    description?: string | null;
  };
  onClose: () => void;
};

const updateServiceMutation = graphql`
  mutation EditServiceDialogUpdateMutation($input: UpdateVendorServiceInput!) {
    updateVendorService(input: $input) {
      vendorService {
        ...VendorServicesTabFragment_service
      }
    }
  }
`;

export function EditServiceDialog({ serviceId, service, onClose }: Props) {
  const { __ } = useTranslate();

  const schema = z.object({
    name: z.string().min(1, __("Service name is required")),
    description: z.string().optional(),
  });

  const { register, handleSubmit, formState } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        name: service.name || "",
        description: service.description || "",
      },
    },
  );

  const [updateService, isLoading] = useMutationWithToasts(
    updateServiceMutation,
    {
      successMessage: __("Service updated successfully."),
      errorMessage: __("Failed to update service"),
    },
  );

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const cleanData = cleanFormData(data);

    await updateService({
      variables: {
        input: {
          id: serviceId,
          ...cleanData,
        },
      },
      onSuccess: () => {
        onClose();
      },
    });
  };

  const dialogRef = useDialogRef();

  useEffect(() => {
    dialogRef.current?.open();
  }, [dialogRef]);

  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      onClose={onClose}
      title={
        <Breadcrumb items={[__("Services"), __("Edit Service")]} />
      }
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name")}
            type="text"
            error={formState.errors.name?.message}
            placeholder={__("Service name")}
            required
          />
          <Field
            label={__("Description")}
            {...register("description")}
            type="textarea"
            error={formState.errors.description?.message}
            placeholder={__("Brief description of the service")}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isLoading}>
            {__("Save")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
