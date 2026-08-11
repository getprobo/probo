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
import { graphql } from "relay-runtime";

import type { CompliancePortalBrandPageQuery } from "#/__generated__/core/CompliancePortalBrandPageQuery.graphql";

import { CompliancePortalCustomLinksSection } from "./_components/CompliancePortalCustomLinksSection";
import { CompliancePortalDomainsSection } from "./_components/CompliancePortalDomainsSection";
import { CompliancePortalProfileSection } from "./_components/CompliancePortalProfileSection";
import { CompliancePortalVisualIdentitySection } from "./_components/CompliancePortalVisualIdentitySection";

export const compliancePortalBrandPageQuery = graphql`
  query CompliancePortalBrandPageQuery($compliancePortalId: ID!, $organizationId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        ...CompliancePortalProfileSection_compliancePortalFragment
        ...CompliancePortalVisualIdentitySection_compliancePortalFragment
        ...CompliancePortalCustomLinksSection_compliancePortalFragment
        ...CompliancePortalDomainsSection_compliancePortalFragment
      }
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...CompliancePortalDomainsSection_organizationFragment
      }
    }
  }
`;

interface CompliancePortalBrandPageProps {
  queryRef: PreloadedQuery<CompliancePortalBrandPageQuery>;
}

export default function CompliancePortalBrandPage({ queryRef }: CompliancePortalBrandPageProps) {
  const data = usePreloadedQuery<CompliancePortalBrandPageQuery>(compliancePortalBrandPageQuery, queryRef);
  if (data.compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid type for node");
  }
  if (data.organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const { compliancePortal, organization } = data;

  return (
    <div className="space-y-8">
      <CompliancePortalProfileSection compliancePortalRef={compliancePortal} />

      <CompliancePortalDomainsSection
        organizationRef={organization}
        compliancePortalRef={compliancePortal}
      />

      <CompliancePortalVisualIdentitySection compliancePortalRef={compliancePortal} />

      <CompliancePortalCustomLinksSection compliancePortalRef={compliancePortal} />
    </div>
  );
}
