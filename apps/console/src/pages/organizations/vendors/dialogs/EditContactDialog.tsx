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

import type { VendorContactsTabFragment_contact$data } from "#/__generated__/core/VendorContactsTabFragment_contact.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

type Props = {
  contactId: string;
  contact: VendorContactsTabFragment_contact$data;
  onClose: () => void;
};

const updateContactMutation = graphql`
  mutation EditContactDialogUpdateMutation($input: UpdateVendorContactInput!) {
    updateVendorContact(input: $input) {
      vendorContact {
        ...VendorContactsTabFragment_contact
      }
    }
  }
`;

const phoneRegex = /^\+[0-9]{8,15}$/;

export function EditContactDialog({ contactId, contact, onClose }: Props) {
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

  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      fullName: contact.fullName || "",
      email: contact.email || "",
      phone: contact.phone || "",
      role: contact.role || "",
    },
  });

  const [updateContact, isLoading] = useMutationWithToasts(
    updateContactMutation,
    {
      successMessage: __("Contact updated successfully."),
      errorMessage: __("Failed to update contact"),
    },
  );

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const cleanData = cleanFormData(data);

    await updateContact({
      variables: {
        input: {
          id: contactId,
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
      title={<Breadcrumb items={[__("Contacts"), __("Edit Contact")]} />}
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
            {__("Save")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
