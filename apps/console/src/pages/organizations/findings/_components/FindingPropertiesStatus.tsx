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
  CircleHalfIcon,
  CircleIcon,
  ProhibitInsetIcon,
  ShieldWarningIcon,
} from "@phosphor-icons/react";
import { DropdownRadioGroup } from "@probo/ui/src/v2/Dropdown/DropdownRadioGroup";
import { DropdownRadioItem } from "@probo/ui/src/v2/Dropdown/DropdownRadioItem";
import { useTranslation } from "react-i18next";

import {
  findingChoice,
  type FindingStatus,
  findingStatuses,
  isFindingStatus,
} from "../_lib/findingOptions";

import { FindingPropertyMenu } from "./FindingPropertyMenu";
import { FindingPropertyOption } from "./FindingPropertyOption";
import { findingStatusIcon } from "./variants";

export interface FindingPropertiesStatusProps {
  status: FindingStatus;
  canUpdate: boolean;
  disabled?: boolean;
  onChange: (status: FindingStatus) => void;
}

export function FindingPropertiesStatus({
  status,
  canUpdate,
  disabled = false,
  onChange,
}: FindingPropertiesStatusProps) {
  const { t } = useTranslation("organizations/findings");
  const label = t(`list.statuses.${status.toLowerCase()}`);

  return (
    <FindingPropertyMenu
      label={label}
      icon={<StatusIcon status={status} />}
      canUpdate={canUpdate}
      disabled={disabled}
      menu={close => (
        <DropdownRadioGroup
          value={status}
          onValueChange={(next) => {
            const value = findingChoice(next, findingStatuses);
            if (value != null) {
              onChange(value);
              close();
            }
          }}
        >
          {findingStatuses.map(item => (
            <DropdownRadioItem key={item} value={item}>
              <FindingPropertyOption
                icon={<StatusIcon status={item} />}
                label={t(`list.statuses.${item.toLowerCase()}`)}
              />
            </DropdownRadioItem>
          ))}
        </DropdownRadioGroup>
      )}
    />
  );
}

function StatusIcon({ status }: { status: string }) {
  const className = findingStatusIcon({
    status: isFindingStatus(status) ? status : "FALSE_POSITIVE",
  });
  switch (status) {
    case "OPEN":
      return <CircleIcon className={className} weight="bold" aria-hidden />;
    case "IN_PROGRESS":
      return <CircleHalfIcon className={className} weight="fill" aria-hidden />;
    case "CLOSED":
    case "MITIGATED":
      return <CheckCircleIcon className={className} weight="fill" aria-hidden />;
    case "RISK_ACCEPTED":
      return <ShieldWarningIcon className={className} weight="fill" aria-hidden />;
    default:
      return <ProhibitInsetIcon className={className} weight="bold" aria-hidden />;
  }
}
