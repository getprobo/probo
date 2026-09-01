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
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { useCreateTaskComment } from "../_lib/useCreateTaskComment";
import { taskCommentForm } from "../variants";

const taskCommentMaxLength = 5000;

export function TaskCommentForm() {
  const { t } = useTranslation("organizations/tasks");
  const [createTaskComment, isCreating] = useCreateTaskComment();
  const [description, setDescription] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const { root, actions } = taskCommentForm();

  function handleSubmit() {
    const next = description.trim();
    if (!next) {
      setErrors({ description: t("detailsPage.comments.errors.descriptionRequired") });
      return;
    }

    void createTaskComment(next).then(
      () => {
        setDescription("");
        setErrors({});
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Form className={root()} errors={errors} onFormSubmit={handleSubmit}>
      <Field
        label={t("detailsPage.comments.fields.description")}
        error={errors.description}
      >
        <Textarea
          name="description"
          rows={3}
          required
          maxLength={taskCommentMaxLength}
          value={description}
          disabled={isCreating}
          placeholder={t("detailsPage.comments.placeholder")}
          onChange={(event) => {
            setDescription(event.currentTarget.value);
            setErrors({});
          }}
        />
      </Field>
      <div className={actions()}>
        <Button
          type="submit"
          variant="solid"
          color="neutral"
          highContrast
          loading={isCreating}
        >
          {t("detailsPage.comments.actions.submit")}
        </Button>
      </div>
    </Form>
  );
}
