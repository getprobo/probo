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

import { TaskListToolbar } from "@probo/tasks";
import { Suspense, useEffect } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";

import type { TasksPageQuery } from "#/__generated__/core/TasksPageQuery.graphql";
import { PageHeader } from "#/pages/_components/PageHeader";

import { TasksPageList } from "./_components/TasksPageList";
import { TasksPageListSkeleton } from "./_components/TasksPageListSkeleton";
import { useTasksFilters } from "./_lib/useTasksFilters";
import { useTasksSearch } from "./_lib/useTasksSearch";

export const tasksPageQuery = graphql`
  query TasksPageQuery(
    $organizationId: ID!
    $query: String
    $includeInProgress: Boolean!
    $includeTodo: Boolean!
    $includeBacklog: Boolean!
    $includeDone: Boolean!
    $includeCanceled: Boolean!
    $includeDuplicate: Boolean!
  )
  @throwOnFieldError {
    viewer @required(action: THROW) {
      ...TasksPageList_viewer
        @arguments(
          organizationId: $organizationId
          query: $query
          includeInProgress: $includeInProgress
          includeTodo: $includeTodo
          includeBacklog: $includeBacklog
          includeDone: $includeDone
          includeCanceled: $includeCanceled
          includeDuplicate: $includeDuplicate
        )
    }
  }
`;

interface TasksPageProps {
  queryRef: PreloadedQuery<TasksPageQuery>;
  onReady?: () => void;
}

export function TasksPage({ queryRef, onReady }: TasksPageProps) {
  const { t } = useTranslation("tasks");
  const { t: tApp } = useTranslation();
  const { state: selectedState, setState } = useTasksFilters();
  const [queryInput, setQueryInput] = useTasksSearch();
  const { viewer } = usePreloadedQuery<TasksPageQuery>(
    tasksPageQuery,
    queryRef,
  );
  useEffect(() => {
    onReady?.();
  }, [onReady]);

  return (
    <main className="mx-auto flex w-full max-w-5xl flex-col gap-10 px-8 pt-8 pb-32">
      <PageHeader
        homeLabel={tApp("homePage.breadcrumb")}
        currentLabel={t("breadcrumb")}
        title={t("title")}
      />
      <div className="flex flex-col gap-4">
        <TaskListToolbar
          query={queryInput}
          onQueryChange={setQueryInput}
          state={selectedState}
          onStateChange={setState}
        />
        <Suspense fallback={<TasksPageListSkeleton />}>
          <TasksPageList viewerKey={viewer} />
        </Suspense>
      </div>
    </main>
  );
}
