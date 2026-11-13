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

import { ShieldWarningIcon } from "@phosphor-icons/react";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { FindingPropertyMenu } from "./FindingPropertyMenu";
import { FindingRiskCombobox, type FindingRiskOption } from "./FindingRiskCombobox";
import { findingRisk } from "./variants";

export interface FindingPropertiesRiskProps {
  risk: FindingRiskOption | null;
  disabled?: boolean;
  error?: string | null;
  onSelect: (risk: FindingRiskOption) => void;
}

export function FindingPropertiesRisk({
  risk,
  disabled = false,
  error,
  onSelect,
}: FindingPropertiesRiskProps) {
  const { t } = useTranslation("organizations/findings");
  const [editing, setEditing] = useState(risk == null);
  const { root, field, icon } = findingRisk();
  const label = risk?.name ?? t("details.placeholders.risk");

  return (
    <div className={root()}>
      {editing
        ? (
            <div className={field()}>
              <FindingRiskCombobox
                selected={risk}
                disabled={disabled}
                onSelect={(next) => {
                  if (next == null) {
                    return;
                  }
                  onSelect(next);
                  setEditing(false);
                }}
              />
            </div>
          )
        : (
            <FindingPropertyMenu
              label={label}
              icon={<ShieldWarningIcon className={icon()} weight="fill" aria-hidden />}
              canUpdate
              disabled={disabled}
              faint={risk == null}
              menu={null}
              onOpen={() => {
                setEditing(true);
              }}
            />
          )}
      {error != null && <Text size={1} color="red">{error}</Text>}
    </div>
  );
}
