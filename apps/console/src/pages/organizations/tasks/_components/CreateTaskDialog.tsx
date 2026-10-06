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
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { Dialog, type DialogProps } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectLabel } from "@probo/ui/src/v2/Select/SelectLabel";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactElement, type ReactNode, Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";

import type { CreateTaskDialogLinearQuery } from "#/__generated__/core/CreateTaskDialogLinearQuery.graphql";
import type { CreateTaskDialogLinkMutation } from "#/__generated__/core/CreateTaskDialogLinkMutation.graphql";
import type { CreateTaskDialogPublishMutation } from "#/__generated__/core/CreateTaskDialogPublishMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { RichEditor } from "#/pages/organizations/_components/RichEditor";
import { isRichEditorContentEmpty } from "#/pages/organizations/_lib/richEditorContent";

import type { TaskPriority, TaskState } from "../_lib/taskState";
import {
  taskPriorities,
  taskStateKeys,
  taskStates,
} from "../_lib/taskState";
import { useCreateTask } from "../_lib/useCreateTask";
import { createTaskDialog } from "../variants";

import { type LinearDraft, type LinearIssueDraft, TaskLinearDraftField } from "./TaskLinearDraftField";
import { TaskPriorityIcon } from "./TaskPriorityIcon";
import { TaskStateIcon } from "./TaskStateIcon";

const linearConnectionQuery = graphql`
  query CreateTaskDialogLinearQuery($organizationId: ID!) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
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
  }
`;

const publishMutation = graphql`
  mutation CreateTaskDialogPublishMutation($input: PublishTaskToLinearInput!) {
    publishTaskToLinear(input: $input) {
      task {
        id
      }
    }
  }
`;

const linkMutation = graphql`
  mutation CreateTaskDialogLinkMutation($input: LinkTaskToLinearInput!) {
    linkTaskToLinear(input: $input) {
      task {
        ...TaskDetailsPage_task
        ...TaskListItem_task
      }
    }
  }
`;

const taskNameMaxLength = 1000;

// Server name validation counts UTF-8 bytes, which is stricter than the
// input's character maxLength for non-ASCII titles.
function clampTaskName(title: string) {
  const encoder = new TextEncoder();
  let result = "";
  let bytes = 0;

  for (const char of title.trim()) {
    const size = encoder.encode(char).length;
    if (bytes + size > taskNameMaxLength) {
      break;
    }

    result += char;
    bytes += size;
  }

  return result;
}

type DialogOpenChangeDetails = Parameters<
  NonNullable<DialogProps["onOpenChange"]>
>[1];

function isFloatingDismiss(details: DialogOpenChangeDetails) {
  if (details.reason !== "outside-press" && details.reason !== "focus-out") {
    return false;
  }

  const nodes: Array<EventTarget | null> = [details.event.target];
  if ("relatedTarget" in details.event) {
    nodes.push(details.event.relatedTarget);
  }

  return nodes.some(node =>
    node instanceof Element
    && (
      node.closest("[data-rich-editor-floating]") != null
      || node.closest("[data-dialog-overlay-root]") != null
    ),
  );
}

interface CreateTaskDialogSessionProps {
  connectionId: string;
  internalControlId?: string;
  onCompleted?: () => void;
  onClose: () => void;
}

interface CreateTaskDialogFormState {
  name: string;
  content: string;
  editorKey: number;
  errors: Record<string, string>;
  state: TaskState;
  priority: TaskPriority;
  linear: LinearDraft | null;
  uploading: boolean;
  isSaving: boolean;
  setName: (value: string) => void;
  setContent: (value: string) => void;
  setErrors: (update: (current: Record<string, string>) => Record<string, string>) => void;
  setState: (state: TaskState) => void;
  setPriority: (priority: TaskPriority) => void;
  setLinear: (selection: LinearDraft) => void;
  setUploading: (uploading: boolean) => void;
  applyLinearIssue: (issue: LinearIssueDraft) => void;
  reset: () => void;
  submit: (selection: LinearDraft) => void;
}

function useCreateTaskDialogForm({
  connectionId,
  internalControlId,
  onCompleted,
  onClose,
}: CreateTaskDialogSessionProps): CreateTaskDialogFormState {
  const { t } = useTranslation("organizations/tasks");
  const [name, setName] = useState("");
  const [content, setContent] = useState("");
  const [uploading, setUploading] = useState(false);
  const [editorKey, setEditorKey] = useState(0);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [state, setState] = useState<TaskState>("TODO");
  const [priority, setPriority] = useState<TaskPriority>("MEDIUM");
  const [createTask, isCreating] = useCreateTask();
  const [linear, setLinearState] = useState<LinearDraft | null>(null);
  const [publish, isPublishing] = useMutation<CreateTaskDialogPublishMutation>(
    publishMutation,
    {
      successMessage: t("detailsPage.linear.published"),
      errorToast: t("detailsPage.linear.errors.publish"),
    },
  );
  const [link, isLinking] = useMutation<CreateTaskDialogLinkMutation>(
    linkMutation,
    {
      successMessage: t("detailsPage.linear.linked"),
      errorToast: t("detailsPage.linear.errors.link"),
    },
  );
  const isSaving = isCreating || isPublishing || isLinking;

  function applyLinearIssue(issue: LinearIssueDraft) {
    setName(clampTaskName(issue.title));
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

  function reset() {
    setName("");
    setContent("");
    setEditorKey(key => key + 1);
    setErrors({});
    setState("TODO");
    setPriority("MEDIUM");
    setLinearState(null);
    setUploading(false);
  }

  function submit(linearSelection: LinearDraft) {
    if (uploading) {
      return;
    }

    const nextName = name.trim();
    if (!nextName) {
      setErrors({ name: t("createDialog.errors.nameRequired") });
      return;
    }
    setErrors({});

    void createTask(
      {
        name: nextName,
        content: isRichEditorContentEmpty(content) ? null : content,
        state,
        priority,
        internalControlIds: internalControlId ? [internalControlId] : undefined,
        linearTeamId: null,
      },
      connectionId,
    ).then(
      async (taskId) => {
        try {
          if (linearSelection.teamId != null) {
            if (linearSelection.issueId != null) {
              await link({
                variables: {
                  input: {
                    taskId,
                    teamId: linearSelection.teamId,
                    issueId: linearSelection.issueId,
                  },
                },
              });
            } else {
              await publish({
                variables: {
                  input: {
                    taskId,
                    teamId: linearSelection.teamId,
                  },
                },
              });
            }
          }
        } catch {
          // The task exists; the Linear toast already explains the failure.
        }
        reset();
        onClose();
        onCompleted?.();
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return {
    name,
    content,
    editorKey,
    errors,
    state,
    priority,
    linear,
    uploading,
    isSaving,
    setName,
    setContent,
    setErrors,
    setState,
    setPriority,
    setLinear: setLinearState,
    setUploading,
    applyLinearIssue,
    reset,
    submit,
  };
}

interface CreateTaskDialogFieldsProps {
  form: CreateTaskDialogFormState;
  linear: ReactNode;
  submitDisabled?: boolean;
  onSubmit: () => void;
}

function CreateTaskDialogFields({
  form,
  linear,
  submitDisabled,
  onSubmit,
}: CreateTaskDialogFieldsProps) {
  const { t } = useTranslation("organizations/tasks");
  const { form: formClass, fields, descriptionField, editor, value } = createTaskDialog();
  const {
    name,
    content,
    editorKey,
    errors,
    state,
    priority,
    isSaving,
    uploading,
    setUploading,
    setName,
    setContent,
    setErrors,
    setState,
    setPriority,
  } = form;

  return (
    <Form className={formClass()} errors={errors} onFormSubmit={onSubmit}>
      <DialogHeader>
        <DialogTitle>{t("createDialog.title")}</DialogTitle>
      </DialogHeader>
      <DialogBody className={fields()}>
        <Field label={t("detailsPage.fields.name")} error={errors.name}>
          <TextField
            name="name"
            required
            maxLength={taskNameMaxLength}
            value={name}
            disabled={isSaving}
            placeholder={t("createDialog.fields.namePlaceholder")}
            onValueChange={(next) => {
              setName(next);
              if (next.trim()) {
                setErrors((current) => {
                  const nextErrors = { ...current };
                  delete nextErrors.name;
                  return nextErrors;
                });
              }
            }}
          />
        </Field>
        <Field className={descriptionField()} label={t("detailsPage.fields.description")}>
          <RichEditor
            key={editorKey}
            className={editor()}
            content={content}
            disabled={isSaving}
            aria-label={t("detailsPage.fields.description")}
            onChangeContent={setContent}
            onUploadingChange={setUploading}
          />
        </Field>
        <Select
          value={state}
          disabled={isSaving}
          onValueChange={(next: TaskState | null) => {
            if (next) {
              setState(next);
            }
          }}
        >
          <Field>
            <SelectLabel>{t("detailsPage.fields.state")}</SelectLabel>
            <SelectTrigger>
              {(selected: TaskState | null) =>
                selected
                  ? (
                      <span className={value()}>
                        <TaskStateIcon state={selected} />
                        {t(`detailsPage.states.${taskStateKeys[selected]}`)}
                      </span>
                    )
                  : null}
            </SelectTrigger>
          </Field>
          <SelectPopup>
            {taskStates.map(item => (
              <SelectItem key={item} value={item}>
                <span className={value()}>
                  <TaskStateIcon state={item} />
                  {t(`detailsPage.states.${taskStateKeys[item]}`)}
                </span>
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
        <Select
          value={priority}
          disabled={isSaving}
          onValueChange={(next: TaskPriority | null) => {
            if (next) {
              setPriority(next);
            }
          }}
        >
          <Field>
            <SelectLabel>{t("detailsPage.fields.priority")}</SelectLabel>
            <SelectTrigger>
              {(selected: TaskPriority | null) =>
                selected
                  ? (
                      <span className={value()}>
                        <TaskPriorityIcon priority={selected} />
                        {t(`detailsPage.priorities.${selected.toLowerCase()}`)}
                      </span>
                    )
                  : null}
            </SelectTrigger>
          </Field>
          <SelectPopup>
            {taskPriorities.map(item => (
              <SelectItem key={item} value={item}>
                <span className={value()}>
                  <TaskPriorityIcon priority={item} />
                  {t(`detailsPage.priorities.${item.toLowerCase()}`)}
                </span>
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
        {linear}
      </DialogBody>
      <DialogFooter>
        <DialogClose
          render={(
            <Button variant="soft" color="neutral" disabled={isSaving}>
              {t("detailsPage.actions.cancel")}
            </Button>
          )}
        />
        <Button
          type="submit"
          variant="solid"
          color="neutral"
          highContrast
          disabled={submitDisabled || uploading}
          loading={isSaving}
        >
          {t("createDialog.actions.create")}
        </Button>
      </DialogFooter>
    </Form>
  );
}

function CreateTaskDialogSkeleton() {
  const { t } = useTranslation("organizations/tasks");
  const { fields } = createTaskDialog();

  return (
    <>
      <DialogHeader>
        <DialogTitle>{t("createDialog.title")}</DialogTitle>
      </DialogHeader>
      <DialogBody className={fields()}>
        <SelectSkeleton className="w-full" />
        <SelectSkeleton className="w-full" />
        <SelectSkeleton className="w-full" />
      </DialogBody>
    </>
  );
}

interface CreateTaskDialogLoadedProps {
  queryRef: PreloadedQuery<CreateTaskDialogLinearQuery>;
  form: CreateTaskDialogFormState;
}

function CreateTaskDialogLoaded({
  queryRef,
  form,
}: CreateTaskDialogLoadedProps) {
  const { t } = useTranslation("organizations/tasks");
  const data = usePreloadedQuery<CreateTaskDialogLinearQuery>(
    linearConnectionQuery,
    queryRef,
  );
  const organization = data.node?.__typename === "Organization" ? data.node : null;
  const defaultTeamId = organization != null && organization.connectors.length > 0
    ? (organization.linearSettings.defaultTeam?.id ?? null)
    : null;

  return (
    <CreateTaskDialogFields
      form={form}
      linear={organization != null && organization.connectors.length > 0
        ? (
            <Field label={t("detailsPage.fields.linear")}>
              <TaskLinearDraftField
                organizationKey={organization}
                disabled={form.isSaving}
                onChange={form.setLinear}
                onIssue={form.applyLinearIssue}
              />
            </Field>
          )
        : null}
      onSubmit={() => {
        form.submit(form.linear ?? {
          teamId: defaultTeamId,
          issueId: null,
        });
      }}
    />
  );
}

interface CreateTaskDialogFallbackProps {
  form: CreateTaskDialogFormState;
  onRetry: () => void;
}

function CreateTaskDialogFallback({
  form,
  onRetry,
}: CreateTaskDialogFallbackProps) {
  const { t } = useTranslation("organizations/tasks");
  const [skipLinear, setSkipLinear] = useState(false);
  const { linearFallback, linearFallbackChoice } = createTaskDialog();

  return (
    <CreateTaskDialogFields
      form={form}
      submitDisabled={!skipLinear}
      linear={(
        <Field label={t("detailsPage.fields.linear")}>
          <div className={linearFallback()}>
            <Text size={2} color="faint">{t("detailsPage.linear.errors.load")}</Text>
            <Button
              type="button"
              variant="ghost"
              color="neutral"
              disabled={form.isSaving}
              onClick={onRetry}
            >
              {t("detailsPage.linear.retry")}
            </Button>
            <label className={linearFallbackChoice()}>
              <Checkbox
                checked={skipLinear}
                disabled={form.isSaving}
                onCheckedChange={checked => setSkipLinear(checked === true)}
              />
              <Text size={2}>{t("detailsPage.linear.createWithoutTeam")}</Text>
            </label>
          </div>
        </Field>
      )}
      onSubmit={() => {
        if (!skipLinear) {
          return;
        }
        form.submit({ teamId: null, issueId: null });
      }}
    />
  );
}

interface CreateTaskDialogProps {
  connectionId: string;
  internalControlId?: string;
  children: ReactElement;
  onCompleted?: () => void;
}

export function CreateTaskDialog({
  connectionId,
  internalControlId,
  children,
  onCompleted,
}: CreateTaskDialogProps) {
  const organizationId = useOrganizationId();
  const [open, setOpen] = useState(false);
  const [session, setSession] = useState(0);
  const [queryRef, loadQuery] = useQueryLoader<CreateTaskDialogLinearQuery>(
    linearConnectionQuery,
  );

  function close() {
    setOpen(false);
  }

  const form = useCreateTaskDialogForm({
    connectionId,
    internalControlId,
    onCompleted,
    onClose: close,
  });

  function handleOpenChange(next: boolean, details: DialogOpenChangeDetails) {
    if (!next && isFloatingDismiss(details)) {
      details.cancel();
      return;
    }

    if (next) {
      form.reset();
      loadQuery({ organizationId }, { fetchPolicy: "store-and-network" });
      setSession(current => current + 1);
    }
    setOpen(next);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={children} />
      <DialogPopup lockScroll>
        {queryRef == null
          ? (open ? <CreateTaskDialogSkeleton /> : null)
          : (
              <ErrorBoundary
                key={session}
                fallback={(_error, reset) => (
                  <CreateTaskDialogFallback
                    form={form}
                    onRetry={() => {
                      loadQuery(
                        { organizationId },
                        { fetchPolicy: "network-only" },
                      );
                      reset();
                    }}
                  />
                )}
              >
                <Suspense fallback={<CreateTaskDialogSkeleton />}>
                  <CreateTaskDialogLoaded queryRef={queryRef} form={form} />
                </Suspense>
              </ErrorBoundary>
            )}
      </DialogPopup>
    </Dialog>
  );
}
