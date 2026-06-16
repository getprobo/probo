// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { cleanFormData, formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useEffect } from "react";
import { graphql, useFragment, useMutation } from "react-relay";
import { z } from "zod";

import type { EditContactDialog_contact$key } from "#/__generated__/core/EditContactDialog_contact.graphql";
import type { EditContactDialogUpdateMutation } from "#/__generated__/core/EditContactDialogUpdateMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

type Props = {
  contactKey: EditContactDialog_contact$key;
  onClose: () => void;
};

const editContactDialogFragment = graphql`
  fragment EditContactDialog_contact on ThirdPartyContact {
    id
    fullName
    email
    phone
    role
  }
`;

const updateContactMutation = graphql`
  mutation EditContactDialogUpdateMutation($input: UpdateThirdPartyContactInput!) {
    updateThirdPartyContact(input: $input) {
      thirdPartyContact {
        ...ThirdPartyContactRow_contact
        ...EditContactDialog_contact
      }
    }
  }
`;

const phoneRegex = /^\+[0-9]{8,15}$/;

export function EditContactDialog({ contactKey, onClose }: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const contact = useFragment(editContactDialogFragment, contactKey);

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

  const [updateContact, isUpdating]
    = useMutation<EditContactDialogUpdateMutation>(updateContactMutation);

  const onSubmit = (data: z.infer<typeof schema>) => {
    const cleanData = cleanFormData(data);

    updateContact({
      variables: {
        input: {
          id: contact.id,
          ...cleanData,
        },
      },
      onCompleted(_response, errors) {
        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to update contact"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Contact updated successfully."),
          variant: "success",
        });
        onClose();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to update contact"), error),
          variant: "error",
        });
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
          <Button type="submit" disabled={isUpdating}>
            {__("Save")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
