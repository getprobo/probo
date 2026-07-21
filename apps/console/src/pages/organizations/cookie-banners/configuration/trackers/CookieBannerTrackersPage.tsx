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

import type { CookieBannerTrackersPageFragment$key } from "#/__generated__/core/CookieBannerTrackersPageFragment.graphql";
import type { CookieBannerTrackersPageQuery } from "#/__generated__/core/CookieBannerTrackersPageQuery.graphql";
import type {
  CookieBannerTrackersPageRefetchQuery,
  CookieSource,
  TrackerPatternOrderField,
  TrackerType,
} from "#/__generated__/core/CookieBannerTrackersPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { TrackerPatternRow } from "./_components/TrackerPatternRow";

export const cookieBannerTrackersPageQuery = graphql`
  query CookieBannerTrackersPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        ...CookieBannerTrackersPageFragment
        categories(first: 50, orderBy: { field: RANK, direction: ASC })
          @required(action: THROW) {
          edges {
            node {
              id
              name
            }
          }
        }
      }
    }
  }
`;

const trackersFragment = graphql`
  fragment CookieBannerTrackersPageFragment on CookieBanner
  @refetchable(queryName: "CookieBannerTrackersPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "TrackerPatternOrder", defaultValue: { field: NAME, direction: ASC } }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    query: { type: "String", defaultValue: null }
    source: { type: "CookieSource", defaultValue: null }
    trackerType: { type: "TrackerType", defaultValue: null }
    cookieCategoryId: { type: "ID", defaultValue: null }
    thirdPartyId: { type: "ID", defaultValue: null }
  ) {
    linkedThirdParties {
      __typename
      ... on ThirdParty {
        id
        name
      }
      ... on CommonThirdParty {
        id
        name
      }
    }
    trackerPatterns(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: { query: $query, source: $source, trackerType: $trackerType, cookieCategoryId: $cookieCategoryId, thirdPartyId: $thirdPartyId }
    )
      @connection(
        key: "CookieBannerTrackersPage_trackerPatterns"
        filters: ["filter", "orderBy"]
      )
      @required(action: THROW) {
      __id
      edges {
        node {
          id
          ...TrackerPatternRowFragment
        }
      }
    }
  }
`;

interface CookieBannerTrackersPageProps {
  queryRef: PreloadedQuery<CookieBannerTrackersPageQuery>;
}

export default function CookieBannerTrackersPage({
  queryRef,
}: CookieBannerTrackersPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const data = usePreloadedQuery<CookieBannerTrackersPageQuery>(cookieBannerTrackersPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const [isPending, startTransition] = useTransition();
  const [queryFilter, setQueryFilter] = useState("");
  const [sourceFilter, setSourceFilter] = useState<CookieSource | null>(null);
  const [trackerTypeFilter, setTrackerTypeFilter] = useState<TrackerType | null>(null);
  const [categoryFilter, setCategoryFilter] = useState<string | null>(null);
  const [thirdPartyFilter, setThirdPartyFilter] = useState<string | null>(null);

  const { data: fragmentData, ...pagination } = usePaginationFragment<
    CookieBannerTrackersPageRefetchQuery,
    CookieBannerTrackersPageFragment$key
  >(trackersFragment, data.node);

  const connectionId = fragmentData.trackerPatterns.__id;
  const patterns = fragmentData.trackerPatterns.edges.map(edge => edge.node) ?? [];
  const linkedThirdParties = (fragmentData.linkedThirdParties ?? []).filter(
    (party): party is Extract<typeof party, { id: string; name: string }> =>
      party.__typename === "ThirdParty"
      || party.__typename === "CommonThirdParty",
  );

  const categories = data.node.__typename === "CookieBanner"
    ? data.node.categories.edges.map(edge => edge.node)
    : [];

  const refetchFilters = (overrides: Record<string, unknown> = {}) => {
    startTransition(() => {
      pagination.refetch(
        {
          query: queryFilter || null,
          source: sourceFilter,
          trackerType: trackerTypeFilter,
          cookieCategoryId: categoryFilter,
          thirdPartyId: thirdPartyFilter,
          ...overrides,
        },
        { fetchPolicy: "network-only" },
      );
    });
  };

  const handleQuerySubmit = () => {
    refetchFilters({ query: queryFilter || null });
  };

  const handleSourceFilterChange = (value: string) => {
    const newSource = value === "ALL" ? null : (value as CookieSource);
    setSourceFilter(newSource);
    refetchFilters({ source: newSource });
  };

  const handleTrackerTypeFilterChange = (value: string) => {
    const newType = value === "ALL" ? null : (value as TrackerType);
    setTrackerTypeFilter(newType);
    refetchFilters({ trackerType: newType });
  };

  const handleCategoryFilterChange = (value: string) => {
    const newCategory = value === "ALL" ? null : value;
    setCategoryFilter(newCategory);
    refetchFilters({ cookieCategoryId: newCategory });
  };

  const handleThirdPartyFilterChange = (value: string) => {
    const newThirdParty = value === "ALL" ? null : value;
    setThirdPartyFilter(newThirdParty);
    refetchFilters({ thirdPartyId: newThirdParty });
  };

  const refetchWithFilters: ComponentProps<typeof SortableTable>["refetch"] = ({ order }) => {
    pagination.refetch({
      order: { direction: order.direction, field: order.field as TrackerPatternOrderField },
      query: queryFilter || null,
      source: sourceFilter,
      trackerType: trackerTypeFilter,
      cookieCategoryId: categoryFilter,
      thirdPartyId: thirdPartyFilter,
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Input
          placeholder={t("trackersPage.filters.search")}
          value={queryFilter}
          onChange={e => setQueryFilter(e.target.value)}
          onKeyDown={e => e.key === "Enter" && handleQuerySubmit()}
          onBlur={handleQuerySubmit}
          className="w-72"
        />
        <Select
          value={thirdPartyFilter ?? "ALL"}
          onValueChange={handleThirdPartyFilterChange}
        >
          <Option value="ALL">{t("trackersPage.filters.allThirdParties")}</Option>
          {linkedThirdParties.map(party => (
            <Option key={party.id} value={party.id}>{party.name}</Option>
          ))}
        </Select>
        <Select
          value={trackerTypeFilter ?? "ALL"}
          onValueChange={handleTrackerTypeFilterChange}
        >
          <Option value="ALL">{t("trackersPage.types.all")}</Option>
          <Option value="COOKIE">{t("trackersPage.types.cookie")}</Option>
          <Option value="LOCAL_STORAGE">{t("trackersPage.types.localStorage")}</Option>
          <Option value="SESSION_STORAGE">{t("trackersPage.types.sessionStorage")}</Option>
          <Option value="INDEXED_DB">{t("trackersPage.types.indexedDb")}</Option>
          <Option value="CACHE_STORAGE">{t("trackersPage.types.cacheStorage")}</Option>
        </Select>
        <Select
          value={sourceFilter ?? "ALL"}
          onValueChange={handleSourceFilterChange}
        >
          <Option value="ALL">{t("trackersPage.sources.all")}</Option>
          <Option value="SCRIPT">{t("trackersPage.sources.script")}</Option>
          <Option value="PRE_EXISTING">{t("trackersPage.sources.preExisting")}</Option>
          <Option value="HTTP">{t("trackersPage.sources.http")}</Option>
          <Option value="EXTENSION">{t("trackersPage.sources.extension")}</Option>
        </Select>
        <Select
          value={categoryFilter ?? "ALL"}
          onValueChange={handleCategoryFilterChange}
        >
          <Option value="ALL">{t("trackersPage.filters.allCategories")}</Option>
          {categories.map(category => (
            <Option key={category.id} value={category.id}>{category.name}</Option>
          ))}
        </Select>
      </div>

      <div className={isPending ? "opacity-50 pointer-events-none transition-opacity" : ""}>
        {patterns.length > 0
          ? (
              <SortableTable
                {...pagination}
                refetch={refetchWithFilters}
                pageSize={50}
              >
                <Thead>
                  <Tr>
                    <SortableTh field="NAME">{t("trackersPage.columns.name")}</SortableTh>
                    <Th>{t("trackersPage.columns.thirdParty")}</Th>
                    <SortableTh field="SOURCE">{t("trackersPage.columns.source")}</SortableTh>
                    <Th>{t("trackersPage.columns.category")}</Th>
                    <Th>{t("trackersPage.columns.maxAge")}</Th>
                    <SortableTh field="LAST_MATCHED_AT">{t("trackersPage.columns.lastMatched")}</SortableTh>
                    <Th className="w-px" />
                  </Tr>
                </Thead>
                <Tbody>
                  {patterns.map(pattern => (
                    <TrackerPatternRow
                      key={pattern.id}
                      patternKey={pattern}
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
                    {t("trackersPage.empty.title")}
                  </h3>
                  <p className="text-txt-tertiary">
                    {t("trackersPage.empty.description")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}
