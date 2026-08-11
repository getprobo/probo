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

import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalAccessPageQuery } from "#/__generated__/core/CompliancePortalAccessPageQuery.graphql";

import { CompliancePortalAccessList } from "./_components/CompliancePortalAccessList";

export const compliancePortalAccessPageQuery = graphql`
  query CompliancePortalAccessPageQuery($compliancePortalId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        ...CompliancePortalAccessListFragment
      }
    }
  }
`;

interface CompliancePortalAccessPageProps {
  queryRef: PreloadedQuery<CompliancePortalAccessPageQuery>;
}

export default function CompliancePortalAccessPage({ queryRef }: CompliancePortalAccessPageProps) {
  const { t } = useTranslation("organizations/compliance-portals");

  const { compliancePortal } = usePreloadedQuery<CompliancePortalAccessPageQuery>(
    compliancePortalAccessPageQuery,
    queryRef,
  );
  if (compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid type for node");
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">
            {t("accessPage.title")}
          </h3>
          <p className="text-sm text-txt-tertiary">
            {t("accessPage.description")}
          </p>
        </div>
      </div>

      <CompliancePortalAccessList fragmentRef={compliancePortal} />
    </div>
  );
}
