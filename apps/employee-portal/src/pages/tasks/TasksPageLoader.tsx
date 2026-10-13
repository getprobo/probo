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

import { Suspense, useCallback, useEffect, useState } from "react";
import { useQueryLoader } from "react-relay";

import type { TasksPageQuery } from "#/__generated__/core/TasksPageQuery.graphql";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

import {
  pickTaskListIncludeFlags,
  taskListIncludeFlagsEqual,
} from "./_lib/taskListIncludes";
import { useTasksFilters } from "./_lib/useTasksFilters";
import { TasksPage, tasksPageQuery } from "./TasksPage";
import { TasksPageSkeleton } from "./TasksPageSkeleton";

export default function TasksPageLoader() {
  const { organizationId } = useEmployeePortalRoute();

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
  const { graphqlQuery, includeFlags } = useTasksFilters();
  const [queryRef, loadQuery] = useQueryLoader<TasksPageQuery>(tasksPageQuery);
  const [pageReady, setPageReady] = useState(false);
  const onReady = useCallback(() => {
    setPageReady(true);
  }, []);

  useEffect(() => {
    if (pageReady) {
      return;
    }
    loadQuery(
      { organizationId, query: graphqlQuery, ...includeFlags },
      { fetchPolicy: "store-and-network" },
    );
  }, [graphqlQuery, includeFlags, loadQuery, organizationId, pageReady]);

  const queryMatchesPendingFilter = queryRef != null
    && queryRef.variables.query === graphqlQuery
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
    return <TasksPageSkeleton />;
  }

  return (
    <Suspense fallback={<TasksPageSkeleton />}>
      <TasksPage queryRef={currentQueryRef} onReady={onReady} />
    </Suspense>
  );
}
