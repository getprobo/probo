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
import {
  Button,
  Card,
  IconMagnifyingGlass,
  IconPlusLarge,
  Input,
} from "@probo/ui";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
  const [search, setSearch] = useState("");

  usePageTitle(t("membershipsPage.pageTitle"));

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
          {t("membershipsPage.title")}
        </h1>
        <div className="space-y-4 w-full">
          {invitingOrganizations.length > 0 && (
            <div className="space-y-3">
              <h2 className="text-xl font-semibold">
                {t("membershipsPage.pendingInvitations")}
              </h2>
              {invitingOrganizations.map(organization => (
                <InvitingOrganizationCard key={organization.id} fKey={organization} />
              ))}
            </div>
          )}
          {initialProfiles.length > 0 && (
            <div className="space-y-3">
              <h2 className="text-xl font-semibold">
                {t("membershipsPage.yourOrganizations")}
              </h2>
              <div className="w-full">
                <Input
                  icon={IconMagnifyingGlass}
                  placeholder={t("membershipsPage.searchPlaceholder")}
                  value={search}
                  onValueChange={setSearch}
                />
              </div>
              {profiles.length === 0
                ? (
                    <div className="text-center text-txt-secondary py-4">
                      {t("membershipsPage.empty")}
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
              {t("membershipsPage.createOrganization.title")}
            </h2>
            <p className="text-txt-tertiary mb-4">
              {t("membershipsPage.createOrganization.description")}
            </p>
            <Button
              to="/organizations/new"
              variant="quaternary"
              icon={IconPlusLarge}
              className="w-full"
            >
              {t("membershipsPage.createOrganization.action")}
            </Button>
          </Card>
        </div>
      </div>
    </>
  );
}
