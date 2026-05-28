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
  const { __ } = useTranslate();
  const data = usePreloadedQuery(cookieBannerTrackersPageQuery, queryRef);

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
          placeholder={__("Search by name or description...")}
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
          <Option value="ALL">{__("All third parties")}</Option>
          {linkedThirdParties.map(party => (
            <Option key={party.id} value={party.id}>{party.name}</Option>
          ))}
        </Select>
        <Select
          value={trackerTypeFilter ?? "ALL"}
          onValueChange={handleTrackerTypeFilterChange}
        >
          <Option value="ALL">{__("All types")}</Option>
          <Option value="COOKIE">{__("Cookie")}</Option>
          <Option value="LOCAL_STORAGE">{__("localStorage")}</Option>
          <Option value="SESSION_STORAGE">{__("sessionStorage")}</Option>
          <Option value="INDEXED_DB">{__("IndexedDB")}</Option>
          <Option value="CACHE_STORAGE">{__("Cache Storage")}</Option>
        </Select>
        <Select
          value={sourceFilter ?? "ALL"}
          onValueChange={handleSourceFilterChange}
        >
          <Option value="ALL">{__("All sources")}</Option>
          <Option value="SCRIPT">{__("Script")}</Option>
          <Option value="PRE_EXISTING">{__("Pre-existing")}</Option>
          <Option value="EXTENSION">{__("Extension")}</Option>
        </Select>
        <Select
          value={categoryFilter ?? "ALL"}
          onValueChange={handleCategoryFilterChange}
        >
          <Option value="ALL">{__("All categories")}</Option>
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
                    <Th>{__("Type")}</Th>
                    <SortableTh field="NAME">{__("Name")}</SortableTh>
                    <Th>{__("Third party")}</Th>
                    <SortableTh field="SOURCE">{__("Source")}</SortableTh>
                    <Th>{__("Category")}</Th>
                    <SortableTh field="LAST_MATCHED_AT">{__("Last Matched")}</SortableTh>
                    <Th className="w-28" />
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
                    {__("No tracker patterns")}
                  </h3>
                  <p className="text-txt-tertiary">
                    {__("No cookie patterns have been detected yet. Patterns will appear here when detected.")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}
