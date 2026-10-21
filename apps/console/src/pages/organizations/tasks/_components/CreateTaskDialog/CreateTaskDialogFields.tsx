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
import {
  taskNameMaxLength,
  taskPriorities,
  type TaskPriority,
  TaskPriorityIcon,
  type TaskState,
  TaskStateIcon,
  taskStateKeys,
  taskStates,
} from "@probo/tasks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectLabel } from "@probo/ui/src/v2/Select/SelectLabel";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { RichEditor } from "#/pages/organizations/_components/RichEditor";

import { createTaskDialog } from "../../variants";

import type { CreateTaskDialogFormState } from "./useCreateTaskDialogForm";

export interface CreateTaskDialogFieldsProps {
  form: CreateTaskDialogFormState;
  linear: ReactNode;
  submitDisabled?: boolean;
  onSubmit: () => void;
}

export function CreateTaskDialogFields({
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
