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

import { WarningDiamondIcon } from "@phosphor-icons/react";
import { tv } from "tailwind-variants/lite";

import type { TaskPriority } from "./taskState";

const taskPriorityIcon = tv({
  slots: {
    root: "flex size-4 shrink-0 items-end justify-center gap-px p-px",
    bar: "w-[3px] rounded-full",
    urgent: "size-4 shrink-0 text-red-11",
  },
  variants: {
    tone: {
      idle: { bar: "bg-sand-6" },
      low: { bar: "bg-sand-10" },
      medium: { bar: "bg-sand-11" },
      high: { bar: "bg-sand-12" },
    },
  },
});

const barHeights = ["h-1", "h-2", "h-3"] as const;

const activeTone = {
  LOW: "low",
  MEDIUM: "medium",
  HIGH: "high",
} as const;

const activeBars = {
  LOW: 1,
  MEDIUM: 2,
  HIGH: 3,
} as const;

export interface TaskPriorityIconProps {
  priority: TaskPriority;
}

export function TaskPriorityIcon({ priority }: TaskPriorityIconProps) {
  const slots = taskPriorityIcon();
  if (priority === "URGENT") {
    return <WarningDiamondIcon className={slots.urgent()} weight="fill" aria-hidden />;
  }

  const filled = activeBars[priority];
  const tone = activeTone[priority];

  return (
    <span className={slots.root()} aria-hidden>
      {barHeights.map((height, index) => (
        <span
          key={height}
          className={`${taskPriorityIcon({ tone: index < filled ? tone : "idle" }).bar()} ${height}`}
        />
      ))}
    </span>
  );
}
