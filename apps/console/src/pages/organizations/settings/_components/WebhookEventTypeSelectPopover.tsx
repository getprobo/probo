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

import { MagnifyingGlassIcon } from "@phosphor-icons/react";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
import { type ReactElement, useState } from "react";
import { useTranslation } from "react-i18next";

import {
  filterWebhookEventTypes,
  type WebhookEventTypeValue,
} from "../_lib/webhookEventTypes";
import { webhookEventTypeSelect } from "../variants";

interface WebhookEventTypeSelectPopoverProps {
  selectedEvents: readonly string[];
  disabled?: boolean;
  children: ReactElement;
  onToggle: (event: WebhookEventTypeValue) => void;
}

export function WebhookEventTypeSelectPopover({
  selectedEvents,
  disabled = false,
  children,
  onToggle,
}: WebhookEventTypeSelectPopoverProps) {
  const { t } = useTranslation();
  const { popup, results, row, label, empty } = webhookEventTypeSelect();
  const [query, setQuery] = useState("");
  const visibleEvents = filterWebhookEventTypes(query);

  function handleOpenChange(open: boolean) {
    if (!open) {
      setQuery("");
    }
  }

  return (
    <Popover onOpenChange={handleOpenChange}>
      <PopoverTrigger render={children} disabled={disabled} />
      <PopoverPopup
        side="bottom"
        align="start"
        className={popup()}
        aria-label={t("webhooksSettingsPage.editEvents")}
      >
        <TextField
          size={1}
          icon={<MagnifyingGlassIcon />}
          value={query}
          onValueChange={setQuery}
          placeholder={t("webhooksSettingsPage.searchEvents")}
          aria-label={t("webhooksSettingsPage.searchEvents")}
        />
        <div className={results()}>
          {visibleEvents.length === 0
            ? (
                <p className={empty()}>{t("webhooksSettingsPage.emptySearch")}</p>
              )
            : visibleEvents.map((event) => {
                const checked = selectedEvents.includes(event.value);
                const isLastSelected = checked && selectedEvents.length === 1;

                return (
                  <label key={event.value} className={row()}>
                    <Checkbox
                      checked={checked}
                      disabled={disabled || isLastSelected}
                      onCheckedChange={() => {
                        if (!isLastSelected) {
                          onToggle(event.value);
                        }
                      }}
                    />
                    <span className={label()}>{event.label}</span>
                  </label>
                );
              })}
        </div>
      </PopoverPopup>
    </Popover>
  );
}
