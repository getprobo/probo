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
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectLabel } from "@probo/ui/src/v2/Select/SelectLabel";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { type ReactElement, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import { isRichEditorContentEmpty } from "../_lib/richEditorContent";
import type { TaskPriority, TaskState } from "../_lib/taskState";
import {
  taskPriorities,
  taskStateKeys,
  taskStates,
} from "../_lib/taskState";
import { useCreateTask } from "../_lib/useCreateTask";
import { createTaskDialog } from "../variants";

const taskNameMaxLength = 1000;

type DialogOpenChangeDetails = Parameters<
  NonNullable<DialogProps["onOpenChange"]>
>[1];

function isRichEditorFloatingDismiss(details: DialogOpenChangeDetails) {
  if (details.reason !== "outside-press" && details.reason !== "focus-out") {
    return false;
  }

  const target = details.event.target;
  return target instanceof Element && target.closest("[data-rich-editor-floating]") != null;
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
  const bodyRef = useRef<HTMLDivElement>(null);
  const { form, fields, descriptionField, editor, value } = createTaskDialog();

  function reset() {
    setName("");
    setContent("");
    setEditorKey(key => key + 1);
    setErrors({});
    setState("TODO");
    setPriority("MEDIUM");
  }

  function handleOpenChange(next: boolean, details: DialogOpenChangeDetails) {
    if (!next && isRichEditorFloatingDismiss(details)) {
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
      () => {
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
          <DialogBody ref={bodyRef} className={fields()}>
            <Field label={t("detailsPage.fields.name")} error={errors.name}>
              <TextField
                name="name"
                required
                maxLength={taskNameMaxLength}
                value={name}
                disabled={isCreating}
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
                disabled={isCreating}
                aria-label={t("detailsPage.fields.description")}
                onChangeContent={setContent}
              />
            </Field>
            <Select
              value={state}
              disabled={isCreating}
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
              <SelectPopup container={bodyRef}>
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
              disabled={isCreating}
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
              <SelectPopup container={bodyRef}>
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
          </DialogBody>
          <DialogFooter>
            <DialogClose
              render={(
                <Button variant="soft" color="neutral" disabled={isCreating}>
                  {t("detailsPage.actions.cancel")}
                </Button>
              )}
            />
            <Button
              type="submit"
              variant="solid"
              color="neutral"
              highContrast
              loading={isCreating}
            >
              {t("createDialog.actions.create")}
            </Button>
          </DialogFooter>
        </Form>
      </DialogPopup>
    </Dialog>
  );
}
