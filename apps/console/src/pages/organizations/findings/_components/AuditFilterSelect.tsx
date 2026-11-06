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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectList } from "@probo/ui/src/v2/Select/SelectList";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { AuditFilterSelect_organization$key } from "#/__generated__/core/AuditFilterSelect_organization.graphql";
import type { AuditFilterSelectPaginationQuery } from "#/__generated__/core/AuditFilterSelectPaginationQuery.graphql";
import type { AuditFilterSelectQuery } from "#/__generated__/core/AuditFilterSelectQuery.graphql";
import type { AuditFrameworkNames_audit$key } from "#/__generated__/core/AuditFrameworkNames_audit.graphql";
import { useAuditFrameworkNames } from "#/components/audits/AuditFrameworkNames";

import { auditFilterSelect } from "./variants";

const auditFilterPageSize = 20;

const auditFilterSelectQuery = graphql`
  query AuditFilterSelectQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...AuditFilterSelect_organization
      }
    }
  }
`;

const auditFilterSelectFragment = graphql`
  fragment AuditFilterSelect_organization on Organization
  @refetchable(queryName: "AuditFilterSelectPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    audits(first: $first, after: $after, orderBy: { field: CREATED_AT, direction: DESC })
      @connection(key: "AuditFilterSelect_audits") {
      edges {
        node {
          id
          name
          ...AuditFrameworkNames_audit
        }
      }
    }
  }
`;

interface AuditFilterSelectProps {
  organizationId: string;
  value: string | null;
  onChange: (value: string | null) => void;
}

export function AuditFilterSelect(props: AuditFilterSelectProps) {
  const { t } = useTranslation("organizations/findings");

  return (
    <Suspense
      fallback={(
        <Select value={null} disabled>
          <SelectTrigger
            size={2}
            placeholder={t("list.filters.allAudits")}
            aria-label={t("list.filters.allAudits")}
          />
        </Select>
      )}
    >
      <AuditFilterSelectQueryRenderer {...props} />
    </Suspense>
  );
}

function AuditFilterSelectQueryRenderer({ organizationId, ...props }: AuditFilterSelectProps) {
  const data = useLazyLoadQuery<AuditFilterSelectQuery>(
    auditFilterSelectQuery,
    { organizationId },
  );

  if (data.organization?.__typename !== "Organization") {
    return null;
  }

  return <AuditFilterSelectOptions organizationKey={data.organization} {...props} />;
}

function AuditFilterSelectOptions({
  organizationKey,
  value,
  onChange,
}: Omit<AuditFilterSelectProps, "organizationId"> & {
  organizationKey: AuditFilterSelect_organization$key;
}) {
  const { t } = useTranslation("organizations/findings");
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment<
    AuditFilterSelectPaginationQuery,
    AuditFilterSelect_organization$key
  >(auditFilterSelectFragment, organizationKey);
  const audits = data.audits?.edges.map(edge => edge.node) ?? [];
  const { loadMore } = auditFilterSelect();

  return (
    <Select<string | null> value={value} onValueChange={onChange}>
      <SelectTrigger
        size={2}
        placeholder={t("list.filters.allAudits")}
        aria-label={t("list.filters.allAudits")}
      >
        {(selected: string | null) => {
          if (selected == null) {
            return t("list.filters.allAudits");
          }
          const audit = audits.find(item => item.id === selected);
          return audit ? <AuditLabel audit={audit} /> : selected;
        }}
      </SelectTrigger>
      <SelectPopup align="end">
        <SelectList>
          <SelectItem value={null}>{t("list.filters.allAudits")}</SelectItem>
          {audits.map(audit => (
            <SelectItem key={audit.id} value={audit.id}>
              <AuditLabel audit={audit} />
            </SelectItem>
          ))}
        </SelectList>
        {hasNext && (
          <Button
            variant="ghost"
            color="neutral"
            size={1}
            className={loadMore()}
            loading={isLoadingNext}
            onClick={() => {
              loadNext(auditFilterPageSize);
            }}
          >
            {t("list.filters.loadMoreAudits")}
          </Button>
        )}
      </SelectPopup>
    </Select>
  );
}

function AuditLabel({
  audit,
}: {
  audit: AuditFrameworkNames_audit$key & { id: string; name?: string | null };
}) {
  const { label: frameworkLabel } = useAuditFrameworkNames(audit);
  return audit.name || frameworkLabel || audit.id;
}
