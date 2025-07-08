import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
} from "@probo/ui";
import type { PropsWithChildren } from "react";
import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";

const inviteMutation = graphql`
  mutation InviteUserDialogMutation($input: InviteUserInput!) {
    inviteUser(input: $input) {
      success
    }
  }
`;

const schema = z.object({
  email: z.string().email(),
  fullName: z.string(),
});

export function InviteUserDialog({ children }: PropsWithChildren) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const [inviteUser, isInviting] = useMutationWithToasts(inviteMutation, {
    successMessage: __("User invited successfully"),
    errorMessage: __("Failed to invite user"),
  });
  const { register, handleSubmit, formState, reset } = useFormWithSchema(
    schema,
    {},
  );

  const dialogRef = useDialogRef();

  const onSubmit = handleSubmit((data) => {
    inviteUser({
      variables: {
        input: {
          organizationId,
          email: data.email,
          fullName: data.fullName,
        },
      },
      onSuccess: () => {
        reset();
        dialogRef.current?.close();
      },
    });
  });

  return (
    <Dialog
      title={__("Invite member")}
      trigger={children}
      className="max-w-lg"
      ref={dialogRef}
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-4">
          <p className="text-txt-secondary text-sm">
            Send an invitation to join your workspace.
          </p>
          <Field
            type="email"
            label={__("Email")}
            placeholder={__("Email")}
            {...register("email")}
            error={formState.errors.email?.message}
          />
          <Field
            type="text"
            label={__("Full name")}
            placeholder={__("Full name")}
            {...register("fullName")}
            error={formState.errors.fullName?.message}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isInviting}>
            {__("Invite user")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
