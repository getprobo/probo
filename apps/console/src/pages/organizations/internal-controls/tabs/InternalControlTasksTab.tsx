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

import type { InternalControlTasksTab_internalControl$key } from "#/__generated__/core/InternalControlTasksTab_internalControl.graphql";
import type { InternalControlTasksTabQuery } from "#/__generated__/core/InternalControlTasksTabQuery.graphql";
import type { InternalControlTasksTabRefetchQuery } from "#/__generated__/core/InternalControlTasksTabRefetchQuery.graphql";
import { TasksCard } from "#/components/tasks/TasksCard";
import { CreateTaskDialog } from "#/pages/organizations/tasks/_components/CreateTaskDialog";
import { useTasksCardFilters } from "#/pages/organizations/tasks/_lib/useTasksCardFilters";

import { internalControlTasksTab } from "./variants";

const internalControlTasksTabFragment = graphql`
  fragment InternalControlTasksTab_internalControl on InternalControl
  @refetchable(queryName: "InternalControlTasksTabRefetchQuery")
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
  query InternalControlTasksTabQuery($internalControlId: ID!, $filter: TaskFilter) {
    node(id: $internalControlId) @required(action: THROW) {
      __typename
      ... on InternalControl {
        ...InternalControlTasksTab_internalControl @arguments(filter: $filter)
      }
    }
  }
`;

interface InternalControlTasksPanelProps {
  internalControlKey: InternalControlTasksTab_internalControl$key;
  internalControlId: string;
}

function InternalControlTasksPanel({ internalControlKey, internalControlId }: InternalControlTasksPanelProps) {
  const { t } = useTranslation();
  const [internalControl, refetch] = useRefetchableFragment<
    InternalControlTasksTabRefetchQuery,
    InternalControlTasksTab_internalControl$key
  >(internalControlTasksTabFragment, internalControlKey);
  const connectionId = internalControl.tasks.__id;
  const { root, create } = internalControlTasksTab();

  return (
    <div className={root()}>
      {internalControl.canCreateTask && (
        <CreateTaskDialog connectionId={connectionId} internalControlId={internalControlId}>
          <Button
            variant="surface"
            iconStart={<PlusIcon />}
            className={create()}
          >
            {t("internalControlTasksTab.actions.newTask")}
          </Button>
        </CreateTaskDialog>
      )}
      <TasksCard
        tasks={internalControl.tasks.edges}
        refetch={refetch}
      />
    </div>
  );
}

export default function InternalControlTasksTab() {
  const { internalControlId } = useParams<{ internalControlId: string }>();
  if (!internalControlId) {
    throw new Error("Missing :internalControlId param in route");
  }

  return <InternalControlTasksQuery key={internalControlId} internalControlId={internalControlId} />;
}

interface InternalControlTasksQueryProps {
  internalControlId: string;
}

function InternalControlTasksQuery({ internalControlId }: InternalControlTasksQueryProps) {
  const { graphqlFilter } = useTasksCardFilters();
  const [queryVariables] = useState({ internalControlId, filter: graphqlFilter });
  const { node } = useLazyLoadQuery<InternalControlTasksTabQuery>(
    tasksQuery,
    queryVariables,
  );
  if (node.__typename !== "InternalControl") {
    throw new Error("invalid node type");
  }

  return <InternalControlTasksPanel internalControlKey={node} internalControlId={internalControlId} />;
}
