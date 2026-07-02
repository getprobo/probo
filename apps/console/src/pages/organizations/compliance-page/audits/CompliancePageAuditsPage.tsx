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

import { useTranslate } from "@probo/i18n";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePageAuditsPageQuery } from "#/__generated__/core/CompliancePageAuditsPageQuery.graphql";

import { CompliancePageAuditList } from "./_components/CompliancePageAuditList";

export const compliancePageAuditsPageQuery = graphql`
  query CompliancePageAuditsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ...CompliancePageAuditListFragment
    }
  }
`;

export function CompliancePageAuditsPage(props: { queryRef: PreloadedQuery<CompliancePageAuditsPageQuery> }) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<CompliancePageAuditsPageQuery>(compliancePageAuditsPageQuery, queryRef);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Audits")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Manage audit reports and compliance certifications")}
          </p>
        </div>
      </div>

      <CompliancePageAuditList fragmentRef={organization} />
    </div>
  );
}
