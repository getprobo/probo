import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Checkbox,
  Select,
  Option,
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
import { Suspense } from "react";
import { getAssignableRoles } from "/permissions";

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
  role: z.enum(["OWNER", "ADMIN", "EMPLOYEE", "FULL", "VIEWER"]).default("VIEWER"),
  createPeople: z.boolean().default(false),
});

type Props = PropsWithChildren & {
  connectionId?: string;
  onRefetch: () => void;
};

function InviteUserDialogContent({ children, connectionId, onRefetch }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const assignableRoles = getAssignableRoles(organizationId);
  const [inviteUser, isInviting] = useMutationWithToasts(inviteMutation, {
    successMessage: __("Invitation sent successfully"),
    errorMessage: __("Failed to send invitation"),
  });
  const { register, handleSubmit, formState, reset, control } = useFormWithSchema(
    schema,
    { defaultValues: { role: "VIEWER", createPeople: false } },
  );

  const dialogRef = useDialogRef();

  const onSubmit = handleSubmit((data) => {
    inviteUser({
      variables: {
        input: {
          organizationId,
          email: data.email,
          fullName: data.fullName,
          role: data.role,
          createPeople: data.createPeople,
        },
        connections: connectionId ? [connectionId] : ["SettingsPageInvitations_invitations"],
      },
      onCompleted: () => {
        reset();
        dialogRef.current?.close();
        onRefetch();
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
                    {assignableRoles.includes("OWNER") && <Option value="OWNER">{__("Owner")}</Option>}
                    {assignableRoles.includes("ADMIN") && <Option value="ADMIN">{__("Admin")}</Option>}
                    {assignableRoles.includes("VIEWER") && <Option value="VIEWER">{__("Viewer")}</Option>}
                    {assignableRoles.includes("EMPLOYEE") && <Option value="EMPLOYEE">{__("Employee")}</Option>}
                  </Select>
                  <div className="mt-2 text-sm text-txt-tertiary">
                    {field.value === "OWNER" && (
                      <p>{__("Full access to everything")}</p>
                    )}
                    {field.value === "ADMIN" && (
                      <p>{__("Full access except organization setup and API keys")}</p>
                    )}
                    {field.value === "VIEWER" && (
                      <p>{__("Read-only access")}</p>
                    )}
                    {field.value === "EMPLOYEE" && (
                      <p>{__("Access only to documents with requested signature")}</p>
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

export function InviteUserDialog(props: Props) {
  return (
    <Suspense fallback={props.children}>
      <InviteUserDialogContent {...props} />
    </Suspense>
  );
}
