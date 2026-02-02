import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  IconMagnifyingGlass,
  IconPlusLarge,
  Input,
} from "@probo/ui";
import { useMemo, useState } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { MembershipsPageQuery } from "#/__generated__/iam/MembershipsPageQuery.graphql";

import { InvitationCard } from "./_components/InvitationCard";
import { MembershipCard } from "./_components/MembershipCard";

export const membershipsPageQuery = graphql`
  query MembershipsPageQuery {
    viewer @required(action: THROW) {
      memberships(
        first: 1000
        orderBy: { direction: ASC, field: ORGANIZATION_NAME }
      )
        @connection(key: "MembershipsPage_memberships")
        @required(action: THROW) {
        __id
        edges @required(action: THROW) {
          node {
            id
            ...MembershipCardFragment
            organization @required(action: THROW) {
              name
            }
          }
        }
      }
      pendingInvitations(
        first: 1000
        orderBy: { direction: DESC, field: CREATED_AT }
      )
        @connection(key: "MembershipsPage_pendingInvitations")
        @required(action: THROW) {
        __id
        edges @required(action: THROW) {
          node {
            id
            ...InvitationCardFragment
          }
        }
      }
    }
  }
`;

export function MembershipsPage(props: {
  queryRef: PreloadedQuery<MembershipsPageQuery>;
}) {
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");

  usePageTitle(__("Select an organization"));

  const { queryRef } = props;
  const {
    viewer: {
      memberships: { __id: membershipConnectionId, edges: initialMemberships },
      pendingInvitations: {
        __id: pendingInvitationsConnectionId,
        edges: invitations,
      },
    },
  } = usePreloadedQuery<MembershipsPageQuery>(membershipsPageQuery, queryRef);

  const memberships = useMemo(() => {
    if (!search.trim()) {
      return initialMemberships;
    }
    return initialMemberships.filter(({ node }) =>
      node.organization.name.toLowerCase().includes(search.toLowerCase()),
    );
  }, [initialMemberships, search]);

  return (
    <>
      <div className="space-y-6 w-full py-6">
        <h1 className="text-3xl font-bold text-center">
          {__("Select an organization")}
        </h1>
        <div className="space-y-4 w-full">
          {invitations.length > 0 && (
            <div className="space-y-3">
              <h2 className="text-xl font-semibold">
                {__("Pending invitations")}
              </h2>
              {invitations.map(({ node }) => (
                <InvitationCard
                  pendingInvitationsConnectionId={
                    pendingInvitationsConnectionId
                  }
                  membershipConnectionId={membershipConnectionId}
                  key={node.id}
                  fKey={node}
                />
              ))}
            </div>
          )}
          {initialMemberships.length > 0 && (
            <div className="space-y-3">
              <h2 className="text-xl font-semibold">
                {__("Your organizations")}
              </h2>
              <div className="w-full">
                <Input
                  icon={IconMagnifyingGlass}
                  placeholder={__("Search organizations...")}
                  value={search}
                  onValueChange={setSearch}
                />
              </div>
              {memberships.length === 0
                ? (
                    <div className="text-center text-txt-secondary py-4">
                      {__("No organizations found")}
                    </div>
                  )
                : (
                    memberships.map(({ node }) => (
                      <MembershipCard key={node.id} fKey={node} />
                    ))
                  )}
            </div>
          )}
          <Card padded>
            <h2 className="text-xl font-semibold mb-1">
              {__("Create an organization")}
            </h2>
            <p className="text-txt-tertiary mb-4">
              {__("Add a new organization to your account")}
            </p>
            <Button
              to="/organizations/new"
              variant="quaternary"
              icon={IconPlusLarge}
              className="w-full"
            >
              {__("Create organization")}
            </Button>
          </Card>
        </div>
      </div>
    </>
  );
}
