// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { Form } from "@base-ui/react/form";
import { CaretLeftIcon, UserCircleDashedIcon, XIcon } from "@phosphor-icons/react";
import { clampUtf8 } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import {
  TaskDetails,
  taskListItem,
  TaskListItemAssignee,
  taskNameField,
  taskNameMaxLength,
  type TaskPriority,
  TaskProperties,
  type TaskPropertiesChange,
  TASKS_NAMESPACE,
  type TaskState,
} from "@probo/tasks";
import { Linear } from "@probo/ui/src/Atoms/ThirdParties/Linear";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { Navigate, useLocation, useNavigate } from "react-router";

import type { NewTaskPageLinkMutation } from "#/__generated__/core/NewTaskPageLinkMutation.graphql";
import type { NewTaskPagePublishMutation } from "#/__generated__/core/NewTaskPagePublishMutation.graphql";
import type { NewTaskPageQuery } from "#/__generated__/core/NewTaskPageQuery.graphql";
import type { TaskLinearDraftField_organization$key } from "#/__generated__/core/TaskLinearDraftField_organization.graphql";
import {
  TaskAssigneeMenu,
  type TaskAssigneeOption,
} from "#/components/tasks/TaskAssigneeMenu";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { RichEditor } from "#/pages/organizations/_components/RichEditor";
import { isRichEditorContentEmpty } from "#/pages/organizations/_lib/richEditorContent";
import { useMentionMembers } from "#/pages/organizations/_lib/useMentionMembers";

import { InternalControlCombobox, type InternalControlOption } from "./_components/InternalControlCombobox";
import {
  type LinearDraft,
  type LinearIssueDraft,
  TaskLinearDraftField,
} from "./_components/TaskLinearDraftField";
import {
  taskBackTarget,
  taskDetailsPath,
  taskListPath,
  taskListSearch,
} from "./_lib/taskPath";
import { useCreateTask } from "./_lib/useCreateTask";
import { taskInternalControls, taskLinearField } from "./variants";

const descriptionEditorClassName = "min-h-40 bg-transparent! py-0! pr-0! shadow-none!";
const unassignedValue = "__NONE__";

const publishMutation = graphql`
  mutation NewTaskPagePublishMutation($input: PublishTaskToLinearInput!) {
    publishTaskToLinear(input: $input) {
      task {
        id
        ...TaskListItem_task
      }
    }
  }
`;

const linkMutation = graphql`
  mutation NewTaskPageLinkMutation($input: LinkTaskToLinearInput!) {
    linkTaskToLinear(input: $input) {
      task {
        ...TaskDetailsPage_task
        ...TaskListItem_task
      }
    }
  }
`;

export const newTaskPageQuery = graphql`
  query NewTaskPageQuery(
    $organizationId: ID!
    $internalControlId: ID!
    $includeInternalControl: Boolean!
  ) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateTask: permission(action: "core:task:create")
        canListInternalControls: permission(action: "core:internal-control:list")
        connectors(filter: { providers: [LINEAR_SYNC] }) {
          __typename
        }
        linearSettings {
          defaultTeam {
            id
          }
        }
        ...TaskLinearDraftField_organization
      }
    }
    internalControl: node(id: $internalControlId) @include(if: $includeInternalControl) {
      __typename
      ... on InternalControl {
        id
        name
      }
    }
  }
`;

interface NewTaskPageProps {
  queryRef: PreloadedQuery<NewTaskPageQuery>;
}

export function NewTaskPage({ queryRef }: NewTaskPageProps) {
  const data = usePreloadedQuery<NewTaskPageQuery>(newTaskPageQuery, queryRef);
  const { t } = useTranslation("organizations/tasks");
  const { t: tTasks } = useTranslation(TASKS_NAMESPACE);
  const navigate = useNavigate();
  const location = useLocation();
  const organizationId = useOrganizationId();
  const mentionSource = useMentionMembers();
  const organization = data.organization?.__typename === "Organization"
    ? data.organization
    : null;
  const preselected = preselectedInternalControl(data.internalControl);
  const [name, setName] = useState("");
  const [content, setContent] = useState("");
  const [editorKey, setEditorKey] = useState(0);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [state, setState] = useState<TaskState>("TODO");
  const [priority, setPriority] = useState<TaskPriority>("MEDIUM");
  const [assignee, setAssignee] = useState<TaskAssigneeOption | null>(null);
  const [timeEstimate, setTimeEstimate] = useState<string | null>(null);
  const [deadline, setDeadline] = useState<string | null>(null);
  const [recurrenceInterval, setRecurrenceInterval] = useState<string | null>(null);
  const [internalControls, setInternalControls] = useState<InternalControlOption[]>(
    preselected == null ? [] : [preselected],
  );
  const [linear, setLinear] = useState<LinearDraft | null>(null);
  const [uploading, setUploading] = useState(false);
  const submitLock = useRef(false);
  const [createTask, isCreating] = useCreateTask();
  const [publish, isPublishing] = useMutation<NewTaskPagePublishMutation>(
    publishMutation,
    {
      successMessage: t("detailsPage.linear.published"),
      errorToast: t("detailsPage.linear.errors.publish"),
    },
  );
  const [link, isLinking] = useMutation<NewTaskPageLinkMutation>(
    linkMutation,
    {
      successMessage: t("detailsPage.linear.linked"),
      errorToast: t("detailsPage.linear.errors.link"),
    },
  );
  const isSaving = isCreating || isPublishing || isLinking;
  const { root: nameRoot, input: nameInput } = taskNameField();
  const { stateOption } = taskListItem();
  usePageTitle(name.trim() || t("newPage.title"));

  if (organization == null) {
    throw new Error("invalid type for node");
  }

  if (!organization.canCreateTask) {
    return (
      <Navigate
        to={{
          pathname: taskListPath(organizationId),
          search: taskListSearch(location.search),
        }}
        replace
      />
    );
  }

  const linearConnected = organization.connectors.length > 0;
  const defaultTeamId = linearConnected
    ? (organization.linearSettings.defaultTeam?.id ?? null)
    : null;
  const showInternalControls = organization.canListInternalControls
    || internalControls.length > 0;
  const backTarget = taskBackTarget(organizationId, location.search);

  function clearNameError(next: string) {
    if (!next.trim()) {
      return;
    }
    setErrors((current) => {
      if (current.name == null) {
        return current;
      }
      const nextErrors = { ...current };
      delete nextErrors.name;
      return nextErrors;
    });
  }

  function applyLinearIssue(issue: LinearIssueDraft) {
    setName(clampUtf8(issue.title.trim(), taskNameMaxLength));
    setContent(issue.content ?? "");
    setEditorKey(key => key + 1);
    if (issue.state != null) {
      setState(issue.state);
    }
    if (issue.priority != null) {
      setPriority(issue.priority);
    }
    setErrors({});
  }

  function onPropertiesChange(change: TaskPropertiesChange) {
    if (change.state != null) {
      setState(change.state);
    }
    if (change.priority != null) {
      setPriority(change.priority);
    }
    if ("timeEstimate" in change) {
      setTimeEstimate(change.timeEstimate ?? null);
    }
    if ("deadline" in change) {
      setDeadline(change.deadline ?? null);
    }
    if ("recurrenceInterval" in change) {
      setRecurrenceInterval(change.recurrenceInterval ?? null);
    }
  }

  function submit() {
    if (submitLock.current || uploading || isSaving) {
      return;
    }

    const nextName = name.trim();
    if (!nextName) {
      setErrors({ name: t("newPage.errors.nameRequired") });
      return;
    }
    setErrors({});
    submitLock.current = true;

    const selection = linear ?? {
      teamId: defaultTeamId,
      issueId: null,
    };

    void createTask({
      name: nextName,
      content: isRichEditorContentEmpty(content) ? null : content,
      state,
      priority,
      internalControlIds: internalControls.map(control => control.id),
      // Explicit null skips the server's default-team publish.
      linearTeamId: null,
      assignedToId: assignee?.id,
      timeEstimate,
      deadline,
      recurrenceInterval,
    }).then(
      async (taskId) => {
        try {
          if (selection.teamId != null) {
            if (selection.issueId != null) {
              await link({
                variables: {
                  input: {
                    taskId,
                    teamId: selection.teamId,
                    issueId: selection.issueId,
                  },
                },
              });
            } else {
              await publish({
                variables: {
                  input: {
                    taskId,
                    teamId: selection.teamId,
                  },
                },
              });
            }
          }
        } catch {
          // The task exists; the Linear toast already explains the failure.
        }
        void navigate({
          pathname: taskDetailsPath(organizationId, taskId),
          search: location.search,
        });
      },
      () => {
        submitLock.current = false;
      },
    );
  }

  return (
    <Form noValidate errors={errors} onFormSubmit={submit}>
      <TaskDetails
        back={(
          <Link
            to={backTarget}
            size={2}
            color="neutral"
            underline={false}
            iconStart={<CaretLeftIcon />}
            className="self-start"
          >
            {t("detailsPage.back")}
          </Link>
        )}
        name={(
          <div className={nameRoot()}>
            <Field error={errors.name}>
              <input
                name="name"
                className={nameInput()}
                value={name}
                required
                maxLength={taskNameMaxLength}
                disabled={isSaving}
                placeholder={t("newPage.fields.namePlaceholder")}
                aria-label={t("detailsPage.fields.name")}
                onChange={(event) => {
                  const next = clampUtf8(event.currentTarget.value, taskNameMaxLength);
                  setName(next);
                  clearNameError(next);
                }}
              />
            </Field>
          </div>
        )}
        actions={(
          <Button
            type="submit"
            variant="solid"
            color="neutral"
            highContrast
            disabled={uploading}
            loading={isSaving}
          >
            {t("newPage.actions.create")}
          </Button>
        )}
        description={(
          <RichEditor
            key={editorKey}
            className={descriptionEditorClassName}
            controls={false}
            content={content}
            disabled={isSaving}
            placeholder={t("detailsPage.fields.descriptionPlaceholder")}
            aria-label={t("detailsPage.fields.description")}
            mentionSource={mentionSource}
            onChangeContent={setContent}
            onUploadingChange={setUploading}
          />
        )}
        properties={(
          <div className="flex flex-col gap-4">
            <TaskProperties
              task={{
                state,
                priority,
                timeEstimate,
                deadline,
                recurrenceInterval,
              }}
              canUpdate
              isUpdating={isSaving}
              assignee={(
                <TaskListItemAssignee
                  assignee={assignee}
                  canUpdate
                  isUpdating={isSaving}
                  showLabel
                  placeholder={tTasks("placeholders.assignee")}
                  className={stateOption()}
                  renderAssigneeMenu={() => (
                    <TaskAssigneeMenu
                      value={assignee?.id ?? null}
                      emptyLabel={t("detailsPage.unassigned")}
                      emptyValue={unassignedValue}
                      emptyIcon={<UserCircleDashedIcon />}
                      disabled={isSaving}
                      onValueChange={(assignedToId) => {
                        if (assignedToId == null) {
                          setAssignee(null);
                        }
                      }}
                      onSelect={setAssignee}
                    />
                  )}
                />
              )}
              onChange={onPropertiesChange}
            />
            {linearConnected && (
              <NewTaskLinearCard
                organizationKey={organization}
                disabled={isSaving}
                onChange={setLinear}
                onIssue={applyLinearIssue}
              />
            )}
            {showInternalControls && (
              <NewTaskInternalControls
                controls={internalControls}
                canList={organization.canListInternalControls}
                disabled={isSaving}
                onChange={setInternalControls}
              />
            )}
          </div>
        )}
      />
    </Form>
  );
}

function preselectedInternalControl(
  node: NewTaskPageQuery["response"]["internalControl"],
): InternalControlOption | null {
  if (node == null || node.__typename !== "InternalControl") {
    return null;
  }
  return { id: node.id, name: node.name };
}

function NewTaskLinearCard({
  organizationKey,
  disabled,
  onChange,
  onIssue,
}: {
  organizationKey: TaskLinearDraftField_organization$key;
  disabled: boolean;
  onChange: (selection: LinearDraft) => void;
  onIssue: (issue: LinearIssueDraft) => void;
}) {
  const { t } = useTranslation("organizations/tasks");
  const { root, title, titleIcon } = taskLinearField();

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <div className={title()}>
          <Linear className={titleIcon()} aria-hidden />
          <Heading level={2} size={2} weight="medium">
            {t("detailsPage.linear.linkHint")}
          </Heading>
        </div>
        <TaskLinearDraftField
          organizationKey={organizationKey}
          disabled={disabled}
          onChange={onChange}
          onIssue={onIssue}
        />
      </div>
    </Card>
  );
}

function NewTaskInternalControls({
  controls,
  canList,
  disabled,
  onChange,
}: {
  controls: readonly InternalControlOption[];
  canList: boolean;
  disabled: boolean;
  onChange: (controls: InternalControlOption[]) => void;
}) {
  const { t } = useTranslation("organizations/tasks");
  const { root, names, name, control } = taskInternalControls();
  const linkedIds = new Set(controls.map(item => item.id));

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <Heading level={2} size={2} weight="medium">
          {t("detailsPage.linkInternalControls")}
        </Heading>
        <div className={names()}>
          {controls.length > 0
            ? controls.map(item => (
                <div key={item.id} className={name()}>
                  <Text size={1} title={item.name} className={control()}>
                    {item.name}
                  </Text>
                  <IconButton
                    size={1}
                    variant="ghost"
                    color="neutral"
                    disabled={disabled}
                    className="shrink-0"
                    aria-label={t("detailsPage.actions.removeInternalControl")}
                    onClick={() => {
                      onChange(controls.filter(current => current.id !== item.id));
                    }}
                  >
                    <XIcon />
                  </IconButton>
                </div>
              ))
            : (
                <Text size={1} color="faint">{t("detailsPage.none")}</Text>
              )}
          {canList && (
            <InternalControlCombobox
              excludeIds={linkedIds}
              disabled={disabled}
              onSelect={(internalControl) => {
                if (linkedIds.has(internalControl.id)) {
                  return;
                }
                onChange([...controls, internalControl]);
              }}
            />
          )}
        </div>
      </div>
    </Card>
  );
}
