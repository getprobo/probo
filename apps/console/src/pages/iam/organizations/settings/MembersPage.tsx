import { useTranslate } from "@probo/i18n";
import { Button, Card, TabBadge, TabItem, Tabs } from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { MembersPageQuery } from "/__generated__/iam/MembersPageQuery.graphql";
import { useOrganizationId } from "/hooks/useOrganizationId";

import { InvitationList } from "./_components/InvitationList";
import { InviteUserDialog } from "./_components/InviteUserDialog";
import { MemberList } from "./_components/MemberList";

export const membersPageQuery = graphql`
  query MembersPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canInviteUser: permission(action: "iam:invitation:create")
        ...MemberListFragment
          @arguments(first: 20, order: { direction: ASC, field: FULL_NAME })
        members(first: 20, orderBy: { direction: ASC, field: FULL_NAME })
          @required(action: THROW) {
          totalCount
        }
        ...InvitationListFragment
          @arguments(first: 20, order: { direction: DESC, field: CREATED_AT })
        invitations(first: 20, orderBy: { direction: DESC, field: CREATED_AT })
          @required(action: THROW) {
          totalCount
          ...MembersPage_invitationsTotalCountFragment
        }
      }
    }
  }
`;

export const invitationCountFragment = graphql`
  fragment MembersPage_invitationsTotalCountFragment on InvitationConnection
  @updatable {
    totalCount
  }
`;

export function MembersPage(props: {
  queryRef: PreloadedQuery<MembersPageQuery>;
}) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const [activeTab, setActiveTab] = useState<"memberships" | "invitations">(
    "memberships",
  );

  const { organization } = usePreloadedQuery<MembersPageQuery>(
    membersPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("node is of invalid type");
  }

  const [invitationListConnectionId, setInvitationListConnectionId] = useState(
    ConnectionHandler.getConnectionID(
      organizationId,
      "InvitationListFragment_invitations",
      { orderBy: { direction: "ASC", field: "CREATED_AT" } },
    ),
  );

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-medium">{__("Workspace members")}</h2>
        {organization.canInviteUser && (
          <InviteUserDialog
            connectionId={invitationListConnectionId}
            fKey={organization.invitations}
          >
            <Button variant="secondary">{__("Invite member")}</Button>
          </InviteUserDialog>
        )}
      </div>

      <Tabs>
        <TabItem
          active={activeTab === "memberships"}
          onClick={() => setActiveTab("memberships")}
        >
          {__("Members")}
          {(organization.members.totalCount ?? 0) > 0 && (
            <TabBadge>{organization.members.totalCount}</TabBadge>
          )}
        </TabItem>
        <TabItem
          active={activeTab === "invitations"}
          onClick={() => setActiveTab("invitations")}
        >
          {__("Invitations")}
          {(organization.invitations.totalCount ?? 0) > 0 && (
            <TabBadge>{organization.invitations.totalCount}</TabBadge>
          )}
        </TabItem>
      </Tabs>

      <Card>
        <div className="px-6 pb-6 pt-6">
          {activeTab === "memberships" && <MemberList fKey={organization} />}

          {activeTab === "invitations" && (
            <InvitationList
              fKey={organization}
              totalCount={organization.invitations.totalCount ?? 0}
              totalCountFKey={organization.invitations}
              onConnectionIdChange={setInvitationListConnectionId}
            />
          )}
        </div>
      </Card>
    </div>
  );
}
