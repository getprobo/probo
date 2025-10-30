import { useState } from "react";
import { useOutletContext } from "react-router";
import { usePaginationFragment, graphql } from "react-relay";
import {
  Badge,
  Button,
  Card,
  IconTrashCan,
  Spinner,
  TabBadge,
  TabItem,
  Tabs,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { SortableTable, SortableTh } from "/components/SortableTable";
import { InviteUserDialog } from "/components/organizations/InviteUserDialog";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { sprintf } from "@probo/helpers";
import clsx from "clsx";
import type { NodeOf } from "/types";
import type {
  MembersSettingsTabMembershipsFragment$data,
  MembersSettingsTabMembershipsFragment$key
} from "./__generated__/MembersSettingsTabMembershipsFragment.graphql";
import type {
  MembersSettingsTabInvitationsFragment$data,
  MembersSettingsTabInvitationsFragment$key
} from "./__generated__/MembersSettingsTabInvitationsFragment.graphql";

const paginatedMembershipsFragment = graphql`
  fragment MembersSettingsTabMembershipsFragment on Organization
  @refetchable(queryName: "MembersSettingsTabMembershipsRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: { type: "MembershipOrder", defaultValue: { direction: ASC, field: CREATED_AT } }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    memberships(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "MembersSettingsTabMemberships_memberships") {
      __id
      totalCount
      edges {
        node {
          id
          fullName
          emailAddress
          role
          authMethod
          createdAt
        }
      }
    }
  }
`;

const paginatedInvitationsFragment = graphql`
  fragment MembersSettingsTabInvitationsFragment on Organization
  @refetchable(queryName: "MembersSettingsTabInvitationsRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: { type: "InvitationOrder", defaultValue: { direction: ASC, field: CREATED_AT } }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    invitations(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "MembersSettingsTabInvitations_invitations") {
      __id
      totalCount
      edges {
        node {
          id
          fullName
          email
          role
          status
          createdAt
          expiresAt
          acceptedAt
        }
      }
    }
  }
`;

const removeMemberMutation = graphql`
  mutation MembersSettingsTab_RemoveMemberMutation(
    $input: RemoveMemberInput!
    $connections: [ID!]!
  ) {
    removeMember(input: $input) {
      deletedMemberId @deleteEdge(connections: $connections)
    }
  }
`;

const deleteInvitationMutation = graphql`
  mutation MembersSettingsTab_DeleteInvitationMutation(
    $input: DeleteInvitationInput!
    $connections: [ID!]!
  ) {
    deleteInvitation(input: $input) {
      deletedInvitationId @deleteEdge(connections: $connections)
    }
  }
`;

type OutletContext = {
  organization: MembersSettingsTabMembershipsFragment$key & MembersSettingsTabInvitationsFragment$key & { id: string };
};

export default function MembersSettingsTab() {
  const { __ } = useTranslate();
  const { organization: organizationKey } = useOutletContext<OutletContext>();

  const membershipsPagination = usePaginationFragment(
    paginatedMembershipsFragment,
    organizationKey as MembersSettingsTabMembershipsFragment$key
  );

  const invitationsPagination = usePaginationFragment(
    paginatedInvitationsFragment,
    organizationKey as MembersSettingsTabInvitationsFragment$key
  );

  const refetchMemberships = () => {
    membershipsPagination.refetch({}, { fetchPolicy: 'network-only' });
  };

  const refetchInvitations = () => {
    invitationsPagination.refetch({}, { fetchPolicy: 'network-only' });
  };

  const memberships = membershipsPagination.data.memberships?.edges.map((edge) => edge.node) || [];
  const invitations = invitationsPagination.data.invitations?.edges.map((edge) => edge.node) || [];
  const [activeTab, setActiveTab] = useState<"memberships" | "invitations">("memberships");

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-medium">{__("Workspace members")}</h2>
        <InviteUserDialog
          connectionId={invitationsPagination.data.invitations?.__id}
          onRefetch={refetchInvitations}
        >
          <Button variant="secondary">{__("Invite member")}</Button>
        </InviteUserDialog>
      </div>

      <Tabs>
        <TabItem
          active={activeTab === "memberships"}
          onClick={() => setActiveTab("memberships")}
        >
          {__("Members")}
          {(membershipsPagination.data.memberships?.totalCount || 0) > 0 && (
            <TabBadge>{membershipsPagination.data.memberships?.totalCount}</TabBadge>
          )}
        </TabItem>
        <TabItem
          active={activeTab === "invitations"}
          onClick={() => setActiveTab("invitations")}
        >
          {__("Invitations")}
          {(invitationsPagination.data.invitations?.totalCount || 0) > 0 && (
            <TabBadge>{invitationsPagination.data.invitations?.totalCount}</TabBadge>
          )}
        </TabItem>
      </Tabs>

      <Card>
        <div className="px-6 pb-6 pt-6">
          {activeTab === "memberships" && (
            <SortableTable
              {...membershipsPagination}
              refetch={({ order }: { order: { direction: string; field: string } }) => {
                membershipsPagination.refetch({
                  order: {
                    direction: order.direction as "ASC" | "DESC",
                    field: order.field as "CREATED_AT" | "FULL_NAME" | "EMAIL_ADDRESS" | "ROLE"
                  }
                });
              }}
            >
              <Thead>
                <Tr>
                  <SortableTh field="FULL_NAME">{__("Name")}</SortableTh>
                  <SortableTh field="EMAIL_ADDRESS">{__("Email")}</SortableTh>
                  <SortableTh field="ROLE">{__("Role")}</SortableTh>
                  <SortableTh field="CREATED_AT">{__("Joined")}</SortableTh>
                  <Th></Th>
                </Tr>
              </Thead>
              <Tbody>
                {memberships.length === 0 ? (
                  <Tr>
                    <Td colSpan={5} className="text-center text-txt-secondary">
                      {__("No members")}
                    </Td>
                  </Tr>
                ) : (
                  memberships.map((membership) => (
                    <MembershipRow
                      key={membership.id}
                      membership={membership}
                      connectionId={membershipsPagination.data.memberships?.__id}
                      organizationId={(organizationKey as { id: string }).id}
                      onRefetch={refetchMemberships}
                    />
                  ))
                )}
              </Tbody>
            </SortableTable>
          )}

          {activeTab === "invitations" && (
            <SortableTable
              {...invitationsPagination}
              refetch={({ order }: { order: { direction: string; field: string } }) => {
                invitationsPagination.refetch({
                  order: {
                    direction: order.direction as "ASC" | "DESC",
                    field: order.field as "CREATED_AT" | "EXPIRES_AT" | "FULL_NAME" | "EMAIL" | "ROLE" | "STATUS" | "ACCEPTED_AT"
                  }
                });
              }}
            >
              <Thead>
                <Tr>
                  <SortableTh field="FULL_NAME">{__("Name")}</SortableTh>
                  <SortableTh field="EMAIL">{__("Email")}</SortableTh>
                  <SortableTh field="ROLE">{__("Role")}</SortableTh>
                  <SortableTh field="CREATED_AT">{__("Invited")}</SortableTh>
                  <Th>{__("Status")}</Th>
                  <SortableTh field="ACCEPTED_AT">{__("Accepted at")}</SortableTh>
                  <Th></Th>
                </Tr>
              </Thead>
              <Tbody>
                {invitations.length === 0 ? (
                  <Tr>
                    <Td colSpan={7} className="text-center text-txt-secondary">
                      {__("No invitations")}
                    </Td>
                  </Tr>
                ) : (
                  invitations.map((invitation) => (
                    <InvitationRow
                      key={invitation.id}
                      invitation={invitation}
                      connectionId={invitationsPagination.data.invitations?.__id}
                      organizationId={(organizationKey as { id: string }).id}
                      onRefetch={refetchInvitations}
                    />
                  ))
                )}
              </Tbody>
            </SortableTable>
          )}
        </div>
      </Card>
    </div>
  );
}

function InvitationRow(props: {
  invitation: NodeOf<MembersSettingsTabInvitationsFragment$data["invitations"]>;
  connectionId?: string;
  organizationId: string;
  onRefetch: () => void;
}) {
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const [deleteInvitation, isDeleting] = useMutationWithToasts(
    deleteInvitationMutation,
    {
      successMessage: __("Invitation deleted successfully"),
      errorMessage: __("Failed to delete invitation"),
    }
  );

  const onDelete = () => {
    confirm(
      () => {
        return deleteInvitation({
          variables: {
            input: {
              invitationId: props.invitation.id,
            },
            connections: props.connectionId ? [props.connectionId] : [],
          },
          onCompleted: () => {
            props.onRefetch();
          },
        });
      },
      {
        message: sprintf(
          __("Are you sure you want to delete the invitation for %s?"),
          props.invitation.fullName
        ),
      }
    );
  };

  return (
    <Tr className={clsx(isDeleting && "opacity-60 pointer-events-none")}>
      <Td>
        <div className="font-semibold">{props.invitation.fullName}</div>
      </Td>
      <Td>{props.invitation.email}</Td>
      <Td>
        <Badge>{props.invitation.role}</Badge>
      </Td>
      <Td>{new Date(props.invitation.createdAt).toLocaleDateString()}</Td>
      <Td>
        {props.invitation.status === "ACCEPTED" ? (
          <Badge variant="success">{__("Accepted")}</Badge>
        ) : props.invitation.status === "EXPIRED" ? (
          <Badge variant="danger">{__("Expired")}</Badge>
        ) : (
          <Badge variant="warning">{__("Pending")}</Badge>
        )}
      </Td>
      <Td>
        {props.invitation.acceptedAt ? new Date(props.invitation.acceptedAt).toLocaleDateString() : "-"}
      </Td>
      <Td noLink width={80} className="text-end">
        <div
          className="flex gap-2 justify-end"
          onClick={(e) => e.stopPropagation()}
        >
          {isDeleting ? (
            <Spinner size={16} />
          ) : (
            <Button
              variant="danger"
              onClick={onDelete}
              disabled={isDeleting}
              icon={IconTrashCan}
              aria-label={__("Delete invitation")}
            />
          )}
        </div>
      </Td>
    </Tr>
  );
}

function MembershipRow(props: {
  membership: NodeOf<MembersSettingsTabMembershipsFragment$data["memberships"]>;
  connectionId?: string;
  organizationId: string;
  onRefetch: () => void;
}) {
  const { __ } = useTranslate();
  const [removeMember, isRemoving] = useMutationWithToasts(removeMemberMutation, {
    successMessage: __("Member removed successfully"),
    errorMessage: __("Failed to remove member"),
  });
  const confirm = useConfirm();
  const [isRemoved, setIsRemoved] = useState(false);

  if (isRemoved) {
    return null;
  }

  const onRemove = async () => {
    confirm(
      () => {
        return removeMember({
          variables: {
            input: {
              memberId: props.membership.id,
              organizationId: props.organizationId,
            },
            connections: props.connectionId ? [props.connectionId] : [],
          },
          onCompleted: () => {
            setIsRemoved(true);
            props.onRefetch();
          },
        });
      },
      {
        message: sprintf(
          __("Are you sure you want to remove %s?"),
          props.membership.fullName
        ),
      }
    );
  };

  return (
    <Tr className={clsx(isRemoving && "opacity-60 pointer-events-none")}>
      <Td>
        <div className="font-semibold">{props.membership.fullName}</div>
      </Td>
      <Td>
        <div className="flex items-center gap-2">
          {props.membership.emailAddress}
          {props.membership.authMethod === "SAML" && (
            <Badge variant="info">SAML</Badge>
          )}
        </div>
      </Td>
      <Td>
        <Badge>{props.membership.role}</Badge>
      </Td>
      <Td>{new Date(props.membership.createdAt).toLocaleDateString()}</Td>
      <Td noLink width={80} className="text-end">
        <div
          className="flex gap-2 justify-end"
          onClick={(e) => e.stopPropagation()}
        >
          {isRemoving ? (
            <Spinner size={16} />
          ) : (
            <Button
              variant="danger"
              onClick={onRemove}
              disabled={isRemoving}
              icon={IconTrashCan}
              aria-label={__("Remove member")}
            />
          )}
        </div>
      </Td>
    </Tr>
  );
}
