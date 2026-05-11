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
  const { __ } = useTranslate();
  const data = usePreloadedQuery(cookieBannerResourcesPageQuery, queryRef);

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
          placeholder={__("Search by origin or path...")}
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
          <Option value="ALL">{__("All types")}</Option>
          <Option value="SCRIPT">{__("Script")}</Option>
          <Option value="IFRAME">{__("Iframe")}</Option>
          <Option value="IMAGE">{__("Image")}</Option>
          <Option value="STYLESHEET">{__("Stylesheet")}</Option>
          <Option value="FONT">{__("Font")}</Option>
          <Option value="BEACON">{__("Beacon")}</Option>
          <Option value="FETCH">{__("Fetch")}</Option>
          <Option value="MEDIA">{__("Media")}</Option>
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
                    <Th>{__("Type")}</Th>
                    <SortableTh field="ORIGIN">{__("Origin")}</SortableTh>
                    <Th>{__("Path")}</Th>
                    <SortableTh field="LAST_DETECTED_AT">{__("Last Detected")}</SortableTh>
                    <Th className="w-28" />
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
                    {__("No uncategorised resources")}
                  </h3>
                  <p className="text-txt-tertiary">
                    {__("All detected resources have been categorised. New resources will appear here when detected.")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}
