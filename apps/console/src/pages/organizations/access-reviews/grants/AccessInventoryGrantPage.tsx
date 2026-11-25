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

import { CaretLeftIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useLocation } from "react-router";
import { graphql } from "relay-runtime";

import type { AccessInventoryGrantPageQuery } from "#/__generated__/core/AccessInventoryGrantPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { AccessInventoryGrantChangeEventList } from "./_components/AccessInventoryGrantChangeEventList";
import { AccessInventoryGrantPropertiesSection } from "./_components/AccessInventoryGrantPropertiesSection";
import { accessInventoryGrantPage } from "./_components/variants";

export const accessInventoryGrantPageQuery = graphql`
  query AccessInventoryGrantPageQuery(
    $grantId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
  ) {
    grant: node(id: $grantId) @required(action: THROW) {
      __typename
      ... on AccessInventoryGrant {
        id
        entitlementDisplayName
        entitlementKey
        presence
        organization {
          id
        }
        principal {
          email
          fullName
        }
        ...AccessInventoryGrantPropertiesSection_grant
        ...AccessInventoryGrantChangeEventList_grant
          @arguments(
            first: $first
            after: $after
            last: $last
            before: $before
          )
      }
    }
  }
`;

interface AccessInventoryGrantPageProps {
  queryRef: PreloadedQuery<AccessInventoryGrantPageQuery>;
}

export function AccessInventoryGrantPage({
  queryRef,
}: AccessInventoryGrantPageProps) {
  const { t } = useTranslation("organizations/access-reviews");
  const location = useLocation();
  const organizationId = useOrganizationId();
  const { root, back, header, intro, titleRow, title, subtitle }
    = accessInventoryGrantPage();

  const { grant } = usePreloadedQuery<AccessInventoryGrantPageQuery>(
    accessInventoryGrantPageQuery,
    queryRef,
  );
  if (
    grant.__typename !== "AccessInventoryGrant"
    || grant.organization.id !== organizationId
  ) {
    throw new NotFoundError(t("grantDetail.notFound"));
  }

  const fullName = grant.principal.fullName.trim();
  const email = grant.principal.email.trim();
  const principalTitle = fullName || email || t("grants.values.unknown");
  const entitlement = grant.entitlementDisplayName.trim()
    || grant.entitlementKey
    || t("grants.values.unknown");
  const disappeared = grant.presence === "DISAPPEARED";

  usePageTitle(entitlement);

  return (
    <div className={root()}>
      <Link
        to={{ pathname: "..", search: location.search }}
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("grantDetail.back")}
      </Link>
      <div className={header()}>
        <div className={intro()}>
          <div className={titleRow()}>
            <Heading
              level={1}
              size={6}
              weight="medium"
              highContrast
              className={title()}
            >
              {entitlement}
            </Heading>
            <Badge
              size={1}
              variant="soft"
              color={disappeared ? "amber" : "green"}
            >
              {t(
                disappeared
                  ? "grants.presence.disappeared"
                  : "grants.presence.present",
              )}
            </Badge>
          </div>
          <Text size={2} color="neutral" className={subtitle()}>
            {principalTitle}
          </Text>
        </div>
      </div>
      <AccessInventoryGrantPropertiesSection grantKey={grant} />
      <AccessInventoryGrantChangeEventList grantKey={grant} />
    </div>
  );
}
