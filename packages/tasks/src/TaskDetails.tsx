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

import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "./i18n";
import { taskDetails } from "./variants";

export interface TaskDetailsProps {
  name: ReactNode;
  overdue?: boolean;
  actions?: ReactNode;
  description: ReactNode;
  engagement?: ReactNode;
  properties: ReactNode;
}

export function TaskDetails({
  name,
  overdue = false,
  actions,
  description,
  engagement,
  properties,
}: TaskDetailsProps) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const slots = taskDetails();

  return (
    <div className={slots.root()}>
      <div className={slots.header()}>
        <div className={slots.titleRow()}>
          <div className={slots.title()}>
            {name}
            {overdue && (
              <Badge variant="soft" color="red" size={1}>
                {t("metadata.overdue")}
              </Badge>
            )}
          </div>
          {actions != null && (
            <div className={slots.actions()}>
              {actions}
            </div>
          )}
        </div>
      </div>
      <div className={slots.body()}>
        <div className={slots.main()}>
          {description}
          {engagement}
        </div>
        {properties}
      </div>
    </div>
  );
}
