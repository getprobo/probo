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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { createTaskDialog } from "../../variants";

import { CreateTaskDialogFields } from "./CreateTaskDialogFields";
import type { CreateTaskDialogFormState } from "./useCreateTaskDialogForm";

export interface CreateTaskDialogFallbackProps {
  form: CreateTaskDialogFormState;
  onRetry: () => void;
}

export function CreateTaskDialogFallback({
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
