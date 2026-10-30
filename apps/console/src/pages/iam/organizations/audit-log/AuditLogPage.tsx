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

import type { AuditLogPageQuery } from "#/__generated__/iam/AuditLogPageQuery.graphql";

import { AuditLogList } from "./_components/AuditLogList/AuditLogList";
import { auditLogPage } from "./_components/variants";
import { auditLogFilterKey } from "./_lib/useAuditLogListFilters";

const AUDIT_LOG_NS = "iam/organizations/audit-log";

export const auditLogPageQuery = graphql`
  query AuditLogPageQuery(
    $organizationId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
    $filter: AuditLogEntryFilter
  ) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...AuditLogList_organization
          @arguments(
            first: $first
            after: $after
            last: $last
            before: $before
            filter: $filter
          )
      }
    }
  }
`;

interface AuditLogPageProps {
  queryRef: PreloadedQuery<AuditLogPageQuery>;
}

export function AuditLogPage({ queryRef }: AuditLogPageProps) {
  const { t } = useTranslation(AUDIT_LOG_NS);
  const { root } = auditLogPage();
  usePageTitle(t("auditLogPage.title"));

  const { organization } = usePreloadedQuery<AuditLogPageQuery>(
    auditLogPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("node is of invalid type");
  }

  return (
    <div className={root()}>
      <AuditLogList
        organizationKey={organization}
        loadedFilterKey={auditLogFilterKey(queryRef.variables.filter)}
      />
    </div>
  );
}
