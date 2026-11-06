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

import { UserCircleDashedIcon } from "@phosphor-icons/react";
import { taskListItem, TaskListItemAssignee } from "@probo/tasks";
import { useTranslation } from "react-i18next";

import { PeopleMenu } from "#/components/people/PeopleMenu";

export interface FindingOwnerValue {
  id: string;
  fullName: string;
  emailAddress?: string | null;
  avatarUrl?: string | null;
}

export interface FindingPropertiesOwnerProps {
  owner: FindingOwnerValue | null;
  canUpdate: boolean;
  // Offers an "Unassigned" choice. A saved finding cannot drop its owner.
  clearable?: boolean;
  disabled?: boolean;
  onChange: (ownerId: string | null) => void;
  onSelect?: (owner: FindingOwnerValue | null) => void;
}

export function FindingPropertiesOwner({
  owner,
  canUpdate,
  clearable = false,
  disabled = false,
  onChange,
  onSelect,
}: FindingPropertiesOwnerProps) {
  const { t } = useTranslation("organizations/findings");
  const { stateOption } = taskListItem({ field: true });

  return (
    <TaskListItemAssignee
      assignee={owner == null
        ? null
        : {
            id: owner.id,
            fullName: owner.fullName,
            emailAddress: owner.emailAddress ?? "",
            avatarUrl: owner.avatarUrl,
          }}
      canUpdate={canUpdate}
      isUpdating={disabled}
      showLabel
      placeholder={t("details.placeholders.owner")}
      className={stateOption()}
      renderAssigneeMenu={canUpdate
        ? () => (
            <PeopleMenu
              value={owner?.id ?? null}
              emptyLabel={clearable ? t("list.values.unassigned") : undefined}
              emptyValue={null}
              emptyIcon={<UserCircleDashedIcon />}
              disabled={disabled}
              onValueChange={onChange}
              onSelect={onSelect == null
                ? undefined
                : (assignee) => {
                    onSelect(assignee == null
                      ? null
                      : {
                          id: assignee.id,
                          fullName: assignee.fullName,
                          emailAddress: assignee.emailAddress,
                          avatarUrl: assignee.avatarUrl,
                        });
                  }}
            />
          )
        : undefined}
    />
  );
}
