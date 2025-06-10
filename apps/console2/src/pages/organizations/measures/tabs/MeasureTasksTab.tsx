import { graphql } from "relay-runtime";
import type { MeasureTasksTabFragment$key } from "./__generated__/MeasureTasksTabFragment.graphql";
import { useOutletContext } from "react-router";
import { useFragment } from "react-relay";
import TasksCard from "/components/tasks/TasksCard";
import { Button, IconPlusLarge } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import TaskFormDialog from "/components/tasks/TaskFormDialog";

export const tasksFragment = graphql`
  fragment MeasureTasksTabFragment on Measure {
    tasks(first: 100) @connection(key: "Measure__tasks") {
      __id
      edges {
        node {
          id
          name
          state
          description
          ...TaskFormDialogFragment
          assignedTo {
            id
            fullName
          }
        }
      }
    }
  }
`;

export default function MeasureTasksTab() {
  const { __ } = useTranslate();
  const { measure } = useOutletContext<{
    measure: MeasureTasksTabFragment$key & { id: string };
  }>();
  const data = useFragment(tasksFragment, measure);
  const connectionId = data.tasks.__id;
  const tasks = data.tasks?.edges?.map((edge) => edge.node) ?? [];

  return (
    <div className="relative">
      <TasksCard connectionId={connectionId} tasks={tasks} />
      <TaskFormDialog connection={connectionId} measureId={measure.id}>
        <Button
          variant="secondary"
          icon={IconPlusLarge}
          className="absolute top-3 right-6"
        >
          {__("New task")}
        </Button>
      </TaskFormDialog>
    </div>
  );
}
