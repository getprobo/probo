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

import { graphql, useFragment, useLazyLoadQuery } from "react-relay";

import type { useDefaultEmployeePortalId_organization$key } from "#/__generated__/core/useDefaultEmployeePortalId_organization.graphql";
import type { useDefaultEmployeePortalIdQuery } from "#/__generated__/core/useDefaultEmployeePortalIdQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

// Keep the oldest-portal pick in sync with ViewerMembershipMenu (IAM schema).
// Listing portals is a second request: the field is non-null, so selecting it
// before the permission check fails the query for roles that cannot list.
const defaultEmployeePortalFragment = graphql`
  fragment useDefaultEmployeePortalId_organization on Organization {
    employeePortals(
      first: 1
      orderBy: { field: CREATED_AT, direction: ASC }
    ) {
      edges {
        node {
          id
        }
      }
    }
  }
`;

const defaultEmployeePortalQuery = graphql`
  query useDefaultEmployeePortalIdQuery(
    $organizationId: ID!
    $includePortals: Boolean!
  ) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        canListEmployeePortals: permission(
          action: "employee-portal:portal:list"
        )
        ...useDefaultEmployeePortalId_organization @include(if: $includePortals)
      }
    }
  }
`;

export function useDefaultEmployeePortalId(): string | null {
  const organizationId = useOrganizationId();
  const access = useLazyLoadQuery<useDefaultEmployeePortalIdQuery>(
    defaultEmployeePortalQuery,
    { organizationId, includePortals: false },
  );
  const canListEmployeePortals = access.node?.__typename === "Organization"
    && access.node.canListEmployeePortals;

  const data = useLazyLoadQuery<useDefaultEmployeePortalIdQuery>(
    defaultEmployeePortalQuery,
    { organizationId, includePortals: canListEmployeePortals },
  );

  const organizationKey: useDefaultEmployeePortalId_organization$key | null
    = canListEmployeePortals
      && data.node != null
      && data.node.__typename === "Organization"
      ? data.node
      : null;
  const organization = useFragment(defaultEmployeePortalFragment, organizationKey);

  return organization?.employeePortals.edges[0]?.node.id ?? null;
}
