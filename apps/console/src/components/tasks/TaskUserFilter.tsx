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

import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";

import { usePeople } from "#/hooks/graph/PeopleGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { tasksCard } from "./variants";

interface TaskUserFilterProps {
  value: string | null;
  onValueChange: (value: string | null) => void;
}

interface PersonOption {
  id: string;
  fullName: string;
  emailAddress: string;
  avatarUrl?: string;
}

export function TaskUserFilter({ value, onValueChange }: TaskUserFilterProps) {
  return (
    <Suspense fallback={<TaskUserFilterSelect value={value} disabled people={[]} onValueChange={onValueChange} />}>
      <TaskUserFilterLoaded value={value} onValueChange={onValueChange} />
    </Suspense>
  );
}

function TaskUserFilterLoaded({ value, onValueChange }: TaskUserFilterProps) {
  const organizationId = useOrganizationId();
  const people = usePeople(organizationId);

  return (
    <TaskUserFilterSelect
      value={value}
      onValueChange={onValueChange}
      people={people.map(person => ({
        id: person.id,
        fullName: person.fullName,
        emailAddress: person.emailAddress,
        avatarUrl: person.avatar?.downloadUrl ?? undefined,
      }))}
    />
  );
}

interface TaskUserFilterSelectProps extends TaskUserFilterProps {
  disabled?: boolean;
  people: PersonOption[];
}

function TaskUserFilterSelect({
  value,
  onValueChange,
  disabled,
  people,
}: TaskUserFilterSelectProps) {
  const { t } = useTranslation();
  const { userOption } = tasksCard();
  const names = new Map(people.map(person => [person.id, person]));
  const selected = value == null ? undefined : names.get(value);

  return (
    <Select
      value={value}
      disabled={disabled}
      onValueChange={onValueChange}
    >
      <SelectTrigger
        size={2}
        placeholder={t("tasksCard.filters.allUsers")}
        aria-label={t("tasksCard.filters.user")}
      >
        {(selectedId: string | null) => {
          if (selectedId == null) {
            return t("tasksCard.filters.allUsers");
          }
          const person = names.get(selectedId) ?? selected;
          if (person == null) {
            return selectedId;
          }
          return (
            <span className={userOption()}>
              <Avatar
                size={1}
                radius="full"
                name={person.fullName}
                email={person.emailAddress}
                src={person.avatarUrl}
              />
              <span className="truncate">{person.fullName}</span>
            </span>
          );
        }}
      </SelectTrigger>
      <SelectPopup align="start">
        <SelectItem value={null}>
          {t("tasksCard.filters.allUsers")}
        </SelectItem>
        {value != null && selected == null && (
          <SelectItem value={value}>{value}</SelectItem>
        )}
        {people.map(person => (
          <SelectItem key={person.id} value={person.id}>
            <span className={userOption()}>
              <Avatar
                size={1}
                radius="full"
                name={person.fullName}
                email={person.emailAddress}
                src={person.avatarUrl}
              />
              <span className="truncate">{person.fullName}</span>
            </span>
          </SelectItem>
        ))}
      </SelectPopup>
    </Select>
  );
}
