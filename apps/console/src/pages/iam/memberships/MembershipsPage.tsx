// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

import { InvitingOrganizationCard } from "./_components/InvitingOrganizationCard";
import { MembershipCard } from "./_components/MembershipCard";

export const membershipsPageQuery = graphql`
  query MembershipsPageQuery {
    viewer @required(action: THROW) {
      profiles(
        first: 1000
        orderBy: { direction: ASC, field: ORGANIZATION_NAME }
        filter: { state: ACTIVE }
      )
        @connection(key: "MembershipsPage_profiles")
        @required(action: THROW) {
        edges @required(action: THROW) {
          node {
            id
            ...MembershipCardFragment
            organization @required(action: THROW) {
              name
              ...MembershipCard_organizationFragment
            }
          }
        }
      }
      invitingOrganizations {
        id
        ...InvitingOrganizationCardFragment
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
      profiles: { edges: initialProfiles },
      invitingOrganizations,
    },
  } = usePreloadedQuery<MembershipsPageQuery>(membershipsPageQuery, queryRef);

  const profiles = useMemo(() => {
    if (!search.trim()) {
      return initialProfiles;
    }
    return initialProfiles.filter(({ node }) =>
      node.organization.name.toLowerCase().includes(search.toLowerCase()),
    );
  }, [initialProfiles, search]);

  return (
    <>
      <div className="space-y-6 w-full py-6">
        <h1 className="text-3xl font-bold text-center">
          {__("Select an organization")}
        </h1>
        <div className="space-y-4 w-full">
          {invitingOrganizations.length > 0 && (
            <div className="space-y-3">
              <h2 className="text-xl font-semibold">
                {__("Pending invitations")}
              </h2>
              {invitingOrganizations.map(organization => (
                <InvitingOrganizationCard key={organization.id} fKey={organization} />
              ))}
            </div>
          )}
          {initialProfiles.length > 0 && (
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
              {profiles.length === 0
                ? (
                    <div className="text-center text-txt-secondary py-4">
                      {__("No organizations found")}
                    </div>
                  )
                : (
                    profiles.map(({ node }) => (
                      <MembershipCard
                        key={node.id}
                        fKey={node}
                        organizationFragmentRef={node.organization}
                      />
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
