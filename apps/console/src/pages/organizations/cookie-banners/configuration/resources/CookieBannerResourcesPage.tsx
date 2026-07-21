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

import type { CookieBannerResourcesPageFragment$key } from "#/__generated__/core/CookieBannerResourcesPageFragment.graphql";
import type { CookieBannerResourcesPageQuery } from "#/__generated__/core/CookieBannerResourcesPageQuery.graphql";
import type {
  CookieBannerResourcesPageRefetchQuery,
  TrackerResourceOrderField,
  TrackerResourceType,
} from "#/__generated__/core/CookieBannerResourcesPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { TrackerResourceRow } from "./_components/TrackerResourceRow";

export const cookieBannerResourcesPageQuery = graphql`
  query CookieBannerResourcesPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        ...CookieBannerResourcesPageFragment
      }
    }
  }
`;

const resourcesFragment = graphql`
  fragment CookieBannerResourcesPageFragment on CookieBanner
  @refetchable(queryName: "CookieBannerResourcesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "TrackerResourceOrder", defaultValue: { field: LAST_DETECTED_AT, direction: DESC } }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    query: { type: "String", defaultValue: null }
    type: { type: "TrackerResourceType", defaultValue: null }
  ) {
    uncategorisedTrackerResources(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: { query: $query, type: $type }
    )
      @connection(
        key: "CookieBannerResourcesPage_uncategorisedTrackerResources"
        filters: ["filter", "orderBy"]
      )
      @required(action: THROW) {
      __id
      edges {
        node {
          id
          ...TrackerResourceRowFragment
        }
      }
    }
  }
`;

interface CookieBannerResourcesPageProps {
  queryRef: PreloadedQuery<CookieBannerResourcesPageQuery>;
}

export default function CookieBannerResourcesPage({
  queryRef,
}: CookieBannerResourcesPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const data = usePreloadedQuery<CookieBannerResourcesPageQuery>(cookieBannerResourcesPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const [isPending, startTransition] = useTransition();
  const [queryFilter, setQueryFilter] = useState("");
  const [typeFilter, setTypeFilter] = useState<TrackerResourceType | null>(null);

  const { data: fragmentData, ...pagination } = usePaginationFragment<
    CookieBannerResourcesPageRefetchQuery,
    CookieBannerResourcesPageFragment$key
  >(resourcesFragment, data.node);

  const connectionId = fragmentData.uncategorisedTrackerResources.__id;
  const resources = fragmentData.uncategorisedTrackerResources.edges.map(edge => edge.node) ?? [];

  const refetchFilters = (overrides: Record<string, unknown> = {}) => {
    startTransition(() => {
      pagination.refetch(
        {
          query: queryFilter || null,
          type: typeFilter,
          ...overrides,
        },
        { fetchPolicy: "network-only" },
      );
    });
  };

  const handleQuerySubmit = () => {
    refetchFilters({ query: queryFilter || null });
  };

  const handleTypeFilterChange = (value: string) => {
    const newType = value === "ALL" ? null : (value as TrackerResourceType);
    setTypeFilter(newType);
    refetchFilters({ type: newType });
  };

  const refetchWithFilters: ComponentProps<typeof SortableTable>["refetch"] = ({ order }) => {
    pagination.refetch({
      order: { direction: order.direction, field: order.field as TrackerResourceOrderField },
      query: queryFilter || null,
      type: typeFilter,
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Input
          placeholder={t("resourcesPage.filters.search")}
          value={queryFilter}
          onChange={e => setQueryFilter(e.target.value)}
          onKeyDown={e => e.key === "Enter" && handleQuerySubmit()}
          onBlur={handleQuerySubmit}
          className="w-72"
        />
        <Select
          value={typeFilter ?? "ALL"}
          onValueChange={handleTypeFilterChange}
        >
          <Option value="ALL">{t("resourcesPage.types.all")}</Option>
          <Option value="SCRIPT">{t("resourcesPage.types.script")}</Option>
          <Option value="IFRAME">{t("resourcesPage.types.iframe")}</Option>
          <Option value="IMAGE">{t("resourcesPage.types.image")}</Option>
          <Option value="STYLESHEET">{t("resourcesPage.types.stylesheet")}</Option>
          <Option value="FONT">{t("resourcesPage.types.font")}</Option>
          <Option value="BEACON">{t("resourcesPage.types.beacon")}</Option>
          <Option value="FETCH">{t("resourcesPage.types.fetch")}</Option>
          <Option value="MEDIA">{t("resourcesPage.types.media")}</Option>
          <Option value="SERVICE_WORKER">{t("resourcesPage.types.serviceWorker")}</Option>
        </Select>
      </div>

      <div className={isPending ? "opacity-50 pointer-events-none transition-opacity" : ""}>
        {resources.length > 0
          ? (
              <SortableTable
                {...pagination}
                refetch={refetchWithFilters}
                pageSize={50}
              >
                <Thead>
                  <Tr>
                    <Th>{t("resourcesPage.columns.type")}</Th>
                    <SortableTh field="ORIGIN">{t("resourcesPage.columns.origin")}</SortableTh>
                    <Th>{t("resourcesPage.columns.path")}</Th>
                    <Th>{t("resourcesPage.columns.category")}</Th>
                    <SortableTh field="LAST_DETECTED_AT">{t("resourcesPage.columns.lastDetected")}</SortableTh>
                    <Th className="w-px" />
                  </Tr>
                </Thead>
                <Tbody>
                  {resources.map(resource => (
                    <TrackerResourceRow
                      key={resource.id}
                      resourceKey={resource}
                      connectionId={connectionId}
                    />
                  ))}
                </Tbody>
              </SortableTable>
            )
          : (
              <Card padded>
                <div className="text-center py-12">
                  <h3 className="text-lg font-semibold mb-2">
                    {t("resourcesPage.empty.title")}
                  </h3>
                  <p className="text-txt-tertiary">
                    {t("resourcesPage.empty.description")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}
