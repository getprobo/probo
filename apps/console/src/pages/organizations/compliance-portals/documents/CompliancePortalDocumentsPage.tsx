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

import type { CompliancePortalDocumentsPageQuery } from "#/__generated__/core/CompliancePortalDocumentsPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { CompliancePortalDocumentList } from "./_components/CompliancePortalDocumentList";

export const compliancePortalDocumentsPageQuery = graphql`
  query CompliancePortalDocumentsPageQuery(
    $organizationId: ID!
    $compliancePortalId: ID!
  ) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...CompliancePortalDocumentList_organization
          @arguments(compliancePortalId: $compliancePortalId)
      }
    }
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        ...CompliancePortalDocumentList_compliancePortal
      }
    }
  }
`;

interface CompliancePortalDocumentsPageProps {
  queryRef: PreloadedQuery<CompliancePortalDocumentsPageQuery>;
}

export function CompliancePortalDocumentsPage({ queryRef }: CompliancePortalDocumentsPageProps) {
  const data = usePreloadedQuery<CompliancePortalDocumentsPageQuery>(
    compliancePortalDocumentsPageQuery,
    queryRef,
  );
  if (data.organization?.__typename !== "Organization") {
    throw new NotFoundError("Organization not found");
  }
  if (data.compliancePortal?.__typename !== "CompliancePortal") {
    throw new NotFoundError("Compliance portal not found");
  }

  return (
    <CompliancePortalDocumentList
      organizationKey={data.organization}
      compliancePortalKey={data.compliancePortal}
    />
  );
}
