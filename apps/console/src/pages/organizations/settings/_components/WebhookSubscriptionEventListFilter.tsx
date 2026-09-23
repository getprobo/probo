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

import {
  isWebhookEventListStatusOption,
  useWebhookEventListFilters,
  type WebhookEventListStatusOption,
  webhookEventListStatusOptions,
} from "../_lib/useWebhookEventListFilters";
import { webhookSubscriptionEventList } from "../variants";

export function WebhookSubscriptionEventListFilter() {
  const { t } = useTranslation();
  const { status, setStatus } = useWebhookEventListFilters();
  const { filter } = webhookSubscriptionEventList();
  const allLabel = t("webhooksSettingsPage.filterAll");

  function optionLabel(option: WebhookEventListStatusOption): string {
    switch (option) {
      case "pending":
        return t("webhooksSettingsPage.status.pending");
      case "succeeded":
        return t("webhooksSettingsPage.status.succeeded");
      case "failed":
        return t("webhooksSettingsPage.status.failed");
    }
  }

  return (
    <div className={filter()}>
      <Select
        value={status === "all" ? null : status}
        onValueChange={(value: string | null) => {
          if (value == null) {
            setStatus("all");
            return;
          }
          if (isWebhookEventListStatusOption(value)) {
            setStatus(value);
          }
        }}
      >
        <SelectTrigger
          size={2}
          placeholder={allLabel}
          aria-label={t("webhooksSettingsPage.filterStatus")}
        >
          {(value: WebhookEventListStatusOption | null) => (
            value != null ? optionLabel(value) : allLabel
          )}
        </SelectTrigger>
        <SelectPopup align="end">
          <SelectItem value={null}>{allLabel}</SelectItem>
          {webhookEventListStatusOptions.map(option => (
            <SelectItem key={option} value={option}>
              {optionLabel(option)}
            </SelectItem>
          ))}
        </SelectPopup>
      </Select>
    </div>
  );
}
