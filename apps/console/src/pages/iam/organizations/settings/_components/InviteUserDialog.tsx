import { getAssignableRoles } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Dialog,
  DialogContent,
  Field,
  Select,
  Option,
  DialogFooter,
  Button,
  Checkbox,
  useDialogRef,
} from "@probo/ui";
import { Controller } from "react-hook-form";
import { useFragment } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";
import z from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { PropsWithChildren } from "react";
import type { InviteUserDialog_currentRoleFragment$key } from "/__generated__/iam/InviteUserDialog_currentRoleFragment.graphql";

const currentRoleFragment = graphql`
  fragment InviteUserDialog_currentRoleFragment on Organization {
    viewerMembership @required(action: THROW) {
      role
    }
  }
`;

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
        }
      }
    }
  }
`;

const schema = z.object({
  email: z.string().email(),
  fullName: z.string(),
  role: z
    .enum(["OWNER", "ADMIN", "FULL", "VIEWER", "AUDITOR", "EMPLOYEE"])
    .default("VIEWER"),
  createPeople: z.boolean().default(false),
});

type InviteUserDialogProps = PropsWithChildren<{
  viewerMembershipFKey: InviteUserDialog_currentRoleFragment$key;
}>;

export function InviteUserDialog(props: InviteUserDialogProps) {
  const { children, viewerMembershipFKey } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();

  const { viewerMembership } =
    useFragment<InviteUserDialog_currentRoleFragment$key>(
      currentRoleFragment,
      viewerMembershipFKey,
    );
  const [inviteUser, isInviting] = useMutationWithToasts(inviteMutation, {
    successMessage: __("Invitation sent successfully"),
    errorMessage: __("Failed to send invitation"),
  });

  const assignableRoles = getAssignableRoles(viewerMembership.role);

  const { register, handleSubmit, formState, reset, control } =
    useFormWithSchema(schema, {
      defaultValues: { role: "VIEWER", createPeople: false },
    });

  const onSubmit = handleSubmit((data) => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      "InvitationListFragment_invitations",
    );
    inviteUser({
      variables: {
        input: {
          organizationId,
          email: data.email,
          fullName: data.fullName,
          role: data.role,
          createPeople: data.createPeople,
        },
        connections: [connectionId],
      },
      onCompleted: () => {
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
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <Controller
                name="createPeople"
                control={control}
                render={({ field }) => (
                  <>
                    <Checkbox
                      checked={field.value ?? false}
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
              {__(
                "Creates a people record for this user in addition to the user account",
              )}
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
