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

import { MagnifyingGlassIcon, PlusIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { List } from "@probo/ui/src/v2/List/List";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { MembershipsPageQuery } from "#/__generated__/iam/MembershipsPageQuery.graphql";

import { InvitingOrganizationListItem } from "./_components/InvitingOrganizationListItem";
import { OrganizationListItem } from "./_components/OrganizationListItem";

export const membershipsPageQuery = graphql`
  query MembershipsPageQuery @throwOnFieldError {
    signUpEnabled
    viewer @required(action: THROW) {
      profiles(
        first: 1000
        orderBy: { direction: ASC, field: ORGANIZATION_NAME }
        filter: { states: [ACTIVE] }
      )
        @connection(key: "MembershipsPage_profiles")
        @required(action: THROW) {
        edges @required(action: THROW) {
          node {
            id
            membership @required(action: THROW) {
              role
            }
            ...OrganizationListItem_profile
            organization @required(action: THROW) {
              name
            }
          }
        }
      }
      invitingOrganizations {
        id
        ...InvitingOrganizationListItem_organization
      }
    }
  }
`;

interface MembershipsPageProps {
  queryRef: PreloadedQuery<MembershipsPageQuery>;
}

export function MembershipsPage({ queryRef }: MembershipsPageProps) {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");

  usePageTitle(t("membershipsPage.pageTitle"));

  const {
    signUpEnabled,
    viewer: {
      profiles: { edges: initialProfiles },
      invitingOrganizations,
    },
  } = usePreloadedQuery<MembershipsPageQuery>(membershipsPageQuery, queryRef);

  // Mirrors the server rule: with signup disabled, only owners of an existing
  // organization may create another one.
  const canCreateOrganization
    = signUpEnabled
      || initialProfiles.some(({ node }) => node.membership.role === "OWNER");
  const hasNoAccess
    = !signUpEnabled
      && initialProfiles.length === 0
      && invitingOrganizations.length === 0;

  const profiles = useMemo(() => {
    const query = search.trim().toLowerCase();
    if (!query) {
      return initialProfiles;
    }

    return initialProfiles.filter(({ node }) =>
      node.organization.name.toLowerCase().includes(query),
    );
  }, [initialProfiles, search]);

  return (
    <main className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-8 py-8">
      <Heading level={1} size={6} weight="medium" highContrast>
        {t("membershipsPage.title")}
      </Heading>

      {invitingOrganizations.length > 0 && (
        <List>
          {invitingOrganizations.map(organization => (
            <InvitingOrganizationListItem
              key={organization.id}
              organizationKey={organization}
            />
          ))}
        </List>
      )}

      {initialProfiles.length > 0
        ? (
            <>
              <TextField
                size={2}
                icon={<MagnifyingGlassIcon />}
                placeholder={t("membershipsPage.searchPlaceholder")}
                aria-label={t("membershipsPage.searchPlaceholder")}
                value={search}
                onValueChange={setSearch}
              />
              {profiles.length === 0
                ? (
                    <Text size={2} color="faint">
                      {t("membershipsPage.emptySearch")}
                    </Text>
                  )
                : (
                    <List>
                      {profiles.map(({ node }) => (
                        <OrganizationListItem key={node.id} profileKey={node} />
                      ))}
                    </List>
                  )}
            </>
          )
        : hasNoAccess
          ? (
              <Callout color="neutral" variant="soft">
                {t("noOrganizationAccess.description")}
              </Callout>
            )
          : null}

      {canCreateOrganization && (
        <Card padding={2} variant="soft">
          <div className="flex flex-col gap-3">
            <Text size={2} weight="medium" color="neutral" highContrast>
              {t("membershipsPage.createOrganization.title")}
            </Text>
            <Text size={2} color="faint">
              {t("membershipsPage.createOrganization.description")}
            </Text>
            <div>
              <ButtonLink
                to="/organizations/new"
                size={2}
                variant="soft"
                color="neutral"
                iconStart={<PlusIcon />}
              >
                {t("membershipsPage.createOrganization.action")}
              </ButtonLink>
            </div>
          </div>
        </Card>
      )}
    </main>
  );
}
