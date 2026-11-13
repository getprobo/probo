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

import { EyeIcon, MinusCircleIcon, WarningIcon, WarningOctagonIcon } from "@phosphor-icons/react";
import { DropdownRadioGroup } from "@probo/ui/src/v2/Dropdown/DropdownRadioGroup";
import { DropdownRadioItem } from "@probo/ui/src/v2/Dropdown/DropdownRadioItem";
import { useTranslation } from "react-i18next";

import {
  findingChoice,
  type FindingKind,
  findingKindKey,
  findingKinds,
  isFindingKind,
} from "../_lib/findingOptions";

import { FindingPropertyMenu } from "./FindingPropertyMenu";
import { FindingPropertyOption } from "./FindingPropertyOption";
import { findingKindIcon } from "./variants";

export interface FindingPropertiesKindProps {
  kind: string;
  canUpdate?: boolean;
  disabled?: boolean;
  onChange?: (kind: FindingKind) => void;
}

export function FindingPropertiesKind({
  kind,
  canUpdate = false,
  disabled = false,
  onChange,
}: FindingPropertiesKindProps) {
  const { t } = useTranslation("organizations/findings");
  const label = t(`list.kinds.${findingKindKey(kind)}`);

  return (
    <FindingPropertyMenu
      label={label}
      icon={<KindIcon kind={kind} />}
      canUpdate={canUpdate && onChange != null}
      disabled={disabled}
      menu={close => (
        <DropdownRadioGroup
          value={kind}
          onValueChange={(next) => {
            const value = findingChoice(next, findingKinds);
            if (value != null) {
              onChange?.(value);
              close();
            }
          }}
        >
          {findingKinds.map(item => (
            <DropdownRadioItem key={item} value={item}>
              <FindingPropertyOption
                icon={<KindIcon kind={item} />}
                label={t(`list.kinds.${findingKindKey(item)}`)}
              />
            </DropdownRadioItem>
          ))}
        </DropdownRadioGroup>
      )}
    />
  );
}

function KindIcon({ kind }: { kind: string }) {
  const className = findingKindIcon({ kind: isFindingKind(kind) ? kind : "EXCEPTION" });
  switch (kind) {
    case "MAJOR_NONCONFORMITY":
      return <WarningOctagonIcon className={className} weight="fill" aria-hidden />;
    case "MINOR_NONCONFORMITY":
      return <WarningIcon className={className} weight="fill" aria-hidden />;
    case "OBSERVATION":
      return <EyeIcon className={className} weight="bold" aria-hidden />;
    default:
      return <MinusCircleIcon className={className} weight="bold" aria-hidden />;
  }
}
