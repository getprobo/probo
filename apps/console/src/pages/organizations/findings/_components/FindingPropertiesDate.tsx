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

import { CalendarCheckIcon } from "@phosphor-icons/react";
import { taskListItem, TaskListItemDeadline } from "@probo/tasks";

import { findingIdentifiedIcon } from "./variants";

export interface FindingPropertiesDateProps {
  value?: string | null;
  placeholder: string;
  // Date identified uses a check calendar. Due date keeps the deadline calendar.
  identified?: boolean;
  overdue?: boolean;
  canUpdate: boolean;
  disabled?: boolean;
  onChange: (value: string | null) => void;
}

export function FindingPropertiesDate({
  value,
  placeholder,
  identified = false,
  overdue = false,
  canUpdate,
  disabled = false,
  onChange,
}: FindingPropertiesDateProps) {
  const slots = taskListItem({ field: true });

  return (
    <TaskListItemDeadline
      deadline={value}
      isOverdue={identified ? false : overdue}
      canUpdate={canUpdate}
      isUpdating={disabled}
      size={2}
      placeholder={placeholder}
      slots={slots}
      icon={identified
        ? <CalendarCheckIcon className={findingIdentifiedIcon()} weight="bold" aria-hidden />
        : undefined}
      onDeadlineChange={onChange}
    />
  );
}
