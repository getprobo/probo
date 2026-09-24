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

import { PlusIcon } from "@phosphor-icons/react";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { UsersList_organization$key } from "#/__generated__/iam/UsersList_organization.graphql";
import type { UsersListRefetchQuery } from "#/__generated__/iam/UsersListRefetchQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import { useUsersListFilters } from "../_lib/useUsersListFilters";
import { usersList, usersPage } from "../variants";

import { UserListItem } from "./UserListItem";
import { UsersListFilters } from "./UsersListFilters";

export const USER_PAGE_SIZE = 15;

export const usersListFragment = graphql`
  fragment UsersList_organization on Organization
  @refetchable(queryName: "UsersListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 15 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    filter: { type: "ProfileFilter", defaultValue: null }
  ) {
    canCreateUser: permission(action: "iam:membership-profile:create", attributes: { target_role: "VIEWER" })
    profiles(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { direction: DESC, field: CREATED_AT }
      filter: $filter
    ) @required(action: THROW) {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          ...UserListItem_profile
        }
      }
    }
  }
`;

interface UsersListProps {
  organizationKey: UsersList_organization$key;
}

export function UsersList({ organizationKey }: UsersListProps) {
  const { t } = useTranslation();
  const { graphqlFilter } = useUsersListFilters();
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [organization, refetch] = useRefetchableFragment<
    UsersListRefetchQuery,
    UsersList_organization$key
  >(usersListFragment, organizationKey);

  const pageVariablesRef = useRef<CursorPaginationVariables>({
    first: USER_PAGE_SIZE,
    after: null,
    last: null,
    before: null,
  });

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    pageVariablesRef.current = variables;
    refetch({ ...variables, filter: graphqlFilter }, { fetchPolicy: "store-or-network" });
  }, [graphqlFilter, refetch]);

  const { isPending: isPagePending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    organization.profiles.pageInfo,
    USER_PAGE_SIZE,
  );

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    pageVariablesRef.current = {
      first: USER_PAGE_SIZE,
      after: null,
      last: null,
      before: null,
    };
    startRefetchTransition(() => {
      refetch(
        { ...pageVariablesRef.current, filter: graphqlFilter },
        { fetchPolicy: "store-or-network" },
      );
    });
  }, [graphqlFilter, refetch]);

  const edges = organization.profiles.edges;
  const pageInfo = organization.profiles.pageInfo;
  const isPending = isRefetchPending || isPagePending;
  const { header, intro } = usersPage();
  const { root, results, grid, empty, pager } = usersList({ pending: isPending });

  function refetchCurrentPage() {
    startRefetchTransition(() => {
      refetch(
        { ...pageVariablesRef.current, filter: graphqlFilter },
        { fetchPolicy: "network-only" },
      );
    });
  }

  function handleDeleted() {
    if (edges.length === 1 && pageInfo.hasPreviousPage) {
      goPrevious();
      return;
    }
    refetchCurrentPage();
  }

  return (
    <>
      <div className={header()}>
        <div className={intro()}>
          <Heading level={1} size={6} weight="medium" highContrast>
            {t("nav.users")}
          </Heading>
        </div>
        {organization.canCreateUser && (
          <ButtonLink to="new" variant="solid" iconStart={<PlusIcon />}>
            {t("usersPage.actions.add")}
          </ButtonLink>
        )}
      </div>
      <div className={root()}>
        <UsersListFilters />
        {edges.length === 0
          ? (
              <Card variant="soft" size={2}>
                <div className={empty()}>
                  <Text size={2} color="faint">
                    {t("usersPage.empty")}
                  </Text>
                </div>
              </Card>
            )
          : (
              <>
                <div aria-busy={isPending} className={results()}>
                  <div className={grid()}>
                    {edges.map(({ node }) => (
                      <UserListItem
                        key={node.id}
                        profileKey={node}
                        onDeactivated={refetchCurrentPage}
                        onDeleted={handleDeleted}
                      />
                    ))}
                  </div>
                </div>
                <div className={pager()}>
                  <Pagination
                    hasPrevious={pageInfo.hasPreviousPage}
                    hasNext={pageInfo.hasNextPage}
                    previousLabel={t("usersPage.actions.previous")}
                    nextLabel={t("usersPage.actions.next")}
                    showLabels
                    variant="soft"
                    disabled={isPending}
                    onPrevious={goPrevious}
                    onNext={goNext}
                  />
                </div>
              </>
            )}
      </div>
    </>
  );
}
