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
import { PriorityLevel, RichEditor, TaskStateIcon } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
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
import { type ReactElement, Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { CreateTaskDialogLinearQuery } from "#/__generated__/core/CreateTaskDialogLinearQuery.graphql";
import type { CreateTaskDialogLinkMutation } from "#/__generated__/core/CreateTaskDialogLinkMutation.graphql";
import type { CreateTaskDialogPublishMutation } from "#/__generated__/core/CreateTaskDialogPublishMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { isRichEditorContentEmpty } from "#/pages/organizations/_lib/richEditorContent";

import type { TaskPriority, TaskState } from "../_lib/taskState";
import {
  taskPriorities,
  taskStateKeys,
  taskStates,
} from "../_lib/taskState";
import { useCreateTask } from "../_lib/useCreateTask";
import { createTaskDialog } from "../variants";

import { type LinearDraft, type LinearIssueDraft, TaskLinearDraftField } from "./TaskLinearPublishField";

const linearConnectionQuery = graphql`
  query CreateTaskDialogLinearQuery($organizationId: ID!) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        connectors(filter: { providers: [LINEAR_SYNC] }) {
          id
        }
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
        ...TasksCard_TaskRowFragment
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

interface CreateTaskDialogProps {
  connectionId: string;
  measureId?: string;
  children: ReactElement;
  onCompleted?: () => void;
}

export function CreateTaskDialog({
  connectionId,
  measureId,
  children,
  onCompleted,
}: CreateTaskDialogProps) {
  const { t } = useTranslation("organizations/tasks");
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [content, setContent] = useState("");
  const [editorKey, setEditorKey] = useState(0);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [state, setState] = useState<TaskState>("TODO");
  const [priority, setPriority] = useState<TaskPriority>("MEDIUM");
  const [createTask, isCreating] = useCreateTask();
  const [linear, setLinear] = useState<LinearDraft>({ teamId: null, issueId: null });
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
  const { form, fields, descriptionField, editor, value } = createTaskDialog();

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
    setLinear({ teamId: null, issueId: null });
  }

  function handleOpenChange(next: boolean, details: DialogOpenChangeDetails) {
    if (!next && isFloatingDismiss(details)) {
      details.cancel();
      return;
    }

    setOpen(next);
    if (!next) {
      reset();
    }
  }

  function handleSubmit() {
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
        measureId,
      },
      connectionId,
    ).then(
      async (taskId) => {
        try {
          if (linear.teamId != null) {
            if (linear.issueId != null) {
              await link({
                variables: {
                  input: {
                    taskId,
                    teamId: linear.teamId,
                    issueId: linear.issueId,
                  },
                },
              });
            } else {
              await publish({
                variables: {
                  input: {
                    taskId,
                    teamId: linear.teamId,
                  },
                },
              });
            }
          }
        } catch {
          // The task exists; the Linear toast already explains the failure.
        }
        setOpen(false);
        reset();
        onCompleted?.();
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={children} />
      <DialogPopup lockScroll>
        <Form className={form()} errors={errors} onFormSubmit={handleSubmit}>
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
                            <PriorityLevel level={selected} />
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
                      <PriorityLevel level={item} />
                      {t(`detailsPage.priorities.${item.toLowerCase()}`)}
                    </span>
                  </SelectItem>
                ))}
              </SelectPopup>
            </Select>
            {open && (
              <ErrorBoundary
                fallback={(
                  <Field label={t("detailsPage.fields.linear")}>
                    <Text size={2} color="faint">{t("detailsPage.linear.errors.load")}</Text>
                  </Field>
                )}
              >
                <Suspense fallback={<SelectSkeleton className="w-full" />}>
                  <CreateTaskLinearSection
                    disabled={isSaving}
                    onChange={setLinear}
                    onIssue={applyLinearIssue}
                  />
                </Suspense>
              </ErrorBoundary>
            )}
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
              loading={isSaving}
            >
              {t("createDialog.actions.create")}
            </Button>
          </DialogFooter>
        </Form>
      </DialogPopup>
    </Dialog>
  );
}

function CreateTaskLinearSection({
  disabled,
  onChange,
  onIssue,
}: {
  disabled?: boolean;
  onChange: (selection: LinearDraft) => void;
  onIssue: (issue: LinearIssueDraft) => void;
}) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const data = useLazyLoadQuery<CreateTaskDialogLinearQuery>(
    linearConnectionQuery,
    { organizationId },
  );
  const organization = data.node?.__typename === "Organization" ? data.node : null;
  if ((organization?.connectors.length ?? 0) === 0) {
    return null;
  }

  return (
    <Field label={t("detailsPage.fields.linear")}>
      <TaskLinearDraftField disabled={disabled} onChange={onChange} onIssue={onIssue} />
    </Field>
  );
}
