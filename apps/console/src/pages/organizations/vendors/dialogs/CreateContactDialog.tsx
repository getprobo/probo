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

const createContactMutation = graphql`
  mutation CreateContactDialogMutation(
    $input: CreateVendorContactInput!
    $connections: [ID!]!
  ) {
    createVendorContact(input: $input) {
      vendorContactEdge @prependEdge(connections: $connections) {
        node {
          canUpdate: permission(action: "core:vendor-contact:update")
          canDelete: permission(action: "core:vendor-contact:delete")
          ...VendorContactsTabFragment_contact
        }
      }
    }
  }
`;

const phoneRegex = /^\+[0-9]{8,15}$/;

export function CreateContactDialog({
  children,
  connectionId,
  vendorId,
}: Props) {
  const { __ } = useTranslate();

  const schema = z.object({
    fullName: z.string().optional(),
    email: z.union([
      z.string().email(__("Please enter a valid email address")),
      z.literal(""),
    ]),
    phone: z.union([
      z
        .string()
        .regex(
          phoneRegex,
          __(
            "Phone number must be in international format (e.g., +1234567890)",
          ),
        ),
      z.literal(""),
    ]),
    role: z.string().optional(),
  });

  const { register, handleSubmit, formState, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        fullName: "",
        email: "",
        phone: "",
        role: "",
      },
    },
  );
  const [createContact, isLoading] = useMutationWithToasts(
    createContactMutation,
    {
      successMessage: __("Contact created successfully."),
      errorMessage: __("Failed to create contact"),
    },
  );

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const cleanData = cleanFormData(data);

    await createContact({
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
      title={<Breadcrumb items={[__("Contacts"), __("New Contact")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("fullName")}
            type="text"
            error={formState.errors.fullName?.message}
            placeholder={__("Contact's full name")}
          />
          <Field
            label={__("Email")}
            {...register("email")}
            type="email"
            error={formState.errors.email?.message}
            placeholder={__("contact@example.com")}
          />
          <Field
            label={__("Phone")}
            {...register("phone")}
            type="text"
            error={formState.errors.phone?.message}
            placeholder={__("e.g., +1234567890")}
            help={__("Use international format starting with +")}
          />
          <Field
            label={__("Role")}
            {...register("role")}
            type="text"
            error={formState.errors.role?.message}
            placeholder={__("e.g., Account Manager, Technical Support")}
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
