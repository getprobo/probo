import { Button, IconPlusLarge, PageHeader } from "@probo/ui";
import {
  usePreloadedQuery,
  useRefetchableFragment,
  type PreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";
import type { TaskGraphQuery } from "/__generated__/core/TaskGraphQuery.graphql";
import { useTranslate } from "@probo/i18n";
import type { TasksPageFragment$key } from "/__generated__/core/TasksPageFragment.graphql";
import { tasksQuery } from "/hooks/graph/TaskGraph";
import { usePageTitle } from "@probo/hooks";
import TasksCard from "/components/tasks/TasksCard";
import TaskFormDialog from "/components/tasks/TaskFormDialog";
import { PermissionsContext } from "/providers/PermissionsContext";
import { use } from "react";

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
    tasks(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "TasksPageFragment_tasks") {
      __id
      edges {
        node {
          id
          name
          state
          description
          timeEstimate
          deadline
          ...TaskFormDialogFragment
          measure {
            id
            name
          }
          assignedTo {
            id
            fullName
          }
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
  const tasks = data.tasks?.edges.map((edge) => edge.node);
  const connectionId = data.tasks.__id;
  const { isAuthorized } = use(PermissionsContext);
  usePageTitle(__("Tasks"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Tasks")}
        description={__(
          "Track your assigned compliance tasks and keep progress on track.",
        )}
      >
        {isAuthorized("Organization", "createTask") && (
          <TaskFormDialog connection={connectionId}>
            <Button icon={IconPlusLarge}>{__("New task")}</Button>
          </TaskFormDialog>
        )}
      </PageHeader>
      <TasksCard connectionId={connectionId} tasks={tasks ?? []} />
    </div>
  );
}
