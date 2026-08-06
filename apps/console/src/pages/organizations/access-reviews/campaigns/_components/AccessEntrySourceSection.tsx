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

import { InfiniteScrollTrigger } from "@probo/ui";
import { startTransition, useCallback, useEffect, useMemo, useRef } from "react";
import { graphql, usePaginationFragment } from "react-relay";
import { readInlineData } from "relay-runtime";

import type { AccessEntrySourceSection_entry$key } from "#/__generated__/core/AccessEntrySourceSection_entry.graphql";
import type { AccessEntrySourceSection_source$key } from "#/__generated__/core/AccessEntrySourceSection_source.graphql";
import type { AccessEntrySourceSectionPaginationQuery } from "#/__generated__/core/AccessEntrySourceSectionPaginationQuery.graphql";

import { AccessEntrySection } from "./AccessEntrySection";
import { AccessEntrySectionList } from "./AccessEntrySectionList";

const PAGE_SIZE = 100;
const EMPTY_IDS: ReadonlyArray<string> = [];

const accessEntrySourceSectionFragment = graphql`
  fragment AccessEntrySourceSection_source on AccessReviewCampaignSource
  @refetchable(queryName: "AccessEntrySourceSectionPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 100 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    id
    fetchAttempts(first: 1) {
      edges {
        node {
          id
          status
        }
      }
    }
    entries(first: $first, after: $after)
      @connection(key: "AccessEntrySourceSection_entries", filters: []) {
      totalCount
      edges {
        node {
          id
          ...AccessEntrySourceSection_entry
          # eslint-disable-next-line relay/must-colocate-fragment-spreads
          ...AccessEntrySectionList_entry
        }
      }
    }
    ...AccessEntrySection_source
  }
`;

const accessEntrySourceSectionEntryFragment = graphql`
  fragment AccessEntrySourceSection_entry on AccessReviewEntry @inline {
    email
    fullName
    isAdmin
    mfaStatus
    authMethod
  }
`;

export type EntryFilters = {
  email: string;
  connectorIds: ReadonlyArray<string>;
  mfa: ReadonlyArray<string>;
  authMethod: ReadonlyArray<string>;
  admin: ReadonlyArray<string>;
};

// What the section contributes to the page-wide selection: the entries it
// currently shows, plus whether more pages could still add to them.
export type SourceMatches = {
  entryIds: ReadonlyArray<string>;
  hasNext: boolean;
};

function entryMatchesFilters(
  entryKey: AccessEntrySourceSection_entry$key,
  filters: EntryFilters,
): boolean {
  const entry = readInlineData(accessEntrySourceSectionEntryFragment, entryKey);

  const query = filters.email;
  if (query) {
    const email = entry.email.toLowerCase();
    const fullName = entry.fullName.toLowerCase();
    if (!email.includes(query) && !fullName.includes(query)) {
      return false;
    }
  }

  if (filters.mfa.length > 0 && !filters.mfa.includes(entry.mfaStatus)) {
    return false;
  }

  if (
    filters.authMethod.length > 0
    && !filters.authMethod.includes(entry.authMethod)
  ) {
    return false;
  }

  if (filters.admin.length > 0) {
    const adminValue = entry.isAdmin ? "YES" : "NO";
    if (!filters.admin.includes(adminValue)) {
      return false;
    }
  }

  return true;
}

interface AccessEntrySourceSectionProps {
  sourceKey: AccessEntrySourceSection_source$key;
  filters: EntryFilters;
  isPendingActions: boolean;
  selectedIds: ReadonlySet<string>;
  onSelectedChange: (entryId: string, event: { shiftKey: boolean }) => void;
  onMatchesChange: (sourceId: string, matches: SourceMatches) => void;
}

// Owns one connector's entries connection: paginates it, applies the toolbar
// filters to the pages loaded so far, and reports the surviving entries up so
// the page can drive select-all and range selection across every connector.
export function AccessEntrySourceSection({
  sourceKey,
  filters,
  isPendingActions,
  selectedIds,
  onSelectedChange,
  onMatchesChange,
}: AccessEntrySourceSectionProps) {
  const {
    data: source,
    loadNext,
    hasNext,
    isLoadingNext,
    refetch,
  } = usePaginationFragment<
    AccessEntrySourceSectionPaginationQuery,
    AccessEntrySourceSection_source$key
  >(accessEntrySourceSectionFragment, sourceKey);

  const sourceId = source.id;
  const latestAttempt = source.fetchAttempts.edges[0]?.node ?? null;
  const fetchStatus = latestAttempt?.status ?? null;
  const isFetchRunning = fetchStatus === "QUEUED" || fetchStatus === "FETCHING";
  const isExcluded = filters.connectorIds.length > 0
    && !filters.connectorIds.includes(sourceId);
  // The connector filter picks whole sections, so only the per-entry filters
  // make the loaded count diverge from the connector's real total.
  const hasEntryFilters = filters.email !== ""
    || filters.mfa.length > 0
    || filters.authMethod.length > 0
    || filters.admin.length > 0;

  const matchedEntries = useMemo(
    () => source.entries.edges
      .map(edge => edge.node)
      .filter(node => entryMatchesFilters(node, filters)),
    [source.entries.edges, filters],
  );
  const matchedIds = useMemo(
    () => matchedEntries.map(entry => entry.id),
    [matchedEntries],
  );

  // An excluded connector contributes nothing and must not keep paginating in
  // the background looking for matches it would never show.
  const contributedIds = isExcluded ? EMPTY_IDS : matchedIds;
  const isPending = !isExcluded && hasNext;

  useEffect(() => {
    onMatchesChange(sourceId, { entryIds: contributedIds, hasNext: isPending });
  }, [sourceId, contributedIds, isPending, onMatchesChange]);

  // Reloading only the first page would drop the pages a reviewer already
  // scrolled through, so the refresh asks for every entry loaded so far.
  const loadedCount = source.entries.edges.length;
  const loadedCountRef = useRef(loadedCount);
  useEffect(() => {
    loadedCountRef.current = loadedCount;
  }, [loadedCount]);

  // A connector writes its entries when its fetch attempt commits, so the
  // section reloads its own connection once the attempt settles. Polling
  // samples statuses every few seconds and can miss the QUEUED and FETCHING
  // states of a quick attempt, and a retry settles on the same status as the
  // attempt it replaces, so the reload keys off the attempt identity instead of
  // off a status transition. The campaign poll only refreshes statuses: it must
  // not replace paginated connections.
  const attemptKey = latestAttempt
    ? `${latestAttempt.id}:${latestAttempt.status}`
    : "";
  const isAttemptSettled = fetchStatus === "SUCCESS" || fetchStatus === "FAILED";
  const previousAttemptKeyRef = useRef(attemptKey);
  useEffect(() => {
    const previous = previousAttemptKeyRef.current;
    previousAttemptKeyRef.current = attemptKey;
    if (previous === attemptKey || !isAttemptSettled) {
      return;
    }
    // A transition keeps the section on screen instead of suspending it while
    // the refreshed pages are in flight.
    startTransition(() => {
      refetch(
        { first: Math.max(loadedCountRef.current, PAGE_SIZE) },
        { fetchPolicy: "network-only" },
      );
    });
  }, [attemptKey, isAttemptSettled, refetch]);

  const handleView = useCallback(() => {
    loadNext(PAGE_SIZE);
  }, [loadNext]);

  if (isExcluded) {
    return null;
  }

  // Filters run over loaded pages only, so an empty section still means "keep
  // looking" until the connection is exhausted. A source that is still fetching
  // or that failed keeps its section too, so filters cannot hide its progress
  // or its error.
  if (
    hasEntryFilters
    && matchedEntries.length === 0
    && !hasNext
    && !isFetchRunning
    && fetchStatus !== "FAILED"
  ) {
    return null;
  }

  return (
    <AccessEntrySection
      sourceKey={source}
      count={hasEntryFilters ? matchedEntries.length : source.entries.totalCount}
      footer={hasNext
        ? <InfiniteScrollTrigger onView={handleView} loading={isLoadingNext} />
        : null}
    >
      <AccessEntrySectionList
        entryKeys={matchedEntries}
        isPendingActions={isPendingActions}
        selectedIds={selectedIds}
        onSelectedChange={onSelectedChange}
      />
    </AccessEntrySection>
  );
}
