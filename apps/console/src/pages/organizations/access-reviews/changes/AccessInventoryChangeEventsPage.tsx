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

import { usePageTitle } from "@probo/hooks";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessInventoryChangeEventsPageQuery } from "#/__generated__/core/AccessInventoryChangeEventsPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { AccessInventoryChangeEventsList } from "./_components/AccessInventoryChangeEventsList";
import { accessInventoryChangeEventsPage } from "./_components/variants";

export const accessInventoryChangeEventsPageQuery = graphql`
  query AccessInventoryChangeEventsPageQuery(
    $organizationId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
    $filter: AccessInventoryChangeEventFilter
  ) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...AccessInventoryChangeEventsList_organization
          @arguments(
            first: $first
            after: $after
            last: $last
            before: $before
            filter: $filter
          )
      }
    }
  }
`;

interface AccessInventoryChangeEventsPageProps {
  queryRef: PreloadedQuery<AccessInventoryChangeEventsPageQuery>;
}

export function AccessInventoryChangeEventsPage({
  queryRef,
}: AccessInventoryChangeEventsPageProps) {
  const { t } = useTranslation("organizations/access-reviews");
  const { root, header, intro } = accessInventoryChangeEventsPage();
  usePageTitle(t("changes.title"));

  const { organization } = usePreloadedQuery<AccessInventoryChangeEventsPageQuery>(
    accessInventoryChangeEventsPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new NotFoundError("invalid type for organization node");
  }

  return (
    <div className={root()}>
      <div className={header()}>
        <div className={intro()}>
          <Heading level={1} size={6} weight="medium" highContrast>
            {t("changes.title")}
          </Heading>
          <Text size={2} color="neutral">
            {t("changes.description")}
          </Text>
        </div>
      </div>
      <AccessInventoryChangeEventsList organizationKey={organization} />
    </div>
  );
}
