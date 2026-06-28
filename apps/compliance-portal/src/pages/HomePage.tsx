// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";

import { Hero } from "#/components/Hero/Hero";
import { OrganizationContactInfo } from "#/components/Hero/OrganizationContactInfo";

import type { HomePageQuery } from "./__generated__/HomePageQuery.graphql";

export const homePageQuery = graphql`
  query HomePageQuery {
    currentTrustCenter @required(action: THROW) {
      organization {
        name
        description
        ...OrganizationContactInfo_organization
      }
    }
  }
`;

interface HomePageProps {
  queryRef: PreloadedQuery<HomePageQuery>;
}

export function HomePage({ queryRef }: HomePageProps) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<HomePageQuery>(homePageQuery, queryRef);
  const { organization } = data.currentTrustCenter;

  return (
    <Hero
      title={t("home.heroTitle", { name: organization.name })}
      description={organization.description}
    >
      <OrganizationContactInfo organizationKey={organization} />
    </Hero>
  );
}
