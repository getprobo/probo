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
import { ListSkeleton } from "@probo/ui/src/v2/List/ListSkeleton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalVisitorsPageQuery } from "#/__generated__/core/CompliancePortalVisitorsPageQuery.graphql";

import { CompliancePortalPageHeader } from "../_components/CompliancePortalPageHeader";

import { CompliancePortalAccessList } from "./_components/CompliancePortalAccessList";
import { CompliancePortalNDASection } from "./_components/CompliancePortalNDASection";
import { accessSection, visitorsPage } from "./variants";

export const compliancePortalVisitorsPageQuery = graphql`
  query CompliancePortalVisitorsPageQuery($compliancePortalId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        canGetNDA: permission(action: "compliance-portal:portal:get-nda")
        canListAccesses: permission(action: "compliance-portal:portal-access:list")
        ...CompliancePortalNDASectionFragment
      }
    }
  }
`;

interface CompliancePortalVisitorsPageProps {
  queryRef: PreloadedQuery<CompliancePortalVisitorsPageQuery>;
}

export function CompliancePortalVisitorsPage({ queryRef }: CompliancePortalVisitorsPageProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const title = t("visitorsPage.title");
  usePageTitle(title);
  const { root, intro } = accessSection();

  const { compliancePortal } = usePreloadedQuery<CompliancePortalVisitorsPageQuery>(
    compliancePortalVisitorsPageQuery,
    queryRef,
  );
  if (compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid type for node");
  }

  return (
    <div className={visitorsPage()}>
      <CompliancePortalPageHeader
        title={title}
        description={t("visitorsPage.description")}
      />
      {compliancePortal.canGetNDA && (
        <CompliancePortalNDASection compliancePortalKey={compliancePortal} />
      )}

      {compliancePortal.canListAccesses && (
        <section className={root()}>
          <div className={intro()}>
            <Text size={2} color="neutral">
              {t("accessPage.description")}
            </Text>
          </div>
          <Suspense fallback={<ListSkeleton count={4} />}>
            <CompliancePortalAccessList />
          </Suspense>
        </section>
      )}
    </div>
  );
}
