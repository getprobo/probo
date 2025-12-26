import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  IconTrashCan,
  Spinner,
  Td,
  Tr,
  useConfirm,
} from "@probo/ui";
import clsx from "clsx";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { sprintf } from "@probo/helpers";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { InvitationListItemFragment$key } from "./__generated__/InvitationListItemFragment.graphql";
import type { InvitationListItem_permissionsFragment$key } from "./__generated__/InvitationListItem_permissionsFragment.graphql";

const fragment = graphql`
  fragment InvitationListItemFragment on Invitation {
    id
    fullName
    email
    role
    status
    createdAt
    expiresAt
    acceptedAt
  }
`;

const permissionsFragment = graphql`
  fragment InvitationListItem_permissionsFragment on Identity
  @argumentDefinitions(organizationId: { type: "ID!" }) {
    canDeleteInvitation: permission(
      action: "iam:invitation:delete"
      id: $organizationId
    )
  }
`;

const deleteInvitationMutation = graphql`
  mutation InvitationListItem_DeleteInvitationMutation(
    $input: DeleteInvitationInput!
    $connections: [ID!]!
  ) {
    deleteInvitation(input: $input) {
      deletedInvitationId @deleteEdge(connections: $connections)
    }
  }
`;

export function InvitationListItem(props: {
  connectionId: string;
  fKey: InvitationListItemFragment$key;
  permissionsFKey: InvitationListItem_permissionsFragment$key;
  onRefetch: () => void;
}) {
  const { connectionId, fKey, permissionsFKey } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const confirm = useConfirm();

  const invitation = useFragment<InvitationListItemFragment$key>(
    fragment,
    fKey,
  );
  const permissions = useFragment<InvitationListItem_permissionsFragment$key>(
    permissionsFragment,
    permissionsFKey,
  );

  const [deleteInvitation, isDeleting] = useMutationWithToasts(
    deleteInvitationMutation,
    {
      successMessage: __("Invitation deleted successfully"),
      errorMessage: __("Failed to delete invitation"),
    },
  );

  const handleDelete = () => {
    confirm(
      () => {
        return deleteInvitation({
          variables: {
            input: {
              organizationId,
              invitationId: invitation.id,
            },
            connections: [connectionId],
          },
        });
      },
      {
        message: sprintf(
          __("Are you sure you want to delete the invitation for %s?"),
          invitation.fullName,
        ),
      },
    );
  };

  return (
    <Tr className={clsx(isDeleting && "opacity-60 pointer-events-none")}>
      <Td>
        <div className="font-semibold">{invitation.fullName}</div>
      </Td>
      <Td>{invitation.email}</Td>
      <Td>
        <Badge>{invitation.role}</Badge>
      </Td>
      <Td>{new Date(invitation.createdAt).toLocaleDateString()}</Td>
      <Td>
        {invitation.status === "ACCEPTED" ? (
          <Badge variant="success">{__("Accepted")}</Badge>
        ) : invitation.status === "EXPIRED" ? (
          <Badge variant="danger">{__("Expired")}</Badge>
        ) : (
          <Badge variant="warning">{__("Pending")}</Badge>
        )}
      </Td>
      <Td>
        {invitation.acceptedAt
          ? new Date(invitation.acceptedAt).toLocaleDateString()
          : "-"}
      </Td>
      <Td noLink width={80} className="text-end">
        <div
          className="flex gap-2 justify-end"
          onClick={(e) => e.stopPropagation()}
        >
          {isDeleting ? (
            <Spinner size={16} />
          ) : (
            permissions.canDeleteInvitation && (
              <Button
                variant="danger"
                onClick={handleDelete}
                disabled={isDeleting}
                icon={IconTrashCan}
                aria-label={__("Delete invitation")}
              />
            )
          )}
        </div>
      </Td>
    </Tr>
  );
}
