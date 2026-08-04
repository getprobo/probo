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

import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { usePaginatedAuditProgramsFragment$key } from "#/__generated__/core/usePaginatedAuditProgramsFragment.graphql";
import type { usePaginatedAuditProgramsQuery } from "#/__generated__/core/usePaginatedAuditProgramsQuery.graphql";
import type { usePaginatedAuditProgramsQuery_fragment } from "#/__generated__/core/usePaginatedAuditProgramsQuery_fragment.graphql";

/* eslint-disable relay/unused-fields */

const auditProgramsQuery = graphql`
  query usePaginatedAuditProgramsQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...usePaginatedAuditProgramsFragment
      }
    }
  }
`;

const auditProgramsFragment = graphql`
  fragment usePaginatedAuditProgramsFragment on Organization
  @refetchable(queryName: "usePaginatedAuditProgramsQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "AuditProgramOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    auditPrograms(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "usePaginatedAuditProgramsQuery_auditPrograms") {
      edges {
        node {
          id
          name
          framework {
            id
          }
        }
      }
    }
  }
`;

/**
 * Hook to retrieve audit programs paginated (used for audit program selectors)
 */
export function usePaginatedAuditPrograms(organizationId: string) {
  const query = useLazyLoadQuery<usePaginatedAuditProgramsQuery>(
    auditProgramsQuery,
    {
      organizationId,
    },
    { fetchPolicy: "network-only" },
  );
  return usePaginationFragment<
    usePaginatedAuditProgramsQuery_fragment,
    usePaginatedAuditProgramsFragment$key
  >(
    auditProgramsFragment,
    query.organization as usePaginatedAuditProgramsFragment$key,
  );
}
