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

import { usePageTitle } from "@probo/hooks";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalHostingPageQuery } from "#/__generated__/core/CompliancePortalHostingPageQuery.graphql";

import { CompliancePortalPageHeader } from "../_components/CompliancePortalPageHeader";

import { CompliancePortalDomainsSection } from "./_components/CompliancePortalDomainsSection";
import { CompliancePortalStatusSection } from "./_components/CompliancePortalStatusSection";
import { hostingPage } from "./variants";

export const compliancePortalHostingPageQuery = graphql`
  query CompliancePortalHostingPageQuery($compliancePortalId: ID!, $organizationId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        ...CompliancePortalStatusSectionFragment
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

interface CompliancePortalHostingPageProps {
  queryRef: PreloadedQuery<CompliancePortalHostingPageQuery>;
}

export function CompliancePortalHostingPage({ queryRef }: CompliancePortalHostingPageProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const title = t("hostingPage.title");
  usePageTitle(title);

  const data = usePreloadedQuery<CompliancePortalHostingPageQuery>(
    compliancePortalHostingPageQuery,
    queryRef,
  );
  if (data.compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid type for node");
  }
  if (data.organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const { compliancePortal, organization } = data;

  return (
    <div className={hostingPage()}>
      <CompliancePortalPageHeader
        title={title}
        description={t("hostingPage.description")}
      />
      <CompliancePortalStatusSection compliancePortalKey={compliancePortal} />
      <CompliancePortalDomainsSection
        organizationKey={organization}
        compliancePortalKey={compliancePortal}
      />
    </div>
  );
}
