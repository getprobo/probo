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

import {
  CellSignalFullIcon,
  CellSignalLowIcon,
  CellSignalMediumIcon,
  WarningCircleIcon,
} from "@phosphor-icons/react";
import { tv } from "tailwind-variants/lite";

import type { TaskPriority } from "../_lib/taskState";

const taskPriorityIcon = tv({
  base: "size-4 shrink-0",
  variants: {
    priority: {
      LOW: "text-sand-10",
      MEDIUM: "text-sand-11",
      HIGH: "text-sand-12",
      URGENT: "text-red-11",
    },
  },
});

interface TaskPriorityIconProps {
  priority: TaskPriority;
}

export function TaskPriorityIcon({ priority }: TaskPriorityIconProps) {
  const className = taskPriorityIcon({ priority });

  switch (priority) {
    case "LOW":
      return <CellSignalLowIcon className={className} aria-hidden />;
    case "MEDIUM":
      return <CellSignalMediumIcon className={className} aria-hidden />;
    case "HIGH":
      return <CellSignalFullIcon className={className} aria-hidden />;
    case "URGENT":
      return <WarningCircleIcon className={className} weight="fill" aria-hidden />;
  }
}
