import { useTranslate } from "@probo/i18n";
import { Button, PageHeader, TabBadge, TabItem, Tabs } from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { PeoplePageQuery } from "#/__generated__/iam/PeoplePageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { InvitationList } from "./_components/InvitationList";
import { InviteUserDialog } from "./_components/InviteUserDialog";
import { PeopleList } from "./_components/PeopleList";

export const peoplePageQuery = graphql`
  query PeoplePageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canInviteUser: permission(action: "iam:invitation:create")
        members(first: 20, orderBy: { direction: ASC, field: FULL_NAME })
          @required(action: THROW) {
          totalCount
        }
        ...PeopleListFragment
          @arguments(first: 20, order: { direction: ASC, field: FULL_NAME })
        invitations(first: 20, orderBy: { direction: DESC, field: CREATED_AT })
          @required(action: THROW) {
          totalCount
          ...PeoplePage_invitationsTotalCountFragment
        }
        ...InvitationListFragment
          @arguments(first: 20, order: { direction: DESC, field: CREATED_AT })
      }
    }
  }
`;

export const invitationCountFragment = graphql`
  fragment PeoplePage_invitationsTotalCountFragment on InvitationConnection
  @updatable {
    totalCount
  }
`;

export function PeoplePage(props: {
  queryRef: PreloadedQuery<PeoplePageQuery>;
}) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const [activeTab, setActiveTab] = useState<"memberships" | "invitations">(
    "memberships",
  );

  const { organization } = usePreloadedQuery<PeoplePageQuery>(
    peoplePageQuery,
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
    <div className="space-y-6">
      <PageHeader title={__("People")}>
        {organization.canInviteUser && (
          <InviteUserDialog
            connectionId={invitationListConnectionId}
            fKey={organization.invitations}
          >
            <Button variant="secondary">{__("Invite member")}</Button>
          </InviteUserDialog>
        )}
      </PageHeader>

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

      <div className="pb-6 pt-6">
        {activeTab === "memberships" && <PeopleList fKey={organization} />}

        {activeTab === "invitations" && (
          <InvitationList
            fKey={organization}
            totalCount={organization.invitations.totalCount ?? 0}
            totalCountFKey={organization.invitations}
            onConnectionIdChange={setInvitationListConnectionId}
          />
        )}
      </div>
    </div>
  );
}
