import { graphql } from "relay-runtime";
import type { MeasureTasksTabQuery } from "/__generated__/core/MeasureTasksTabQuery.graphql";
import { useOutletContext } from "react-router";
import { useLazyLoadQuery } from "react-relay";
import TasksCard from "/components/tasks/TasksCard";
import { Button, IconPlusLarge } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import TaskFormDialog from "/components/tasks/TaskFormDialog";

const tasksQuery = graphql`
  query MeasureTasksTabQuery($measureId: ID!) {
    node(id: $measureId) {
      ... on Measure {
        id
        tasks(first: 100) @connection(key: "Measure__tasks") {
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
              assignedTo {
                id
                fullName
              }
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
  const data = useLazyLoadQuery<MeasureTasksTabQuery>(tasksQuery, {
    measureId: measure.id,
  });
  const node = data.node;
  if (!node || !node.tasks) {
    return null;
  }
  const connectionId = node.tasks.__id;
  const tasks = node.tasks.edges?.map((edge) => edge.node) ?? [];

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
