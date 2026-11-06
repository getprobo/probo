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

import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { type ReactNode, Suspense } from "react";

import { UserSelectOption } from "#/components/form/UserSelectOption";
import { usePeople } from "#/hooks/graph/PeopleGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

interface FindingOwnerSelectProps {
  value: string | null;
  onChange: (value: string | null) => void;
  emptyLabel: string;
  ariaLabel: string;
  disabled?: boolean;
  align?: "start" | "end";
  // Icon shown beside the empty choice. The owner filter leaves this unset.
  emptyIcon?: ReactNode;
}

export function FindingOwnerSelect(props: FindingOwnerSelectProps) {
  const { emptyLabel, ariaLabel } = props;

  return (
    <Suspense
      fallback={(
        <Select value={null} disabled>
          <SelectTrigger size={2} placeholder={emptyLabel} aria-label={ariaLabel} />
        </Select>
      )}
    >
      <FindingOwnerSelectOptions {...props} />
    </Suspense>
  );
}

function FindingOwnerSelectOptions({
  value,
  onChange,
  emptyLabel,
  ariaLabel,
  disabled,
  align = "start",
  emptyIcon,
}: FindingOwnerSelectProps) {
  const organizationId = useOrganizationId();
  const loaded = usePeople(organizationId, { contractEnded: false });
  const people = loaded.flatMap(person => (person == null ? [] : [person]));
  const known = people.some(person => person.id === value);
  const options = known || value == null
    ? people
    : [...people, { id: value, fullName: value, emailAddress: null, avatar: null }];

  function option(person: (typeof options)[number]) {
    return (
      <UserSelectOption
        fullName={person.fullName}
        emailAddress={person.emailAddress}
        avatarUrl={person.avatar?.downloadUrl}
        size={2}
      />
    );
  }

  return (
    <Select<string | null> value={value} disabled={disabled} onValueChange={onChange}>
      <SelectTrigger size={2} placeholder={emptyLabel} aria-label={ariaLabel}>
        {(selected: string | null) => {
          if (selected == null) {
            return emptyIcon == null
              ? emptyLabel
              : (
                  <UserSelectOption
                    fullName={emptyLabel}
                    size={2}
                    fallback={emptyIcon}
                  />
                );
          }
          const person = options.find(candidate => candidate.id === selected);
          return person == null ? selected : option(person);
        }}
      </SelectTrigger>
      <SelectPopup align={align}>
        <SelectItem value={null}>
          {emptyIcon == null
            ? emptyLabel
            : (
                <UserSelectOption
                  fullName={emptyLabel}
                  size={2}
                  fallback={emptyIcon}
                />
              )}
        </SelectItem>
        {options.map(person => (
          <SelectItem key={person.id} value={person.id}>
            {option(person)}
          </SelectItem>
        ))}
      </SelectPopup>
    </Select>
  );
}
