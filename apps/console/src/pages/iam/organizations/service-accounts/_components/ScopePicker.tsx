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

import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { formatAPIScopeLabel } from "#/pages/iam/oauthTokens/_components/scopeLabels";

interface ScopePickerProps {
  disabled?: boolean;
  scopes: readonly string[];
  selectedScopes: readonly string[];
  onChange: (scopes: string[]) => void;
}

export function ScopePicker(props: ScopePickerProps) {
  const {
    disabled = false,
    scopes,
    selectedScopes,
    onChange,
  } = props;
  const { t } = useTranslation("iam/organizations/service-accounts");
  const selected = new Set(selectedScopes);

  function toggleScope(scope: string, checked: boolean) {
    if (checked) {
      selected.add(scope);
    } else {
      selected.delete(scope);
    }
    onChange([...selected]);
  }

  return (
    <fieldset className="space-y-3" disabled={disabled}>
      <div className="flex items-center justify-between gap-4">
        <legend className="text-2 font-medium text-sand-12">
          {t("fields.scopes")}
        </legend>
        <button
          type="button"
          className="text-2 font-medium text-gold-11 hover:text-gold-12"
          onClick={() =>
            onChange(selectedScopes.length === scopes.length ? [] : [...scopes])}
        >
          {selectedScopes.length === scopes.length
            ? t("actions.deselectAll")
            : t("actions.selectAll")}
        </button>
      </div>
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
        {scopes.map(scope => (
          <label
            key={scope}
            className="flex cursor-pointer items-start gap-2 rounded-3 border border-sand-6 bg-sand-2 p-3"
          >
            <Checkbox
              checked={selected.has(scope)}
              onCheckedChange={checked => toggleScope(scope, checked)}
            />
            <span className="min-w-0">
              <Text size={2} weight="medium" highContrast>
                {formatAPIScopeLabel(scope, t)}
              </Text>
              <Text
                className="block break-all"
                size={1}
                color="faint"
              >
                {scope}
              </Text>
            </span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}
