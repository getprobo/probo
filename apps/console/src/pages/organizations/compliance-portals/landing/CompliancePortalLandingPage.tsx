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

import type { CompliancePortalLandingPageQuery } from "#/__generated__/core/CompliancePortalLandingPageQuery.graphql";

import { CompliancePortalCommitmentGroupList } from "./_components/CompliancePortalCommitmentGroupList";
import { CompliancePortalCustomLinksSection } from "./_components/CompliancePortalCustomLinksSection";
import { CompliancePortalFrameworksSection } from "./_components/CompliancePortalFrameworksSection";
import { CompliancePortalProfileSection } from "./_components/CompliancePortalProfileSection";
import { CompliancePortalReferencesSection } from "./_components/CompliancePortalReferencesSection";
import { CompliancePortalVisualIdentitySection } from "./_components/CompliancePortalVisualIdentitySection";

export const compliancePortalLandingPageQuery = graphql`
  query CompliancePortalLandingPageQuery($compliancePortalId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        id
        canListCustomLinks: permission(action: "compliance-portal:compliance-custom-link:list")
        canListFrameworks: permission(action: "compliance-portal:compliance-framework:list")
        canListCommitmentGroups: permission(action: "compliance-portal:commitment-group:list")
        canListReferences: permission(action: "compliance-portal:portal-reference:list")
        canCreateGroup: permission(action: "compliance-portal:commitment-group:create")
        ...CompliancePortalProfileSection_compliancePortalFragment
        ...CompliancePortalVisualIdentitySection_compliancePortalFragment
        ...CompliancePortalCustomLinksSection_compliancePortalFragment
        ...CompliancePortalFrameworksSectionFragment
        ...CompliancePortalCommitmentGroupListFragment
        ...CompliancePortalReferencesSectionFragment
      }
    }
  }
`;

interface CompliancePortalLandingPageProps {
  queryRef: PreloadedQuery<CompliancePortalLandingPageQuery>;
}

export function CompliancePortalLandingPage({ queryRef }: CompliancePortalLandingPageProps) {
  const { compliancePortal } = usePreloadedQuery<CompliancePortalLandingPageQuery>(
    compliancePortalLandingPageQuery,
    queryRef,
  );
  if (compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid type for node");
  }

  return (
    <div className="space-y-8">
      {compliancePortal.canListCustomLinks && (
        <>
          <CompliancePortalProfileSection compliancePortalRef={compliancePortal} />
          <CompliancePortalVisualIdentitySection compliancePortalRef={compliancePortal} />
          <CompliancePortalCustomLinksSection compliancePortalRef={compliancePortal} />
        </>
      )}

      {compliancePortal.canListFrameworks && (
        <CompliancePortalFrameworksSection fragmentRef={compliancePortal} />
      )}

      {compliancePortal.canListCommitmentGroups && (
        <CompliancePortalCommitmentGroupList
          fragmentRef={compliancePortal}
          compliancePortalId={compliancePortal.id}
          canCreate={compliancePortal.canCreateGroup}
        />
      )}

      {compliancePortal.canListReferences && (
        <CompliancePortalReferencesSection fragmentRef={compliancePortal} />
      )}
    </div>
  );
}
