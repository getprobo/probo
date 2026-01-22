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
import { type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

type Props = {
  children: ReactNode;
  connectionId: string;
  vendorId: string;
};

const createServiceMutation = graphql`
  mutation CreateServiceDialogMutation(
    $input: CreateVendorServiceInput!
    $connections: [ID!]!
  ) {
    createVendorService(input: $input) {
      vendorServiceEdge @prependEdge(connections: $connections) {
        node {
          ...VendorServicesTabFragment_service
        }
      }
    }
  }
`;

export function CreateServiceDialog({
  children,
  connectionId,
  vendorId,
}: Props) {
  const { __ } = useTranslate();

  const schema = z.object({
    name: z.string().min(1, __("Service name is required")),
    description: z.string().optional(),
  });

  const { register, handleSubmit, formState, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        name: "",
        description: "",
      },
    },
  );
  const [createService, isLoading] = useMutationWithToasts(
    createServiceMutation,
    {
      successMessage: __("Service created successfully."),
      errorMessage: __("Failed to create service"),
    },
  );

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const cleanData = cleanFormData(data);

    await createService({
      variables: {
        input: {
          vendorId,
          ...cleanData,
        },
        connections: [connectionId],
      },
      onSuccess: () => {
        dialogRef.current?.close();
        reset();
      },
    });
  };

  const dialogRef = useDialogRef();

  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      trigger={children}
      title={
        <Breadcrumb items={[__("Services"), __("New Service")]} />
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
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
