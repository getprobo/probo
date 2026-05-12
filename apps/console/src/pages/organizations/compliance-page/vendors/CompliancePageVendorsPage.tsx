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

import { useTranslate } from "@probo/i18n";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { CompliancePageVendorsPageQuery } from "#/__generated__/core/CompliancePageVendorsPageQuery.graphql";

import { CompliancePageVendorList } from "./_components/CompliancePageVendorList";

export const compliancePageVendorsPageQuery = graphql`
  query CompliancePageVendorsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ...CompliancePageVendorListFragment
    }
  }
`;

export function CompliancePageVendorsPage(props: {
  queryRef: PreloadedQuery<CompliancePageVendorsPageQuery>;
}) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<CompliancePageVendorsPageQuery>(
    compliancePageVendorsPageQuery,
    queryRef,
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Subprocessors")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Manage subprocessor assessments and third-party risk information")}
          </p>
        </div>
      </div>

      <CompliancePageVendorList fragmentRef={organization} />
    </div>
  );
}
