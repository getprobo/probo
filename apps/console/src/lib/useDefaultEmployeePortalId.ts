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

import { graphql, useLazyLoadQuery } from "react-relay";

import type { useDefaultEmployeePortalIdQuery } from "#/__generated__/core/useDefaultEmployeePortalIdQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const defaultEmployeePortalQuery = graphql`
  query useDefaultEmployeePortalIdQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
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
    }
  }
`;

export function useDefaultEmployeePortalId(): string | null {
  const organizationId = useOrganizationId();
  const data = useLazyLoadQuery<useDefaultEmployeePortalIdQuery>(
    defaultEmployeePortalQuery,
    { organizationId },
  );

  if (data.node == null || !("employeePortals" in data.node)) {
    return null;
  }

  return data.node.employeePortals?.edges[0]?.node.id ?? null;
}
