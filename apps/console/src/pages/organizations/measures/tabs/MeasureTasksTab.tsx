// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
import { Button, IconPlusLarge } from "@probo/ui";
import { useLazyLoadQuery } from "react-relay";
import { useOutletContext } from "react-router";
import { graphql } from "relay-runtime";

import type { MeasureTasksTabQuery } from "#/__generated__/core/MeasureTasksTabQuery.graphql";
import TaskFormDialog from "#/components/tasks/TaskFormDialog";
import { TasksCard } from "#/components/tasks/TasksCard";

const tasksQuery = graphql`
  query MeasureTasksTabQuery($measureId: ID!) {
    node(id: $measureId) @required(action: THROW) {
      __typename
      ... on Measure {
        id
        canCreateTask: permission(action: "core:task:create")
        tasks(first: 100, orderBy: { field: PRIORITY, direction: ASC })
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
  const { measure } = useOutletContext<{
    measure: { id: string };
  }>();
  const { node } = useLazyLoadQuery<MeasureTasksTabQuery>(tasksQuery, {
    measureId: measure.id,
  });
  if (node.__typename !== "Measure") {
    throw new Error("invalid node type");
  }
  const connectionId = node.tasks.__id;

  return (
    <div className="relative">
      <TasksCard connectionId={connectionId} tasks={node.tasks.edges} />
      {node.canCreateTask && (
        <TaskFormDialog connection={connectionId} measureId={measure.id}>
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
