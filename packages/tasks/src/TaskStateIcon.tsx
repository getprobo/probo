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
  CheckCircleIcon,
  CircleDashedIcon,
  CircleHalfIcon,
  CircleIcon,
  ProhibitInsetIcon,
  XCircleIcon,
} from "@phosphor-icons/react";
import { tv } from "tailwind-variants/lite";

import type { TaskState } from "./taskState";

const taskStateIcon = tv({
  base: "shrink-0",
  variants: {
    size: {
      1: "size-4",
      2: "size-5",
    },
    state: {
      BACKLOG: "text-sand-10",
      TODO: "text-sand-11",
      IN_PROGRESS: "text-amber-11",
      DONE: "text-green-11",
      CANCELED: "text-red-11",
      DUPLICATE: "text-sand-10",
    },
  },
  defaultVariants: {
    size: 1,
  },
});

export interface TaskStateIconProps {
  state: TaskState;
  size?: 1 | 2;
}

export function TaskStateIcon({ state, size = 1 }: TaskStateIconProps) {
  const className = taskStateIcon({ state, size });

  switch (state) {
    case "BACKLOG":
      return <CircleDashedIcon className={className} weight="bold" aria-hidden />;
    case "TODO":
      return <CircleIcon className={className} weight="bold" aria-hidden />;
    case "IN_PROGRESS":
      return <CircleHalfIcon className={className} weight="fill" aria-hidden />;
    case "DONE":
      return <CheckCircleIcon className={className} weight="fill" aria-hidden />;
    case "CANCELED":
      return <XCircleIcon className={className} weight="bold" aria-hidden />;
    case "DUPLICATE":
      return <ProhibitInsetIcon className={className} weight="bold" aria-hidden />;
  }
}
