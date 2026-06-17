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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  IconChevronDown,
  PageHeader,
  Spinner,
  Table,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link } from "react-router";

import type { OAuthTokensPageFragment$key } from "#/__generated__/iam/OAuthTokensPageFragment.graphql";
import type { OAuthTokensPageQuery } from "#/__generated__/iam/OAuthTokensPageQuery.graphql";
import type { OAuthTokensPageRefetchQuery } from "#/__generated__/iam/OAuthTokensPageRefetchQuery.graphql";

import { OAuthTokenRow } from "./_components/OAuthTokenRow";

export const oauthTokensPageQuery = graphql`
  query OAuthTokensPageQuery {
    viewer @required(action: THROW) {
      id
      canCreateOAuth2AccessToken: permission(
        action: "iam:oauth2-access-token:create"
      )
      ...OAuthTokensPageFragment
    }
  }
`;

const oauthTokensPageFragment = graphql`
  fragment OAuthTokensPageFragment on Identity
  @refetchable(queryName: "OAuthTokensPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey" }
  ) {
    oauth2AccessTokens(first: $first, after: $after)
      @connection(key: "OAuthTokensPage_oauth2AccessTokens")
      @required(action: THROW) {
      edges {
        node {
          id
          ...OAuthTokenRowFragment
        }
      }
      totalCount @required(action: THROW)
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

export function OAuthTokensPage(props: {
  queryRef: PreloadedQuery<OAuthTokensPageQuery>;
}) {
  const { queryRef } = props;
  const { __ } = useTranslate();

  usePageTitle(__("OAuth tokens"));

  const { viewer } = usePreloadedQuery(oauthTokensPageQuery, queryRef);
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment<
    OAuthTokensPageRefetchQuery,
    OAuthTokensPageFragment$key
  >(oauthTokensPageFragment, viewer);

  const tokens = data.oauth2AccessTokens.edges;
  const totalCount = data.oauth2AccessTokens.totalCount;

  return (
    <div className="space-y-6 w-full py-6">
      <PageHeader
        title={__("OAuth tokens")}
        description={__(
          "Create bearer tokens with scoped API access for your account.",
        )}
      >
        {viewer.canCreateOAuth2AccessToken && (
          <Button asChild>
            <Link to="/me/oauth-tokens/new">
              {__("Create token")}
            </Link>
          </Button>
        )}
      </PageHeader>

      {tokens.length === 0
        ? (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-medium text-gray-900 mb-2">
                  {__("No OAuth tokens")}
                </h3>
                <p className="text-gray-600">
                  {__(
                    "Create a token to authenticate API requests on your behalf.",
                  )}
                </p>
              </div>
            </Card>
          )
        : (
            <Card padded className="space-y-4">
              {totalCount > tokens.length && (
                <p className="text-sm text-txt-tertiary">
                  {`${__("Showing")} ${tokens.length} ${__("of")} ${totalCount}`}
                </p>
              )}
              <Table>
                <Thead>
                  <Tr>
                    <Th>{__("Name")}</Th>
                    <Th>{__("Scopes")}</Th>
                    <Th>{__("Created")}</Th>
                    <Th>{__("Expires")}</Th>
                    <Th className="w-0" />
                  </Tr>
                </Thead>
                <Tbody>
                  {tokens.map(edge => (
                    <OAuthTokenRow
                      key={edge.node.id}
                      tokenKey={edge.node}
                      identityId={viewer.id}
                    />
                  ))}
                </Tbody>
              </Table>
              {hasNext && (
                <Button
                  variant="tertiary"
                  onClick={() => loadNext(50)}
                  className="mx-auto"
                  disabled={isLoadingNext}
                  icon={isLoadingNext ? Spinner : IconChevronDown}
                >
                  {__("Show more")}
                </Button>
              )}
            </Card>
          )}
    </div>
  );
}
