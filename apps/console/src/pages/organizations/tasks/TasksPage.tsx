import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, IconPlusLarge, PageHeader } from "@probo/ui";
import {
  type PreloadedQuery,
  usePreloadedQuery,
  useRefetchableFragment,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { TaskGraphQuery } from "#/__generated__/core/TaskGraphQuery.graphql";
import type { TasksPageFragment$key } from "#/__generated__/core/TasksPageFragment.graphql";
import TaskFormDialog from "#/components/tasks/TaskFormDialog";
import { TasksCard } from "#/components/tasks/TasksCard";
import { tasksQuery } from "#/hooks/graph/TaskGraph";

const tasksFragment = graphql`
  fragment TasksPageFragment on Organization
  @refetchable(queryName: "TasksPageFragment_query")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 500 }
    order: { type: "TaskOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    canCreateTask: permission(action: "core:task:create")
    tasks(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "TasksPageFragment_tasks") @required(action: THROW) {
      __id
      edges @required(action: THROW) {
        node {
          id
          state
          ...TaskFormDialogFragment
          ...TasksCard_TaskRowFragment
        }
      }
    }
  }
`;

interface Props {
  queryRef: PreloadedQuery<TaskGraphQuery>;
}

export default function TasksPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const query = usePreloadedQuery(tasksQuery, queryRef);
  const [data] = useRefetchableFragment(
    tasksFragment,
    query.organization as TasksPageFragment$key,
  );
  const connectionId = data.tasks.__id;
  usePageTitle(__("Tasks"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Tasks")}
        description={__(
          "Track your assigned compliance tasks and keep progress on track.",
        )}
      >
        {data.canCreateTask && (
          <TaskFormDialog connection={connectionId}>
            <Button icon={IconPlusLarge}>{__("New task")}</Button>
          </TaskFormDialog>
        )}
      </PageHeader>
      <TasksCard connectionId={connectionId} tasks={data.tasks.edges} />
    </div>
  );
}
