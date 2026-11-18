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
import { useTranslation } from "react-i18next";

import type { DocumentMode } from "../_lib/useDocumentModes";

interface DocumentModeSelectProps {
  mode: DocumentMode;
  modes: readonly DocumentMode[];
  onModeChange: (mode: DocumentMode) => void;
}

export function DocumentModeSelect({
  mode,
  modes,
  onModeChange,
}: DocumentModeSelectProps) {
  const { t } = useTranslation("organizations/documents");

  return (
    <Select
      value={mode}
      disabled={modes.length < 2}
      onValueChange={(next: string | null) => {
        if (next != null && modes.includes(next as DocumentMode)) {
          onModeChange(next as DocumentMode);
        }
      }}
    >
      <SelectTrigger
        size={2}
        variant="ghost"
        className="w-auto shrink-0"
        aria-label={t("details.mode.label")}
      >
        {(selected: string | null) => selected == null ? null : t(`details.mode.${selected}`)}
      </SelectTrigger>
      <SelectPopup>
        {modes.map(item => (
          <SelectItem key={item} value={item}>
            {t(`details.mode.${item}`)}
          </SelectItem>
        ))}
      </SelectPopup>
    </Select>
  );
}
