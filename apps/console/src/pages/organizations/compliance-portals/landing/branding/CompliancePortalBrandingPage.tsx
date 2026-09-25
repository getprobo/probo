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

import type { CompliancePortalBrandingPageQuery } from "#/__generated__/core/CompliancePortalBrandingPageQuery.graphql";

import { CompliancePortalCustomLinksSection } from "../_components/CompliancePortalCustomLinksSection";
import { CompliancePortalProfileSection } from "../_components/CompliancePortalProfileSection";

export const compliancePortalBrandingPageQuery = graphql`
  query CompliancePortalBrandingPageQuery($compliancePortalId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        ...CompliancePortalProfileSection_compliancePortalFragment
        ...CompliancePortalCustomLinksSection_compliancePortalFragment
      }
    }
  }
`;

interface CompliancePortalBrandingPageProps {
  queryRef: PreloadedQuery<CompliancePortalBrandingPageQuery>;
}

export function CompliancePortalBrandingPage({ queryRef }: CompliancePortalBrandingPageProps) {
  const { compliancePortal } = usePreloadedQuery<CompliancePortalBrandingPageQuery>(
    compliancePortalBrandingPageQuery,
    queryRef,
  );
  if (compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid type for node");
  }

  return (
    <div className="space-y-8">
      <CompliancePortalProfileSection compliancePortalRef={compliancePortal} />
      <CompliancePortalCustomLinksSection compliancePortalRef={compliancePortal} />
    </div>
  );
}
