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

import { Suspense, use, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useQueryLoader } from "react-relay";
import { useSearchParams } from "react-router";

import type { TasksPageQuery } from "#/__generated__/core/TasksPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

import {
  readTaskListFilterPreference,
  restoredTaskListFilters,
  taskListFilterPreferenceFromParams,
  taskListFilterPreferenceInParams,
  writeTaskListFilterPreference,
} from "./_lib/taskListFilterPreference";
import {
  pickTaskListIncludeFlags,
  taskListIncludeFlags,
  taskListIncludeFlagsEqual,
} from "./_lib/taskListIncludes";
import { useTasksCardFilters } from "./_lib/useTasksCardFilters";
import { TasksPage, tasksPageQuery } from "./TasksPage";

export default function TasksPageLoader() {
  const organizationId = useOrganizationId();

  return (
    <TasksPageQueryLoader
      key={organizationId}
      organizationId={organizationId}
    />
  );
}

interface TasksPageQueryLoaderProps {
  organizationId: string;
}

function TasksPageQueryLoader({ organizationId }: TasksPageQueryLoaderProps) {
  const { profile } = use(CurrentUser);
  const profileId = profile?.id ?? "";
  const [searchParams, setSearchParams] = useSearchParams();
  const filtersInUrl = taskListFilterPreferenceInParams(searchParams);
  const restored = filtersInUrl
    ? null
    : restoredTaskListFilters(
        searchParams,
        readTaskListFilterPreference(organizationId, profileId),
        profileId,
      );
  const restoredSearch = restored?.params.toString() ?? null;
  const nextSearch = restoredSearch == null || restoredSearch === searchParams.toString()
    ? null
    : restoredSearch;
  // Copy saved filters into a bare URL before the first query, so the list
  // does not load unfiltered and then replace itself.
  const filtersReady = nextSearch == null;
  const appliedSearch = useRef<string | null>(null);
  const { graphqlFilter } = useTasksCardFilters();
  const includeFlags = useMemo(
    () => taskListIncludeFlags(graphqlFilter.state),
    [graphqlFilter.state],
  );
  const [queryRef, loadQuery]
    = useQueryLoader<TasksPageQuery>(tasksPageQuery);
  const [pageReady, setPageReady] = useState(false);
  const onReady = useCallback(() => {
    setPageReady(true);
  }, []);

  useEffect(() => {
    if (nextSearch == null || appliedSearch.current === nextSearch) {
      return;
    }
    appliedSearch.current = nextSearch;
    if (readTaskListFilterPreference(organizationId, profileId) == null) {
      writeTaskListFilterPreference(
        organizationId,
        profileId,
        taskListFilterPreferenceFromParams(new URLSearchParams(nextSearch)),
      );
    }
    setSearchParams(new URLSearchParams(nextSearch), { replace: true });
  }, [nextSearch, organizationId, profileId, setSearchParams]);

  useEffect(() => {
    if (!filtersReady || pageReady) {
      return;
    }
    loadQuery({
      organizationId,
      query: graphqlFilter.query,
      assignedToId: graphqlFilter.assignedToId,
      internalControlId: graphqlFilter.internalControlId,
      deadlineOnOrBefore: graphqlFilter.deadlineOnOrBefore,
      deadlineOnOrAfter: graphqlFilter.deadlineOnOrAfter,
      hasDeadline: graphqlFilter.hasDeadline,
      hasAssignee: graphqlFilter.hasAssignee,
      ...includeFlags,
    });
  }, [filtersReady, graphqlFilter, includeFlags, loadQuery, organizationId, pageReady]);

  const queryMatchesPendingFilter = queryRef != null
    && queryRef.variables.query === graphqlFilter.query
    && queryRef.variables.assignedToId === graphqlFilter.assignedToId
    && queryRef.variables.internalControlId === graphqlFilter.internalControlId
    && queryRef.variables.deadlineOnOrBefore === graphqlFilter.deadlineOnOrBefore
    && queryRef.variables.deadlineOnOrAfter === graphqlFilter.deadlineOnOrAfter
    && queryRef.variables.hasDeadline === graphqlFilter.hasDeadline
    && queryRef.variables.hasAssignee === graphqlFilter.hasAssignee
    && taskListIncludeFlagsEqual(
      pickTaskListIncludeFlags(queryRef.variables),
      includeFlags,
    );
  const currentQueryRef = queryRef != null
    && queryRef.variables.organizationId === organizationId
    && (pageReady || queryMatchesPendingFilter)
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <TasksPage queryRef={currentQueryRef} onReady={onReady} />
    </Suspense>
  );
}
