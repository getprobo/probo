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
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";

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
          ...VendorContactsTabFragment_contact
        }
      }
    }
  }
`;

const phoneRegex = /^\+[0-9]{8,15}$/;

/**
 * Dialog to create a vendor contact
 */
export function CreateContactDialog({
  children,
  connectionId,
  vendorId,
}: Props) {
  const { __ } = useTranslate();

  const schema = z.object({
    name: z.string().optional(),
    email: z.string().email(__("Please enter a valid email address")).optional().or(z.literal("")),
    phone: z.string().regex(phoneRegex, __("Phone number must be in international format (e.g., +1234567890)")).optional().or(z.literal("")),
    role: z.string().optional(),
  });

  const { register, handleSubmit, formState, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        name: "",
        email: "",
        phone: "",
        role: "",
      },
    }
  );
  const [createContact, isLoading] = useMutationWithToasts(
    createContactMutation,
    {
      successMessage: __("Contact created successfully."),
      errorMessage: __("Failed to create contact. Please try again."),
    }
  );

  const onSubmit = handleSubmit((data) => {
    // Filter out empty strings
    const cleanData = Object.fromEntries(
      Object.entries(data).filter(([_, value]) => value !== "")
    );

    createContact({
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
  });

  const dialogRef = useDialogRef();

  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      trigger={children}
      title={
        <Breadcrumb items={[__("Contacts"), __("New Contact")]} />
      }
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name")}
            type="text"
            error={formState.errors.name?.message}
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
