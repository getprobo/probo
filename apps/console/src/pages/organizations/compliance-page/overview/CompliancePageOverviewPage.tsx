// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageOverviewPageQuery } from "#/__generated__/core/CompliancePageOverviewPageQuery.graphql";

import { CompliancePageNDASection } from "./_components/CompliancePageNDASection";
import { CompliancePageSlackSection } from "./_components/CompliancePageSlackSection";
import { CompliancePageStatusSection } from "./_components/CompliancePageStatusSection";

export const compliancePageOverviewPageQuery = graphql`
  query CompliancePageOverviewPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        compliancePage: trustCenter {
          canGetNDA: permission(action: "core:trust-center:get-nda")
        }
      }
      ...CompliancePageStatusSectionFragment
      ...CompliancePageNDASectionFragment
      ...CompliancePageSlackSectionFragment
    }
  }
`;

export function CompliancePageOverviewPage(props: { queryRef: PreloadedQuery<CompliancePageOverviewPageQuery> }) {
  const { queryRef } = props;

  const { organization } = usePreloadedQuery<CompliancePageOverviewPageQuery>(
    compliancePageOverviewPageQuery,
    queryRef,
  );

  return (
    <div className="space-y-6">
      <CompliancePageStatusSection fragmentRef={organization} />

      {organization.compliancePage?.canGetNDA && (
        <CompliancePageNDASection fragmentRef={organization} />
      )}

      <CompliancePageSlackSection fragmentRef={organization} />
    </div>
  );
}
