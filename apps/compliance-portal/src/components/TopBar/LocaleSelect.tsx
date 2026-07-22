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

import { GlobeIcon } from "@phosphor-icons/react";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import type { RefObject } from "react";
import { useTranslation } from "react-i18next";

import {
  URL_LOCALE_LABELS,
  URL_LOCALES,
  type UrlLocale,
} from "#/lib/i18n/locale";
import { useChangeLocale } from "#/lib/i18n/useChangeLocale";
import { useLocale } from "#/lib/i18n/useLocale";

interface LocaleSelectProps {
  // Persist the choice on the signed-in identity when true.
  persist?: boolean;
  // Called after a locale change is requested (e.g. close the mobile drawer).
  onLocaleChange?: () => void;
  // Portal target for the menu. Required inside a Drawer/Dialog so the popup
  // is not painted under the modal layer (Select defaults to body + z-3).
  portalContainer?: RefObject<HTMLElement | null>;
}

// Compact locale control for the top bar (guest and mobile). Uses the v2 Select
// (Figma Select / ghost + globe) rather than a custom dropdown.
export function LocaleSelect({
  persist = false,
  onLocaleChange,
  portalContainer,
}: LocaleSelectProps) {
  const { t } = useTranslation();
  const locale = useLocale();
  const [changeLocale, isChanging] = useChangeLocale();

  return (
    <Select
      value={locale}
      onValueChange={(value: UrlLocale | null) => {
        if (value == null || value === locale) {
          return;
        }
        onLocaleChange?.();
        void changeLocale(value, { persist });
      }}
      disabled={isChanging}
    >
      <SelectTrigger
        size={1}
        variant="ghost"
        aria-label={t("locale.label")}
        className="w-auto min-w-0"
      >
        {(value: UrlLocale | null) => (
          <span className="flex items-center gap-1.5">
            <GlobeIcon className="size-3.5 shrink-0" aria-hidden />
            {value ? URL_LOCALE_LABELS[value] : null}
          </span>
        )}
      </SelectTrigger>
      <SelectPopup container={portalContainer}>
        {URL_LOCALES.map(code => (
          <SelectItem key={code} value={code}>
            {URL_LOCALE_LABELS[code]}
          </SelectItem>
        ))}
      </SelectPopup>
    </Select>
  );
}
