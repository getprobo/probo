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

import { CaretDownIcon, UserCircleDashedIcon } from "@phosphor-icons/react";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { selectTrigger } from "@probo/ui/src/v2/Select/variants";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { TaskUserFilterProfileQuery } from "#/__generated__/core/TaskUserFilterProfileQuery.graphql";
import { UserSelectOption } from "#/components/form/UserSelectOption";
import { TaskAssigneeMenu } from "#/components/tasks/TaskAssigneeMenu";
import { unassignedAssignee } from "#/pages/organizations/tasks/_lib/taskPath";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

const taskUserFilterProfileQuery = graphql`
  query TaskUserFilterProfileQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on Profile {
        id
        fullName
        emailAddress
        avatar {
          downloadUrl
        }
      }
    }
  }
`;

interface TaskUserFilterProps {
  value: string | null;
  onValueChange: (value: string | null) => void;
}

export function TaskUserFilter({ value, onValueChange }: TaskUserFilterProps) {
  const { t } = useTranslation();
  const emptyLabel = t("tasksCard.filters.allUsers");
  const unassignedLabel = t("tasksCard.filters.notAssigned");
  const slots = selectTrigger({ size: 2 });
  const unassignedIcon = <UserCircleDashedIcon />;

  return (
    <Dropdown>
      <DropdownTrigger className={slots.trigger()}>
        <span className={slots.value()}>
          {value == null
            ? emptyLabel
            : value === unassignedAssignee
              ? (
                  <UserSelectOption
                    fullName={unassignedLabel}
                    size={2}
                    fallback={unassignedIcon}
                  />
                )
              : (
                  <Suspense fallback={<TextSkeleton size={2} className="w-24" />}>
                    <IAMRelayProvider>
                      <TaskUserFilterProfile id={value} />
                    </IAMRelayProvider>
                  </Suspense>
                )}
        </span>
        <span className={slots.icon()}>
          <CaretDownIcon />
        </span>
      </DropdownTrigger>
      <DropdownPopup size={2} align="start">
        <TaskAssigneeMenu
          value={value}
          emptyLabel={emptyLabel}
          emptyValue={null}
          hideEndedContracts={false}
          unassigned={{
            label: unassignedLabel,
            value: unassignedAssignee,
            icon: unassignedIcon,
          }}
          onValueChange={onValueChange}
        />
      </DropdownPopup>
    </Dropdown>
  );
}

function TaskUserFilterProfile({ id }: { id: string }) {
  const data = useLazyLoadQuery<TaskUserFilterProfileQuery>(
    taskUserFilterProfileQuery,
    { id },
    { fetchPolicy: "store-or-network" },
  );
  const profile = data.node?.__typename === "Profile" ? data.node : null;
  if (profile == null) {
    return id;
  }

  return (
    <UserSelectOption
      fullName={profile.fullName}
      emailAddress={profile.emailAddress}
      avatarUrl={profile.avatar?.downloadUrl}
      size={2}
    />
  );
}
