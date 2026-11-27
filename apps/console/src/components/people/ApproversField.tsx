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

import { PlusIcon, XIcon } from "@phosphor-icons/react";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { PeopleMenu } from "./PeopleMenu";
import { approversField } from "./variants";

export interface Approver {
  id: string;
  fullName: string;
  emailAddress: string;
  avatarUrl?: string | null;
}

interface ApproversFieldProps {
  approvers: Approver[];
  onChange: (approvers: Approver[]) => void;
  label: string;
  description?: string;
  disabled?: boolean;
}

export function ApproversField({
  approvers,
  onChange,
  label,
  description,
  disabled,
}: ApproversFieldProps) {
  const { t } = useTranslation();
  const { root, rows, row, person, name, add } = approversField();

  return (
    <div className={root()}>
      <Text size={2} weight="medium" highContrast>{label}</Text>
      {description && <Text size={1}>{description}</Text>}
      {approvers.length > 0 && (
        <ul className={rows()}>
          {approvers.map(approver => (
            <li key={approver.id} className={row()}>
              <div className={person()}>
                <Avatar
                  size={1}
                  radius="full"
                  name={approver.fullName}
                  email={approver.emailAddress}
                  src={approver.avatarUrl}
                  alt=""
                />
                <Text size={2} className={name()} title={approver.emailAddress}>
                  {approver.fullName}
                </Text>
              </div>
              <IconButton
                size={1}
                variant="ghost"
                color="neutral"
                disabled={disabled}
                aria-label={t("approversField.remove", { name: approver.fullName })}
                onClick={() => {
                  onChange(approvers.filter(item => item.id !== approver.id));
                }}
              >
                <XIcon />
              </IconButton>
            </li>
          ))}
        </ul>
      )}
      <Dropdown>
        <DropdownTrigger
          disabled={disabled}
          render={(
            <Button
              size={1}
              variant="soft"
              color="neutral"
              className={add()}
              iconStart={<PlusIcon />}
            />
          )}
        >
          {t("approversField.add")}
        </DropdownTrigger>
        <DropdownPopup>
          <PeopleMenu
            value={null}
            emptyValue={null}
            disabled={disabled}
            onValueChange={() => {}}
            onSelect={(selected) => {
              if (selected == null) {
                return;
              }
              onChange(approvers.some(item => item.id === selected.id)
                ? approvers
                : [...approvers, {
                    id: selected.id,
                    fullName: selected.fullName,
                    emailAddress: selected.emailAddress,
                    avatarUrl: selected.avatarUrl,
                  }]);
            }}
          />
        </DropdownPopup>
      </Dropdown>
    </div>
  );
}
