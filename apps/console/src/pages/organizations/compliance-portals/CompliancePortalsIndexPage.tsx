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

import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Navigate } from "react-router";
import { graphql } from "relay-runtime";

import type { CompliancePortalsIndexPageQuery } from "#/__generated__/core/CompliancePortalsIndexPageQuery.graphql";

export const compliancePortalsIndexPageQuery = graphql`
  query CompliancePortalsIndexPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        compliancePortals(first: 1, orderBy: { field: CREATED_AT, direction: DESC })
          @required(action: THROW) {
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

interface CompliancePortalsIndexPageProps {
  queryRef: PreloadedQuery<CompliancePortalsIndexPageQuery>;
}

/**
 * The rail and role landings hit /compliance-portals. Send them to the newest
 * portal when one exists; otherwise the create form.
 */
export function CompliancePortalsIndexPage({ queryRef }: CompliancePortalsIndexPageProps) {
  const { organization } = usePreloadedQuery<CompliancePortalsIndexPageQuery>(
    compliancePortalsIndexPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const portalId = organization.compliancePortals.edges[0]?.node.id;
  if (portalId == null) {
    return <Navigate to="new" replace />;
  }

  return <Navigate to={portalId} replace />;
}
