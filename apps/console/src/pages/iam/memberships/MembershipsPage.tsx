// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
