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

import { Card } from "@probo/ui/src/v2/Card/Card";
import type { ReactNode } from "react";

import { taskDetails } from "./variants";

export interface TaskDetailsProps {
  name: ReactNode;
  actions?: ReactNode;
  description: ReactNode;
  engagement?: ReactNode;
  properties: ReactNode;
}

export function TaskDetails({
  name,
  actions,
  description,
  engagement,
  properties,
}: TaskDetailsProps) {
  const slots = taskDetails();

  return (
    <div className={slots.root()}>
      <div className={slots.header()}>
        <div className={slots.titleRow()}>
          <div className={slots.title()}>
            {name}
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
          <Card variant="ghost" size={2} className={slots.editorCard()}>
            {description}
          </Card>
          {engagement}
        </div>
        {properties}
      </div>
    </div>
  );
}
