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

import { MagnifyingGlassIcon } from "@phosphor-icons/react";
import {
  taskListGroupStates,
  TASKS_NAMESPACE,
  type TaskState,
  TaskStateIcon,
  taskStateKeys,
} from "@probo/tasks";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
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
import { tasksPage } from "./variants";

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
  const { t: tTasks } = useTranslation(TASKS_NAMESPACE);
  const { state: selectedState, setState } = useTasksFilters();
  const [queryInput, setQueryInput] = useTasksSearch();
  const { viewer } = usePreloadedQuery<TasksPageQuery>(
    tasksPageQuery,
    queryRef,
  );
  const slots = tasksPage();

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
        <div className={slots.tools()}>
          <div className={slots.search()}>
            <TextField
              icon={<MagnifyingGlassIcon />}
              value={queryInput}
              onValueChange={setQueryInput}
              placeholder={t("searchPlaceholder")}
              aria-label={t("searchPlaceholder")}
            />
          </div>
          <div className={slots.filters()}>
            <div className={slots.filter()}>
              <Select
                value={selectedState}
                onValueChange={setState}
              >
                <SelectTrigger
                  size={2}
                  placeholder={t("filters.allStatuses")}
                  aria-label={t("filters.status")}
                >
                  {(value: TaskState | null) => value == null
                    ? t("filters.allStatuses")
                    : (
                        <span className={slots.stateOption()}>
                          <TaskStateIcon state={value} />
                          {tTasks(`states.${taskStateKeys[value]}`)}
                        </span>
                      )}
                </SelectTrigger>
                <SelectPopup align="start">
                  <SelectItem value={null}>
                    {t("filters.allStatuses")}
                  </SelectItem>
                  {taskListGroupStates.map(state => (
                    <SelectItem key={state} value={state}>
                      <span className={slots.stateOption()}>
                        <TaskStateIcon state={state} />
                        {tTasks(`states.${taskStateKeys[state]}`)}
                      </span>
                    </SelectItem>
                  ))}
                </SelectPopup>
              </Select>
            </div>
          </div>
        </div>
        <Suspense fallback={<TasksPageListSkeleton />}>
          <TasksPageList viewerKey={viewer} />
        </Suspense>
      </div>
    </main>
  );
}
