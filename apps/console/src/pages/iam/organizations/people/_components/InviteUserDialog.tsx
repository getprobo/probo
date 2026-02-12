import { getAssignableRoles } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Option,
  Select,
  useDialogRef,
} from "@probo/ui";
import { type PropsWithChildren, use } from "react";
import { Controller } from "react-hook-form";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { InviteUserDialogMutation } from "#/__generated__/iam/InviteUserDialogMutation.graphql";
import type { PeoplePage_invitationsTotalCountFragment$key } from "#/__generated__/iam/PeoplePage_invitationsTotalCountFragment.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

import { invitationCountFragment } from "../PeoplePage";

const inviteMutation = graphql`
  mutation InviteUserDialogMutation(
    $input: InviteMemberInput!
    $connections: [ID!]!
  ) {
    inviteMember(input: $input) {
      invitationEdge @prependEdge(connections: $connections) {
        node {
          id
          email
          fullName
          role
          expiresAt
          acceptedAt
          createdAt
          canDelete: permission(action: "iam:invitation:delete")
        }
      }
    }
  }
`;

// const totalCountFragment = graphql`
//   fragment InviteUserDialog_totalCountFragment on InvitationConnection
//   @updatable {
//     totalCount
//   }
// `;

const schema = z.object({
  email: z.string().email(),
  fullName: z.string(),
  role: z
    .enum(["OWNER", "ADMIN", "VIEWER", "AUDITOR", "EMPLOYEE"])
    .default("VIEWER"),
});

type InviteUserDialogProps = PropsWithChildren<{
  connectionId: string;
  fKey: PeoplePage_invitationsTotalCountFragment$key;
}>;

export function InviteUserDialog(props: InviteUserDialogProps) {
  const { children, connectionId, fKey } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();

  const { role } = use(CurrentUser);
  const [inviteUser, isInviting]
    = useMutationWithToasts<InviteUserDialogMutation>(inviteMutation, {
      successMessage: __("Invitation sent successfully"),
      errorMessage: __("Failed to send invitation"),
    });

  const assignableRoles = getAssignableRoles(role);

  const { register, handleSubmit, formState, reset, control }
    = useFormWithSchema(schema, {
      defaultValues: { role: "VIEWER" },
    });

  const onSubmit = async (data: z.infer<typeof schema>) => {
    await inviteUser({
      variables: {
        input: {
          organizationId,
          email: data.email,
          fullName: data.fullName,
          role: data.role,
        },
        connections: [connectionId],
      },
      onCompleted: () => {
        reset();
        dialogRef.current?.close();
      },
      updater: (store) => {
        const { updatableData }
          = store.readUpdatableFragment<PeoplePage_invitationsTotalCountFragment$key>(
            invitationCountFragment,
            fKey,
          );

        if (typeof updatableData.totalCount === "number") {
          updatableData.totalCount += 1;
        }
      },
    });
  };

  return (
    <Dialog
      title={__("Invite member")}
      trigger={children}
      className="max-w-lg"
      ref={dialogRef}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
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
          <Field label={__("Role")} required>
            <Controller
              name="role"
              control={control}
              render={({ field }) => (
                <>
                  <Select value={field.value} onValueChange={field.onChange}>
                    {assignableRoles.includes("OWNER") && (
                      <Option value="OWNER">{__("Owner")}</Option>
                    )}
                    {assignableRoles.includes("ADMIN") && (
                      <Option value="ADMIN">{__("Admin")}</Option>
                    )}
                    {assignableRoles.includes("VIEWER") && (
                      <Option value="VIEWER">{__("Viewer")}</Option>
                    )}
                    {assignableRoles.includes("AUDITOR") && (
                      <Option value="AUDITOR">{__("Auditor")}</Option>
                    )}
                    {assignableRoles.includes("EMPLOYEE") && (
                      <Option value="EMPLOYEE">{__("Employee")}</Option>
                    )}
                  </Select>
                  <div className="mt-2 text-sm text-txt-tertiary">
                    {field.value === "OWNER" && (
                      <p>{__("Full access to everything")}</p>
                    )}
                    {field.value === "ADMIN" && (
                      <p>
                        {__(
                          "Full access except organization setup and API keys",
                        )}
                      </p>
                    )}
                    {field.value === "VIEWER" && (
                      <p>{__("Read-only access")}</p>
                    )}
                    {field.value === "AUDITOR" && (
                      <p>
                        {__(
                          "Read-only access without settings, tasks and meetings",
                        )}
                      </p>
                    )}
                    {field.value === "EMPLOYEE" && (
                      <p>{__("Access to employee page")}</p>
                    )}
                  </div>
                </>
              )}
            />
          </Field>
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
