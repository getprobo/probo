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
import {
  Card,
  Input,
  Option,
  Select,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { type ComponentProps, useState, useTransition } from "react";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { CookieBannerConsentRecordsPageFragment$key } from "#/__generated__/core/CookieBannerConsentRecordsPageFragment.graphql";
import type { CookieBannerConsentRecordsPageQuery } from "#/__generated__/core/CookieBannerConsentRecordsPageQuery.graphql";
import type {
  CookieBannerConsentRecordsPageRefetchQuery,
  CookieConsentAction,
  CookieConsentRecordOrderField,
} from "#/__generated__/core/CookieBannerConsentRecordsPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { ConsentRecordRow } from "./_components/ConsentRecordRow";

export const cookieBannerConsentRecordsPageQuery = graphql`
  query CookieBannerConsentRecordsPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        ...CookieBannerConsentRecordsPageFragment
      }
    }
  }
`;

const consentRecordsFragment = graphql`
  fragment CookieBannerConsentRecordsPageFragment on CookieBanner
  @refetchable(queryName: "CookieBannerConsentRecordsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "CookieConsentRecordOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    action: { type: "CookieConsentAction", defaultValue: null }
    visitorId: { type: "String", defaultValue: null }
    version: { type: "Int", defaultValue: null }
  ) {
    consentRecords(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: {
        action: $action
        visitorId: $visitorId
        version: $version
      }
    )
      @connection(
        key: "CookieBannerConsentRecordsPage_consentRecords"
        filters: ["filter", "orderBy"]
      ) @required(action: THROW) {
      edges {
        node {
          id
          ...ConsentRecordRowFragment
        }
      }
    }
  }
`;

interface CookieBannerConsentRecordsPageProps {
  queryRef: PreloadedQuery<CookieBannerConsentRecordsPageQuery>;
}

export default function CookieBannerConsentRecordsPage({
  queryRef,
}: CookieBannerConsentRecordsPageProps) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery(cookieBannerConsentRecordsPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const [isPending, startTransition] = useTransition();
  const [actionFilter, setActionFilter] = useState<CookieConsentAction | null>(null);
  const [visitorIdFilter, setVisitorIdFilter] = useState<string>("");
  const [versionFilter, setVersionFilter] = useState<string>("");
  const [versionError, setVersionError] = useState(false);

  const { data: fragmentData, ...pagination } = usePaginationFragment<
    CookieBannerConsentRecordsPageRefetchQuery,
    CookieBannerConsentRecordsPageFragment$key
  >(consentRecordsFragment, data.node);

  const records = fragmentData.consentRecords.edges.map(edge => edge.node) ?? [];

  const parseVersion = (v: string): number | null => {
    if (!v || !/^\d+$/.test(v)) return null;
    return parseInt(v, 10);
  };

  const refetchFilters = (overrides: Record<string, unknown> = {}) => {
    startTransition(() => {
      pagination.refetch(
        {
          action: actionFilter,
          visitorId: visitorIdFilter || null,
          version: parseVersion(versionFilter),
          ...overrides,
        },
        { fetchPolicy: "network-only" },
      );
    });
  };

  const handleActionFilterChange = (value: string) => {
    const newAction = value === "ALL" ? null : (value as CookieConsentAction);
    setActionFilter(newAction);
    refetchFilters({ action: newAction });
  };

  const handleVisitorIdSubmit = () => {
    refetchFilters({ visitorId: visitorIdFilter || null });
  };

  const handleVersionSubmit = () => {
    if (versionFilter && !/^\d+$/.test(versionFilter)) {
      setVersionError(true);
      return;
    }
    setVersionError(false);
    refetchFilters({ version: parseVersion(versionFilter) });
  };

  const refetchWithFilters: ComponentProps<typeof SortableTable>["refetch"] = ({ order }) => {
    pagination.refetch({
      order: { direction: order.direction, field: order.field as CookieConsentRecordOrderField },
      action: actionFilter,
      visitorId: visitorIdFilter || null,
      version: parseVersion(versionFilter),
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Select
          value={actionFilter ?? "ALL"}
          onValueChange={handleActionFilterChange}
        >
          <Option value="ALL">{__("All actions")}</Option>
          <Option value="ACCEPT_ALL">{__("Accept All")}</Option>
          <Option value="REJECT_ALL">{__("Reject All")}</Option>
          <Option value="CUSTOMIZE">{__("Customize")}</Option>
          <Option value="GPC">{__("GPC")}</Option>
        </Select>
        <Input
          placeholder={__("Visitor ID")}
          value={visitorIdFilter}
          onChange={e => setVisitorIdFilter(e.target.value)}
          onKeyDown={e => e.key === "Enter" && handleVisitorIdSubmit()}
          onBlur={handleVisitorIdSubmit}
          className="w-48"
        />
        <Input
          placeholder={__("Banner version")}
          value={versionFilter}
          invalid={versionError}
          onChange={(e) => {
            setVersionFilter(e.target.value);
            setVersionError(false);
          }}
          onKeyDown={e => e.key === "Enter" && handleVersionSubmit()}
          onBlur={handleVersionSubmit}
          className="w-48"
        />
      </div>

      <div className={isPending ? "opacity-50 pointer-events-none transition-opacity" : ""}>
        {records.length > 0
          ? (
              <SortableTable
                {...pagination}
                refetch={refetchWithFilters}
                pageSize={50}
              >
                <Thead>
                  <Tr>
                    <Th>{__("Visitor ID")}</Th>
                    <Th>{__("Action")}</Th>
                    <Th>{__("Banner Version")}</Th>
                    <Th>{__("IP Address")}</Th>
                    <Th>{__("SDK Version")}</Th>
                    <Th>{__("Regulation")}</Th>
                    <SortableTh field="CREATED_AT">{__("Date")}</SortableTh>
                  </Tr>
                </Thead>
                <Tbody>
                  {records.map(record => (
                    <ConsentRecordRow key={record.id} recordKey={record} />
                  ))}
                </Tbody>
              </SortableTable>
            )
          : (
              <Card padded>
                <div className="text-center py-12">
                  <h3 className="text-lg font-semibold mb-2">
                    {__("No consent records")}
                  </h3>
                  <p className="text-txt-tertiary">
                    {__("Consent records will appear here once visitors interact with your cookie banner.")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}
