// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { MagnifyingGlassIcon } from "@phosphor-icons/react";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { List } from "@probo/ui/src/v2/List/List";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { OrganizationsPageQuery } from "#/__generated__/iam/OrganizationsPageQuery.graphql";
import { TopBarUserMenu } from "#/pages/iam/_components/TopBar/TopBarUserMenu";
import { topBar } from "#/pages/iam/_components/TopBar/variants";

import { OrganizationListItem } from "./_components/OrganizationListItem";

export const organizationsPageQuery = graphql`
  query OrganizationsPageQuery @throwOnFieldError {
    viewer @required(action: THROW) {
      ...TopBarUserMenu_identity
      profiles(
        first: 1000
        orderBy: { direction: ASC, field: ORGANIZATION_NAME }
        filter: { states: [ACTIVE] }
      )
        @connection(key: "OrganizationsPage_profiles")
        @required(action: THROW) {
        edges @required(action: THROW) {
          node {
            id
            organization @required(action: THROW) {
              name
            }
            ...OrganizationListItem_profile
          }
        }
      }
    }
  }
`;

interface OrganizationsPageProps {
  queryRef: PreloadedQuery<OrganizationsPageQuery>;
}

export function OrganizationsPage({ queryRef }: OrganizationsPageProps) {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const slots = topBar();
  const tagline = t("topBar.tagline");

  const { viewer } = usePreloadedQuery<OrganizationsPageQuery>(
    organizationsPageQuery,
    queryRef,
  );
  const initialProfiles = viewer.profiles.edges;

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
    <div className="flex min-h-dvh flex-col bg-sand-2">
      <header className={slots.bar()}>
        <div className={slots.inner()}>
          <span className={slots.brand()}>
            <span className={slots.brandText()}>
              <Text
                size={2}
                weight="medium"
                color="neutral"
                highContrast
                className={slots.brandName()}
              >
                {tagline}
              </Text>
            </span>
          </span>
          <TopBarUserMenu identityKey={viewer} />
        </div>
      </header>

      <main className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-8 py-8">
        <Heading level={1} size={6} weight="medium" highContrast>
          {t("organizationsPage.title")}
        </Heading>

        {initialProfiles.length === 0
          ? (
              <Callout color="neutral" variant="soft">
                {t("organizationsPage.empty")}
              </Callout>
            )
          : (
              <>
                <TextField
                  size={2}
                  icon={<MagnifyingGlassIcon />}
                  placeholder={t("organizationsPage.searchPlaceholder")}
                  value={search}
                  onValueChange={setSearch}
                />
                {profiles.length === 0
                  ? (
                      <Text size={2} color="faint">
                        {t("organizationsPage.emptySearch")}
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
            )}
      </main>
    </div>
  );
}
