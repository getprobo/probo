import { getAssignableRoles, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Option,
  Select,
  Spinner,
} from "@probo/ui";
import { use, useState } from "react";
import { graphql } from "relay-runtime";

import type { PeopleListItemFragment$data } from "#/__generated__/iam/PeopleListItemFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

const updateMembershipMutation = graphql`
  mutation EditMemberDialog_updateMutation($input: UpdateMembershipInput!) {
    updateMembership(input: $input) {
      membership {
        id
        role
      }
    }
  }
`;

export function EditMemberDialog(props: {
  membership: PeopleListItemFragment$data;
  onClose: () => void;
}) {
  const { membership, onClose } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const { role } = use(CurrentUser);
  const availableRoles = getAssignableRoles(role);

  const [selectedRole, setSelectedRole] = useState<string>(membership.role);

  const [updateMembership, isUpdating] = useMutationWithToasts(
    updateMembershipMutation,
    {
      successMessage: __("Role updated successfully"),
      errorMessage: __("Failed to update role"),
    },
  );

  const handleUpdateRole = async (e: React.FormEvent) => {
    e.preventDefault();
    await updateMembership({
      variables: {
        input: {
          membershipId: membership.id,
          organizationId: organizationId,
          role: selectedRole,
        },
      },
      onCompleted: onClose,
    });
  };

  return (
    <Dialog defaultOpen={true} onClose={onClose} title={__("Edit Profile")}>
      <form onSubmit={e => void handleUpdateRole(e)}>
        <DialogContent padded className="space-y-6">
          <div>
            <p className="text-txt-secondary text-sm mb-4">
              {sprintf(
                __("Update the role for %s"),
                membership.profile.fullName,
              )}
            </p>

            <Field label={__("Role")} required>
              <Select value={selectedRole} onValueChange={setSelectedRole}>
                {availableRoles.includes("OWNER") && (
                  <Option value="OWNER">{__("Owner")}</Option>
                )}
                {availableRoles.includes("ADMIN") && (
                  <Option value="ADMIN">{__("Admin")}</Option>
                )}
                {availableRoles.includes("VIEWER") && (
                  <Option value="VIEWER">{__("Viewer")}</Option>
                )}
                {availableRoles.includes("AUDITOR") && (
                  <Option value="AUDITOR">{__("Auditor")}</Option>
                )}
                {availableRoles.includes("EMPLOYEE") && (
                  <Option value="EMPLOYEE">{__("Employee")}</Option>
                )}
              </Select>
            </Field>

            <div className="mt-4 space-y-2 text-sm text-txt-tertiary">
              {selectedRole === "OWNER" && (
                <p>{__("Full access to everything")}</p>
              )}
              {selectedRole === "ADMIN" && (
                <p>
                  {__("Full access except organization setup and API keys")}
                </p>
              )}
              {selectedRole === "VIEWER" && <p>{__("Read-only access")}</p>}
              {selectedRole === "AUDITOR" && (
                <p>
                  {__("Read-only access without settings, tasks and meetings")}
                </p>
              )}
              {selectedRole === "EMPLOYEE" && (
                <p>{__("Access to employee page")}</p>
              )}
            </div>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button
            type="submit"
            disabled={isUpdating || selectedRole === membership.role}
          >
            {isUpdating && <Spinner />}
            {__("Update Role")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
