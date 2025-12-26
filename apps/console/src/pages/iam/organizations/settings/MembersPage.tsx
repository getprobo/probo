import { usePreloadedQuery, type PreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";
import { MemberList } from "./_components/MemberList";
import type { MembersPageQuery } from "./__generated__/MembersPageQuery.graphql";
import { useTranslate } from "@probo/i18n";
import { Card, TabBadge, TabItem, Tabs } from "@probo/ui";
import { useState } from "react";

export const membersPageQuery = graphql`
  query MembersPageQuery($organizationId: ID!) {
    viewer @required(action: THROW) {
      ...MemberListItem_permissionsFragment
        @arguments(organizationId: $organizationId)
    }
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...MemberListFragment
          @arguments(first: 20, order: { direction: ASC, field: CREATED_AT })
        members(first: 20, orderBy: { direction: ASC, field: CREATED_AT })
          @required(action: THROW) {
          totalCount
        }
      }
    }
  }
`;

export function MembersPage(props: {
  queryRef: PreloadedQuery<MembersPageQuery>;
}) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const [activeTab, setActiveTab] = useState<"memberships" | "invitations">(
    "memberships",
  );

  const { organization, viewer } = usePreloadedQuery<MembersPageQuery>(
    membersPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("node is of invalid type");
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-medium">{__("Workspace members")}</h2>
        {/* {isAuthorized("Organization", "inviteUser") && (
          <InviteUserDialog
            connectionId={invitationsPagination.data.invitations?.__id}
            onRefetch={refetchInvitations}
          >
            <Button variant="secondary">{__("Invite member")}</Button>
          </InviteUserDialog>
        )} */}
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
        {/* <TabItem
          active={activeTab === "invitations"}
          onClick={() => setActiveTab("invitations")}
        >
          {__("Invitations")}
          {(invitationsPagination.data.invitations?.totalCount || 0) > 0 && (
            <TabBadge>
              {invitationsPagination.data.invitations?.totalCount}
            </TabBadge>
          )}
        </TabItem> */}
      </Tabs>

      <Card>
        <div className="px-6 pb-6 pt-6">
          {activeTab === "memberships" && (
            <MemberList fKey={organization} viewerFKey={viewer} />
          )}

          {/* {activeTab === "invitations" && (
            <SortableTable
              {...invitationsPagination}
              refetch={({
                order,
              }: {
                order: { direction: string; field: string };
              }) => {
                invitationsPagination.refetch({
                  order: {
                    direction: order.direction as "ASC" | "DESC",
                    field: order.field as
                      | "CREATED_AT"
                      | "EXPIRES_AT"
                      | "FULL_NAME"
                      | "EMAIL"
                      | "ROLE"
                      | "STATUS"
                      | "ACCEPTED_AT",
                  },
                });
              }}
              pageSize={20}
            >
              <Thead>
                <Tr>
                  <SortableTh field="FULL_NAME">{__("Name")}</SortableTh>
                  <SortableTh field="EMAIL">{__("Email")}</SortableTh>
                  <SortableTh field="ROLE">{__("Role")}</SortableTh>
                  <SortableTh field="CREATED_AT">{__("Invited")}</SortableTh>
                  <Th>{__("Status")}</Th>
                  <SortableTh field="ACCEPTED_AT">
                    {__("Accepted at")}
                  </SortableTh>
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
                      connectionId={
                        invitationsPagination.data.invitations?.__id
                      }
                      organizationId={(organizationKey as { id: string }).id}
                      onRefetch={refetchInvitations}
                    />
                  ))
                )}
              </Tbody>
            </SortableTable>
          )} */}
        </div>
      </Card>
    </div>
  );
}
