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
  { contractEnded }: { contractEnded?: boolean } = {},
) {
  const data = useLazyLoadQuery<PeopleGraphQuery>(
    peopleQuery,
    {
      organizationId: organizationId,
      filter: contractEnded !== undefined ? { contractEnded } : null,
    },
    { fetchPolicy: "network-only" },
  );
  return useMemo(() => {
    return data.organization?.profiles?.edges.map(edge => edge.node) ?? [];
  }, [data]);
}
