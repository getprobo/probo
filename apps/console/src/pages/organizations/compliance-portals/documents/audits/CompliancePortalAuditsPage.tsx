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

import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePortalAuditsPageQuery } from "#/__generated__/core/CompliancePortalAuditsPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { CompliancePortalAuditList } from "./_components/CompliancePortalAuditList";

export const compliancePortalAuditsPageQuery = graphql`
  query CompliancePortalAuditsPageQuery(
    $organizationId: ID!
    $compliancePortalId: ID!
  ) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...CompliancePortalAuditList_organization
          @arguments(compliancePortalId: $compliancePortalId)
      }
    }
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        ...CompliancePortalAuditList_compliancePortal
      }
    }
  }
`;

interface CompliancePortalAuditsPageProps {
  queryRef: PreloadedQuery<CompliancePortalAuditsPageQuery>;
}

export function CompliancePortalAuditsPage({ queryRef }: CompliancePortalAuditsPageProps) {
  const data = usePreloadedQuery<CompliancePortalAuditsPageQuery>(
    compliancePortalAuditsPageQuery,
    queryRef,
  );
  if (data.organization?.__typename !== "Organization") {
    throw new NotFoundError("Organization not found");
  }
  if (data.compliancePortal?.__typename !== "CompliancePortal") {
    throw new NotFoundError("Compliance portal not found");
  }

  return (
    <CompliancePortalAuditList
      organizationKey={data.organization}
      compliancePortalKey={data.compliancePortal}
    />
  );
}
