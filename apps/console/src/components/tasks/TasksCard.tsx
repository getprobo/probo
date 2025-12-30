import { promisifyMutation, formatDate, formatDuration } from "@probo/helpers";
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
import { Fragment, use } from "react";
import { graphql, useMutation, useRelayEnvironment } from "react-relay";
import { Link, useLocation, useParams } from "react-router";
import type { TaskFormDialogFragment$key } from "/__generated__/core/TaskFormDialogFragment.graphql";
import TaskFormDialog, {
  taskUpdateMutation,
} from "/components/tasks/TaskFormDialog";
import { updateStoreCounter } from "/hooks/useMutationWithIncrement";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { PermissionsContext } from "/providers/PermissionsContext";
import type { ItemOf } from "/types";

type Props = {
  tasks: ({
    assignedTo?: {
      id: string;
      fullName: string;
    } | null;
    id: string;
    name: string;
    state: "TODO" | "DONE";
    description?: string | null;
    timeEstimate?: string | null;
    deadline?: string | null;
    measure?: {
      id: string;
      name: string;
    } | null;
  } & TaskFormDialogFragment$key)[];
  connectionId: string;
};

export default function TasksCard({ tasks, connectionId }: Props) {
  const { __ } = useTranslate();
  const hash = useLocation().hash.replace("#", "");

  const hashes = [
    { hash: "", label: __("To do"), state: "TODO" },
    { hash: "done", label: __("Done"), state: "DONE" },
    { hash: "all", label: __("All"), state: null },
  ] as const;

  const tasksPerHash = new Map([
    ["", tasks?.filter((t) => t.state === "TODO")],
    ["done", tasks?.filter((t) => t.state === "DONE")],
    ["all", tasks],
  ]);

  const filteredTasks = tasksPerHash.get(hash) ?? [];

  usePageTitle(__("Tasks"));

  const { isAuthorized } = use(PermissionsContext);

  const hasAnyAction =
    isAuthorized("Task", "updateTask") || isAuthorized("Task", "deleteTask");

  return (
    <div className="space-y-6">
      {tasks?.length === 0 ? (
        <p className="text-center py-6 text-txt-secondary">{__("No tasks")}</p>
      ) : (
        <Card>
          <Tabs className="px-6">
            {hashes.map((h) => (
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
              ? // All tabs group the todo using the state
                hashes
                  .slice(0, 2)
                  .filter((h) => tasksPerHash.get(h.hash)?.length)
                  .map((h) => (
                    <Fragment key={h.label}>
                      <h2 className="px-6 py-3 text-sm font-medium flex items-center gap-2 bg-subtle">
                        <TaskStateIcon state={h.state!} />
                        {h.label}
                      </h2>
                      {tasksPerHash.get(h.hash)?.map((task) => (
                        <TaskRow
                          key={task.id}
                          task={task}
                          connectionId={connectionId}
                          hasAnyAction={hasAnyAction}
                        />
                      ))}
                    </Fragment>
                  ))
              : // Todo and Done tab simply list todos
                filteredTasks?.map((task) => (
                  <TaskRow
                    key={task.id}
                    task={task}
                    connectionId={connectionId}
                    hasAnyAction={hasAnyAction}
                  />
                ))}
          </div>
        </Card>
      )}
    </div>
  );
}

type TaskRowProps = {
  task: ItemOf<Props["tasks"]>;
  connectionId: string;
  hasAnyAction: boolean;
};

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
  const { isAuthorized } = use(PermissionsContext);

  const relayEnv = useRelayEnvironment();
  const [updateTask, isUpdating] = useMutation(taskUpdateMutation);

  const onToggle = () => {
    promisifyMutation(updateTask)({
      variables: {
        input: {
          taskId: props.task.id,
          state: props.task.state === "TODO" ? "DONE" : "TODO",
        },
      },
    });
  };

  const onDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteTask)({
          variables: {
            input: { taskId: props.task.id },
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
      <TaskFormDialog task={props.task} ref={dialogRef} />
      <div className="flex items-center justify-between py-3 px-6">
        <div className="flex gap-2 items-start">
          <div className="flex items-center gap-2 pt-[2px]">
            <PriorityLevel level={1} />
            <button
              onClick={onToggle}
              className="cursor-pointer -m-1 p-1 disabled:opacity-60"
              disabled={isUpdating}
            >
              <TaskStateIcon state={props.task.state} />
            </button>
          </div>
          <div className="text-sm space-y-1 flex-1">
            <h2 className="font-medium">{props.task.name}</h2>
            {props.task.description && (
              <p className="text-txt-secondary whitespace-pre-wrap break-words">
                {props.task.description}
              </p>
            )}
            <div className="flex flex-wrap items-center gap-3 text-txt-secondary text-xs">
              {props.task.measure && (
                <span className="flex items-center gap-1">
                  <IconArrowCornerDownLeft className="scale-x-[-1]" size={14} />
                  <Link
                    className="hover:underline"
                    to={`/organizations/${organizationId}/measures/${props.task.measure?.id}`}
                  >
                    {props.task.measure?.name}
                  </Link>
                </span>
              )}
              {props.task.timeEstimate && (
                <span>{formatDuration(props.task.timeEstimate, __)}</span>
              )}
              {props.task.deadline && (
                <time dateTime={props.task.deadline}>
                  {formatDate(props.task.deadline)}
                </time>
              )}
            </div>
          </div>
        </div>
        {props.task.assignedTo?.fullName && (
          <div className="text-sm text-txt-secondary ml-auto mr-8">
            <Link
              className="hover:underline"
              to={`/organizations/${organizationId}/people/${props.task.assignedTo.id}`}
            >
              {props.task.assignedTo.fullName}
            </Link>
          </div>
        )}
        <div className="flex gap-2 items-center">
          {isUpdating && <Spinner size={16} />}
          {props.hasAnyAction && (
            <ActionDropdown>
              {isAuthorized("Task", "updateTask") && (
                <DropdownItem
                  icon={IconPencil}
                  onClick={() => dialogRef.current?.open()}
                >
                  {__("Edit")}
                </DropdownItem>
              )}
              {isAuthorized("Task", "deleteTask") && (
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
