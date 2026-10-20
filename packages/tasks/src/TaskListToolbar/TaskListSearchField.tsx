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

import { MagnifyingGlassIcon } from "@phosphor-icons/react";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "../i18n";
import { taskListToolbar } from "../variants";

export interface TaskListSearchFieldProps {
  value: string;
  onValueChange: (value: string) => void;
}

export function TaskListSearchField({
  value,
  onValueChange,
}: TaskListSearchFieldProps) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const { search } = taskListToolbar();

  return (
    <div className={search()}>
      <TextField
        icon={<MagnifyingGlassIcon />}
        value={value}
        onValueChange={onValueChange}
        placeholder={t("searchPlaceholder")}
        aria-label={t("searchPlaceholder")}
      />
    </div>
  );
}
