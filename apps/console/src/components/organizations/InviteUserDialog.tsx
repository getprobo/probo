import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Checkbox,
  useDialogRef,
} from "@probo/ui";
import type { PropsWithChildren } from "react";
import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { Controller } from "react-hook-form";

const inviteMutation = graphql`
  mutation InviteUserDialogMutation(
    $input: InviteUserInput!
    $connections: [ID!]!
  ) {
    inviteUser(input: $input) {
      invitationEdge @appendEdge(connections: $connections) {
        node {
          id
          email
          fullName
          role
          expiresAt
          acceptedAt
          createdAt
        }
      }
    }
  }
`;

const schema = z.object({
  email: z.string().email(),
  fullName: z.string(),
  createPeople: z.boolean().default(false),
});

type Props = PropsWithChildren & {
  connectionId?: string;
};

export function InviteUserDialog({ children, connectionId }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const [inviteUser, isInviting] = useMutationWithToasts(inviteMutation, {
    successMessage: __("User invited successfully"),
    errorMessage: __("Failed to invite user"),
  });
  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(
    schema,
    { defaultValues: { createPeople: false } },
  );

  const dialogRef = useDialogRef();

  const onSubmit = handleSubmit((data) => {
    inviteUser({
      variables: {
        input: {
          organizationId,
          email: data.email,
          fullName: data.fullName,
          createPeople: data.createPeople,
        },
        connections: connectionId ? [connectionId] : ["SettingsPageInvitations_invitations"],
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
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Controller
                name="createPeople"
                control={control}
                render={({ field }) => (
                  <>
                    <Checkbox
                      checked={field.value}
                      onChange={field.onChange}
                    />
                    <label
                      className="text-sm font-medium cursor-pointer"
                      onClick={() => field.onChange(!field.value)}
                    >
                      {__("Create people record")}
                    </label>
                  </>
                )}
              />
            </div>
            <p className="text-xs text-txt-secondary ml-7">
              {__("Creates a people record for this user in addition to the user account")}
            </p>
          </div>
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
