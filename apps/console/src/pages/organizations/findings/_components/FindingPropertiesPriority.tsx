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

import { TaskPriorityIcon } from "@probo/tasks";
import { DropdownRadioGroup } from "@probo/ui/src/v2/Dropdown/DropdownRadioGroup";
import { DropdownRadioItem } from "@probo/ui/src/v2/Dropdown/DropdownRadioItem";
import { useTranslation } from "react-i18next";

import {
  findingChoice,
  findingPriorities,
  type FindingPriority,
} from "../_lib/findingOptions";

import { FindingPropertyMenu } from "./FindingPropertyMenu";
import { FindingPropertyOption } from "./FindingPropertyOption";

export interface FindingPropertiesPriorityProps {
  priority: FindingPriority;
  canUpdate: boolean;
  disabled?: boolean;
  onChange: (priority: FindingPriority) => void;
}

export function FindingPropertiesPriority({
  priority,
  canUpdate,
  disabled = false,
  onChange,
}: FindingPropertiesPriorityProps) {
  const { t } = useTranslation("organizations/findings");
  const label = t(`list.priorities.${priority.toLowerCase()}`);

  return (
    <FindingPropertyMenu
      label={label}
      icon={<TaskPriorityIcon priority={priority} />}
      canUpdate={canUpdate}
      disabled={disabled}
      menu={close => (
        <DropdownRadioGroup
          value={priority}
          onValueChange={(next) => {
            const value = findingChoice(next, findingPriorities);
            if (value != null) {
              onChange(value);
              close();
            }
          }}
        >
          {findingPriorities.map(item => (
            <DropdownRadioItem key={item} value={item}>
              <FindingPropertyOption
                icon={<TaskPriorityIcon priority={item} />}
                label={t(`list.priorities.${item.toLowerCase()}`)}
              />
            </DropdownRadioItem>
          ))}
        </DropdownRadioGroup>
      )}
    />
  );
}
