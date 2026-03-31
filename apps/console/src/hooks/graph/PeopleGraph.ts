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

import { useMemo } from "react";
import {
  useLazyLoadQuery,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { PeopleGraphQuery } from "#/__generated__/core/PeopleGraphQuery.graphql";

/* eslint-disable relay/unused-fields */

export const peopleQuery = graphql`
  query PeopleGraphQuery($organizationId: ID!, $filter: ProfileFilter) {
    organization: node(id: $organizationId) {
      ... on Organization {
        profiles(
          first: 1000
          orderBy: { direction: ASC, field: FULL_NAME }
          filter: $filter
        ) {
          edges {
            node {
              id
              fullName
              emailAddress
            }
          }
        }
      }
    }
  }
`;

/**
 * Return a list of people (used for people selectors)
 */
export function usePeople(
  organizationId: string,
  { excludeContractEnded }: { excludeContractEnded?: boolean } = {},
) {
  const data = useLazyLoadQuery<PeopleGraphQuery>(
    peopleQuery,
    {
      organizationId: organizationId,
      filter: excludeContractEnded ? { excludeContractEnded: true } : null,
    },
    { fetchPolicy: "network-only" },
  );
  return useMemo(() => {
    return data.organization?.profiles?.edges.map(edge => edge.node) ?? [];
  }, [data]);
}

export const paginatedPeopleQuery = graphql`
  query PeopleGraphPaginatedQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...PeopleGraphPaginatedFragment
      }
    }
  }
`;

export const paginatedPeopleFragment = graphql`
  fragment PeopleGraphPaginatedFragment on Organization
  @refetchable(queryName: "PeopleListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "ProfileOrder"
      defaultValue: { direction: ASC, field: FULL_NAME }
    }
    filter: { type: "ProfileFilter", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    canCreatePeople: permission(action: "iam:membership-profile:create")
    profiles(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "PeopleGraphPaginatedQuery_profiles") {
      __id
      edges {
        node {
          id
          fullName
          emailAddress
          kind
          position
          additionalEmailAddresses
          contractStartDate
          contractEndDate
          canDelete: permission(action: "iam:membership-profile:delete")
          canUpdate: permission(action: "iam:membership-profile:update")
        }
      }
    }
  }
`;
