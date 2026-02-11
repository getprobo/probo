import { formatDate, formatDuration, promisifyMutation } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Card,
  DropdownItem,
  IconArrowCornerDownLeft,
  IconPencil,
  IconTrashCan,
  PriorityLevel,
  Spinner,
  TabBadge,
  TabItem,
  Tabs,
  TaskStateIcon,
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { Fragment } from "react";
import {
  graphql,
  useFragment,
  useMutation,
  useRelayEnvironment,
} from "react-relay";
import { Link, useLocation, useParams } from "react-router";

import type { MeasureTasksTabQuery$data } from "#/__generated__/core/MeasureTasksTabQuery.graphql";
import type { TaskFormDialogFragment$key } from "#/__generated__/core/TaskFormDialogFragment.graphql";
import type { TasksCard_TaskRowFragment$key } from "#/__generated__/core/TasksCard_TaskRowFragment.graphql";
import type { TasksPageFragment$data } from "#/__generated__/core/TasksPageFragment.graphql";
import TaskFormDialog, {
  taskUpdateMutation,
} from "#/components/tasks/TaskFormDialog";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useOrganizationId } from "#/hooks/useOrganizationId";

type Props = {
  tasks:
    | TasksPageFragment$data["tasks"]["edges"]
    | Extract<
      MeasureTasksTabQuery$data["node"],
      { __typename: "Measure" }
    >["tasks"]["edges"];
  connectionId: string;
};

export function TasksCard({ tasks, connectionId }: Props) {
  const { __ } = useTranslate();
  const hash = useLocation().hash.replace("#", "");

  const hashes = [
    { hash: "", label: __("To do"), state: "TODO" },
    { hash: "done", label: __("Done"), state: "DONE" },
    { hash: "all", label: __("All"), state: null },
  ] as const;

  const tasksPerHash = new Map([
    ["", tasks?.filter(({ node }) => node.state === "TODO")],
    ["done", tasks?.filter(({ node }) => node.state === "DONE")],
    ["all", tasks],
  ]);

  const filteredTasks = tasksPerHash.get(hash) ?? [];

  usePageTitle(__("Tasks"));

  return (
    <div className="space-y-6">
      {tasks.length === 0
        ? (
            <p className="text-center py-6 text-txt-secondary">{__("No tasks")}</p>
          )
        : (
            <Card>
              <Tabs className="px-6">
                {hashes.map(h => (
                  <TabItem asChild active={hash === h.hash} key={h.hash}>
                    <Link to={`#${h.hash}`}>
                      {h.label}
                      <TabBadge>{tasksPerHash.get(h.hash)?.length}</TabBadge>
                    </Link>
                  </TabItem>
                ))}
              </Tabs>
              <div className="divide-y divide-border-solid">
                {hash === "all"
                  // All tabs group the todo using the state
                  ? hashes
                      .slice(0, 2)
                      .filter(h => tasksPerHash.get(h.hash)?.length)
                      .map(h => (
                        <Fragment key={h.label}>
                          <h2 className="px-6 py-3 text-sm font-medium flex items-center gap-2 bg-subtle">
                            <TaskStateIcon state={h.state!} />
                            {h.label}
                          </h2>
                          {tasksPerHash.get(h.hash)?.map(({ node: task }) => (
                            <TaskRow
                              key={task.id}
                              fKey={task}
                              connectionId={connectionId}
                            />
                          ))}
                        </Fragment>
                      ))
                  // Todo and Done tab simply list todos
                  : filteredTasks?.map(({ node: task }) => (
                      <TaskRow
                        key={task.id}
                        fKey={task}
                        connectionId={connectionId}
                      />
                    ))}
              </div>
            </Card>
          )}
    </div>
  );
}

type TaskRowProps = {
  fKey: TasksCard_TaskRowFragment$key | TaskFormDialogFragment$key;
  connectionId: string;
};

const fragment = graphql`
  fragment TasksCard_TaskRowFragment on Task {
    id
    name
    state
    description
    timeEstimate
    deadline
    canUpdate: permission(action: "core:task:update")
    canDelete: permission(action: "core:task:delete")
    assignedTo {
      id
      fullName
    }
    measure {
      id
      name
    }
  }
`;

const deleteMutation = graphql`
  mutation TasksCardDeleteMutation(
    $input: DeleteTaskInput!
    $connections: [ID!]!
  ) {
    deleteTask(input: $input) {
      deletedTaskId @deleteEdge(connections: $connections)
    }
  }
`;

function TaskRow(props: TaskRowProps) {
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const [deleteTask] = useMutation(deleteMutation);
  const params = useParams<{ measureId?: string }>();

  const relayEnv = useRelayEnvironment();
  const { canUpdate, canDelete, ...task }
    = useFragment<TasksCard_TaskRowFragment$key>(
      fragment,
      props.fKey as TasksCard_TaskRowFragment$key,
    );
  const [updateTask, isUpdating] = useMutation(taskUpdateMutation);

  const onToggle = async () => {
    await promisifyMutation(updateTask)({
      variables: {
        input: {
          taskId: task.id,
          state: task.state === "TODO" ? "DONE" : "TODO",
        },
      },
    });
  };

  const onDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteTask)({
          variables: {
            input: { taskId: task.id },
            connections: [props.connectionId],
          },
          onCompleted: (_response, errors) => {
            if (!errors && params.measureId) {
              updateStoreCounter(
                relayEnv,
                params.measureId,
                "tasks(first:0)",
                -1,
              );
            }
          },
        }),
      {
        message: "Are you sure you want to delete this task?",
      },
    );
  };

  return (
    <>
      <TaskFormDialog
        task={props.fKey as TaskFormDialogFragment$key}
        ref={dialogRef}
      />
      <div className="flex items-center justify-between py-3 px-6">
        <div className="flex gap-2 items-start">
          <div className="flex items-center gap-2 pt-[2px]">
            <PriorityLevel level={1} />
            <button
              onClick={() => void onToggle()}
              className="cursor-pointer -m-1 p-1 disabled:opacity-60"
              disabled={isUpdating}
            >
              <TaskStateIcon state={task.state} />
            </button>
          </div>
          <div className="text-sm space-y-1 flex-1">
            <h2 className="font-medium">{task.name}</h2>
            {task.description && (
              <p className="text-txt-secondary whitespace-pre-wrap wrap-break-word">
                {task.description}
              </p>
            )}

            <div className="flex flex-wrap items-center gap-3 text-txt-secondary text-xs">
              {task.measure && (
                <span className="flex items-center gap-1">
                  <IconArrowCornerDownLeft className="scale-x-[-1]" size={14} />
                  <Link
                    className="hover:underline"
                    to={`/organizations/${organizationId}/measures/${task.measure?.id}`}
                  >
                    {task.measure?.name}
                  </Link>
                </span>
              )}
              {task.timeEstimate && (
                <span>{formatDuration(task.timeEstimate, __)}</span>
              )}
              {task.deadline && (
                <time dateTime={task.deadline}>
                  {formatDate(task.deadline)}
                </time>
              )}
            </div>
          </div>
        </div>
        {task.assignedTo?.fullName && (
          <div className="text-sm text-txt-secondary ml-auto mr-8">
            <Link
              className="hover:underline"
              to={`/organizations/${organizationId}/people/${task.assignedTo.id}`}
            >
              {task.assignedTo.fullName}
            </Link>
          </div>
        )}
        <div className="flex gap-2 items-center">
          {isUpdating && <Spinner size={16} />}
          {(canUpdate || canDelete) && (
            <ActionDropdown>
              {canUpdate && (
                <DropdownItem
                  icon={IconPencil}
                  onClick={() => dialogRef.current?.open()}
                >
                  {__("Edit")}
                </DropdownItem>
              )}
              {canDelete && (
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={onDelete}
                >
                  {__("Delete")}
                </DropdownItem>
              )}
            </ActionDropdown>
          )}
        </div>
      </div>
    </>
  );
}
