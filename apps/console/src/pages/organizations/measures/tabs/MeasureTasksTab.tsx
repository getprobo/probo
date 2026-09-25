// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { PlusIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useLazyLoadQuery, useRefetchableFragment } from "react-relay";
import { useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { MeasureTasksTab_measure$key } from "#/__generated__/core/MeasureTasksTab_measure.graphql";
import type { MeasureTasksTabQuery } from "#/__generated__/core/MeasureTasksTabQuery.graphql";
import type { MeasureTasksTabRefetchQuery } from "#/__generated__/core/MeasureTasksTabRefetchQuery.graphql";
import { TasksCard } from "#/components/tasks/TasksCard";
import { CreateTaskDialog } from "#/pages/organizations/tasks/_components/CreateTaskDialog";
import { useTasksCardFilters } from "#/pages/organizations/tasks/_lib/useTasksCardFilters";

import { measureTasksTab } from "./variants";

const measureTasksTabFragment = graphql`
  fragment MeasureTasksTab_measure on Measure
  @refetchable(queryName: "MeasureTasksTabRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 100 }
    after: { type: "CursorKey", defaultValue: null }
    filter: { type: "TaskFilter", defaultValue: null }
  ) {
    canCreateTask: permission(action: "core:task:create")
    tasks(
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: $filter
    ) @connection(key: "Measure__tasks") @required(action: THROW) {
      __id
      edges @required(action: THROW) {
        node {
          ...TasksCard_task
          ...TaskListItem_task
        }
      }
    }
  }
`;

const tasksQuery = graphql`
  query MeasureTasksTabQuery($measureId: ID!, $filter: TaskFilter) {
    node(id: $measureId) @required(action: THROW) {
      __typename
      ... on Measure {
        ...MeasureTasksTab_measure @arguments(filter: $filter)
      }
    }
  }
`;

interface MeasureTasksPanelProps {
  measureKey: MeasureTasksTab_measure$key;
  measureId: string;
}

function MeasureTasksPanel({ measureKey, measureId }: MeasureTasksPanelProps) {
  const { t } = useTranslation();
  const [measure, refetch] = useRefetchableFragment<
    MeasureTasksTabRefetchQuery,
    MeasureTasksTab_measure$key
  >(measureTasksTabFragment, measureKey);
  const connectionId = measure.tasks.__id;
  const { root, create } = measureTasksTab();

  return (
    <div className={root()}>
      {measure.canCreateTask && (
        <CreateTaskDialog connectionId={connectionId} measureId={measureId}>
          <Button
            variant="surface"
            iconStart={<PlusIcon />}
            className={create()}
          >
            {t("measureTasksTab.actions.newTask")}
          </Button>
        </CreateTaskDialog>
      )}
      <TasksCard
        connectionId={connectionId}
        tasks={measure.tasks.edges}
        refetch={refetch}
      />
    </div>
  );
}

export default function MeasureTasksTab() {
  const { measureId } = useParams<{ measureId: string }>();
  if (!measureId) {
    throw new Error("Missing :measureId param in route");
  }

  return <MeasureTasksQuery key={measureId} measureId={measureId} />;
}

interface MeasureTasksQueryProps {
  measureId: string;
}

function MeasureTasksQuery({ measureId }: MeasureTasksQueryProps) {
  const { graphqlFilter } = useTasksCardFilters();
  const [queryVariables] = useState({ measureId, filter: graphqlFilter });
  const { node } = useLazyLoadQuery<MeasureTasksTabQuery>(
    tasksQuery,
    queryVariables,
  );
  if (node.__typename !== "Measure") {
    throw new Error("invalid node type");
  }

  return <MeasureTasksPanel measureKey={node} measureId={measureId} />;
}
