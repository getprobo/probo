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

import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge } from "@probo/ui";
import { useLazyLoadQuery } from "react-relay";
import { useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { MeasureTasksTabQuery } from "#/__generated__/core/MeasureTasksTabQuery.graphql";
import TaskFormDialog from "#/components/tasks/TaskFormDialog";
import { TasksCard } from "#/components/tasks/TasksCard";

const tasksQuery = graphql`
  query MeasureTasksTabQuery($measureId: ID!) {
    node(id: $measureId) @required(action: THROW) {
      __typename
      ... on Measure {
        canCreateTask: permission(action: "core:task:create")
        tasks(first: 100, orderBy: { field: PRIORITY_RANK, direction: ASC })
          @connection(key: "Measure__tasks")
          @required(action: THROW) {
          __id
          edges @required(action: THROW) {
            node {
              ...TasksCard_task
              ...TaskFormDialogFragment
              ...TasksCard_TaskRowFragment
            }
          }
        }
      }
    }
  }
`;

export default function MeasureTasksTab() {
  const { __ } = useTranslate();
  const { measureId } = useParams<{ measureId: string }>();
  if (!measureId) {
    throw new Error("Missing :measureId param in route");
  }
  const { node } = useLazyLoadQuery<MeasureTasksTabQuery>(tasksQuery, { measureId });
  if (node.__typename !== "Measure") {
    throw new Error("invalid node type");
  }
  const connectionId = node.tasks.__id;

  return (
    <div className="relative">
      <TasksCard connectionId={connectionId} tasks={node.tasks.edges} />
      {node.canCreateTask && (
        <TaskFormDialog connection={connectionId} measureId={measureId}>
          <Button
            variant="secondary"
            icon={IconPlusLarge}
            className="absolute top-3 right-6"
          >
            {__("New task")}
          </Button>
        </TaskFormDialog>
      )}
    </div>
  );
}
