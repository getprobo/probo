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
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation("organizations/cookie-banners");
  const data = usePreloadedQuery<CookieBannerConsentRecordsPageQuery>(cookieBannerConsentRecordsPageQuery, queryRef);

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
          <Option value="ALL">{t("consentRecordsPage.filters.allActions")}</Option>
          <Option value="ACCEPT_ALL">{t("consentRecordsPage.actions.acceptAll")}</Option>
          <Option value="REJECT_ALL">{t("consentRecordsPage.actions.rejectAll")}</Option>
          <Option value="CUSTOMIZE">{t("consentRecordsPage.actions.customize")}</Option>
          <Option value="GPC">{t("consentRecordsPage.actions.gpc")}</Option>
        </Select>
        <Input
          placeholder={t("consentRecordsPage.filters.visitorId")}
          value={visitorIdFilter}
          onChange={e => setVisitorIdFilter(e.target.value)}
          onKeyDown={e => e.key === "Enter" && handleVisitorIdSubmit()}
          onBlur={handleVisitorIdSubmit}
          className="w-48"
        />
        <Input
          placeholder={t("consentRecordsPage.filters.bannerVersion")}
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
                    <Th>{t("consentRecordsPage.columns.visitorId")}</Th>
                    <Th>{t("consentRecordsPage.columns.action")}</Th>
                    <Th>{t("consentRecordsPage.columns.bannerVersion")}</Th>
                    <Th>{t("consentRecordsPage.columns.ipAddress")}</Th>
                    <Th>{t("consentRecordsPage.columns.sdkVersion")}</Th>
                    <Th>{t("consentRecordsPage.columns.regulation")}</Th>
                    <Th>{t("consentRecordsPage.columns.source")}</Th>
                    <Th>{t("consentRecordsPage.columns.country")}</Th>
                    <SortableTh field="CREATED_AT">{t("consentRecordsPage.columns.date")}</SortableTh>
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
                    {t("consentRecordsPage.empty.title")}
                  </h3>
                  <p className="text-txt-tertiary">
                    {t("consentRecordsPage.empty.description")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}
