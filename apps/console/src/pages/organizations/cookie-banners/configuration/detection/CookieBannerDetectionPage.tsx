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

import type { CookieBannerDetectionPageFragment$key } from "#/__generated__/core/CookieBannerDetectionPageFragment.graphql";
import type { CookieBannerDetectionPageQuery } from "#/__generated__/core/CookieBannerDetectionPageQuery.graphql";
import type {
  CookieBannerDetectionPageRefetchQuery,
  CookiePatternOrderField,
  CookieSource,
} from "#/__generated__/core/CookieBannerDetectionPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { DetectionPatternRow } from "./_components/DetectionPatternRow";

export const cookieBannerDetectionPageQuery = graphql`
  query CookieBannerDetectionPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        ...CookieBannerDetectionPageFragment
      }
    }
  }
`;

const detectionFragment = graphql`
  fragment CookieBannerDetectionPageFragment on CookieBanner
  @refetchable(queryName: "CookieBannerDetectionPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "CookiePatternOrder", defaultValue: { field: NAME, direction: ASC } }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    query: { type: "String", defaultValue: null }
    source: { type: "CookieSource", defaultValue: null }
  ) {
    uncategorisedPatterns(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: { query: $query, source: $source }
    )
      @connection(
        key: "CookieBannerDetectionPage_uncategorisedPatterns"
        filters: ["filter", "orderBy"]
      )
      @required(action: THROW) {
      edges {
        node {
          id
          ...DetectionPatternRowFragment
        }
      }
    }
  }
`;

interface CookieBannerDetectionPageProps {
  queryRef: PreloadedQuery<CookieBannerDetectionPageQuery>;
}

export default function CookieBannerDetectionPage({
  queryRef,
}: CookieBannerDetectionPageProps) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery(cookieBannerDetectionPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const [isPending, startTransition] = useTransition();
  const [queryFilter, setQueryFilter] = useState("");
  const [sourceFilter, setSourceFilter] = useState<CookieSource | null>(null);

  const { data: fragmentData, ...pagination } = usePaginationFragment<
    CookieBannerDetectionPageRefetchQuery,
    CookieBannerDetectionPageFragment$key
  >(detectionFragment, data.node);

  const patterns = fragmentData.uncategorisedPatterns.edges.map(edge => edge.node) ?? [];

  const refetchFilters = (overrides: Record<string, unknown> = {}) => {
    startTransition(() => {
      pagination.refetch(
        {
          query: queryFilter || null,
          source: sourceFilter,
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

  const refetchWithFilters: ComponentProps<typeof SortableTable>["refetch"] = ({ order }) => {
    pagination.refetch({
      order: { direction: order.direction, field: order.field as CookiePatternOrderField },
      query: queryFilter || null,
      source: sourceFilter,
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
          value={sourceFilter ?? "ALL"}
          onValueChange={handleSourceFilterChange}
        >
          <Option value="ALL">{__("All sources")}</Option>
          <Option value="SCRIPT">{__("Script")}</Option>
          <Option value="PRE_EXISTING">{__("Pre-existing")}</Option>
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
                    <SortableTh field="NAME">{__("Name")}</SortableTh>
                    <SortableTh field="SOURCE">{__("Source")}</SortableTh>
                    <SortableTh field="LAST_MATCHED_AT">{__("Last Matched")}</SortableTh>
                    <SortableTh field="UPDATED_AT">{__("Updated")}</SortableTh>
                  </Tr>
                </Thead>
                <Tbody>
                  {patterns.map(pattern => (
                    <DetectionPatternRow key={pattern.id} patternKey={pattern} />
                  ))}
                </Tbody>
              </SortableTable>
            )
          : (
              <Card padded>
                <div className="text-center py-12">
                  <h3 className="text-lg font-semibold mb-2">
                    {__("No uncategorised patterns")}
                  </h3>
                  <p className="text-txt-tertiary">
                    {__("All detected cookie patterns have been categorised. New patterns will appear here when detected.")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}
