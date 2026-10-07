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

import { Option, Select } from "@probo/ui";
import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { AuditFilterSelectFragment$key } from "#/__generated__/core/AuditFilterSelectFragment.graphql";
import type { AuditFilterSelectPaginationQuery } from "#/__generated__/core/AuditFilterSelectPaginationQuery.graphql";
import type { AuditFilterSelectQuery } from "#/__generated__/core/AuditFilterSelectQuery.graphql";

const PAGE_SIZE = 20;
const LOAD_MORE_VALUE = "__LOAD_MORE__";

const auditFilterSelectQuery = graphql`
  query AuditFilterSelectQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...AuditFilterSelectFragment
      }
    }
  }
`;

const auditFilterSelectFragment = graphql`
  fragment AuditFilterSelectFragment on Organization
  @refetchable(queryName: "AuditFilterSelectPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    audits(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: DESC }
    ) @connection(key: "AuditFilterSelect_audits") {
      edges {
        node {
          id
          name
          framework {
            name
          }
        }
      }
    }
  }
`;

type AuditFilterSelectProps = {
  organizationId: string;
  value: string | null;
  onChange: (value: string) => void;
};

export function AuditFilterSelect({
  organizationId,
  value,
  onChange,
}: AuditFilterSelectProps) {
  const { t } = useTranslation();
  const query = useLazyLoadQuery<AuditFilterSelectQuery>(
    auditFilterSelectQuery,
    { organizationId },
  );
  const { data, loadNext, hasNext, isLoadingNext }
    = usePaginationFragment<
      AuditFilterSelectPaginationQuery,
      AuditFilterSelectFragment$key
    >(auditFilterSelectFragment, query.organization as AuditFilterSelectFragment$key);
  const audits = data.audits?.edges?.map(edge => edge.node) ?? [];
  const [isOpen, setIsOpen] = useState(false);
  // Selecting an option closes the popup; loading a page must not.
  const keepOpenRef = useRef(false);

  const handleValueChange = (next: string) => {
    if (next === LOAD_MORE_VALUE) {
      keepOpenRef.current = true;
      loadNext(PAGE_SIZE);
      return;
    }

    onChange(next);
  };

  const handleOpenChange = (open: boolean) => {
    if (!open && keepOpenRef.current) {
      keepOpenRef.current = false;
      return;
    }

    setIsOpen(open);
  };

  return (
    <Select
      value={value ?? "ALL"}
      open={isOpen}
      onOpenChange={handleOpenChange}
      onValueChange={handleValueChange}
    >
      <Option value="ALL">{t("findingsPage.filters.allAudits")}</Option>
      {audits.map(audit => (
        <Option key={audit.id} value={audit.id}>
          {audit.name || audit.framework?.name || audit.id}
        </Option>
      ))}
      {hasNext && (
        <Option
          value={LOAD_MORE_VALUE}
          className="justify-center text-txt-secondary"
        >
          {isLoadingNext
            ? t("findingsPage.loading")
            : t("findingsPage.actions.loadMore")}
        </Option>
      )}
    </Select>
  );
}
