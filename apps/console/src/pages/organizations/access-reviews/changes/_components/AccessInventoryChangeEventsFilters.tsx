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

import { DateField } from "@probo/ui/src/v2/form/DateField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useTranslation } from "react-i18next";

import { accessReviewSourceLabel } from "../../_lib/sourceLabel";
import {
  type AccessInventoryChangeTypeFilter,
  accessInventoryChangeTypes,
  useAccessInventoryChangeEventsFilters,
} from "../_lib/useAccessInventoryChangeEventsFilters";

import { accessInventoryChangeEventsList } from "./variants";

export interface AccessInventoryChangeEventsFilterSource {
  id: string;
  name: string;
  accountName?: string | null;
}

interface AccessInventoryChangeEventsFiltersProps {
  sources: readonly AccessInventoryChangeEventsFilterSource[];
}

export function AccessInventoryChangeEventsFilters({
  sources,
}: AccessInventoryChangeEventsFiltersProps) {
  const { i18n, t } = useTranslation("organizations/access-reviews");
  const {
    sourceId,
    changeType,
    observedAfter,
    observedBefore,
    setSourceId,
    setChangeType,
    setObservedAfter,
    setObservedBefore,
  } = useAccessInventoryChangeEventsFilters();
  const { tools, filters, filter, dateFilter } = accessInventoryChangeEventsList();
  const allSourcesLabel = t("changes.filters.allSources");
  const allChangeTypesLabel = t("changes.filters.allChangeTypes");
  const sourceById = new Map(
    sources.map(source => [
      source.id,
      accessReviewSourceLabel({
        name: source.name,
        accountName: source.accountName,
      }),
    ]),
  );

  return (
    <div className={tools()}>
      <div className={filters()}>
        <div className={filter()}>
          <Select
            value={sourceId}
            onValueChange={(value: string | null) => {
              setSourceId(value);
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allSourcesLabel}
              aria-label={t("changes.filters.source")}
            >
              {(value: string | null) => (
                value != null
                  ? sourceById.get(value) ?? allSourcesLabel
                  : allSourcesLabel
              )}
            </SelectTrigger>
            <SelectPopup align="end">
              <SelectItem value={null}>{allSourcesLabel}</SelectItem>
              {sources.map(source => (
                <SelectItem key={source.id} value={source.id}>
                  {accessReviewSourceLabel({
                    name: source.name,
                    accountName: source.accountName,
                  })}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
        <div className={filter()}>
          <Select
            value={changeType}
            onValueChange={(value: string | null) => {
              if (value == null) {
                setChangeType(null);
                return;
              }
              if ((accessInventoryChangeTypes as readonly string[]).includes(value)) {
                setChangeType(value as AccessInventoryChangeTypeFilter);
              }
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allChangeTypesLabel}
              aria-label={t("changes.filters.changeType")}
            >
              {(value: AccessInventoryChangeTypeFilter | null) => (
                value != null
                  ? t(`changes.changeTypes.${value}`)
                  : allChangeTypesLabel
              )}
            </SelectTrigger>
            <SelectPopup align="end">
              <SelectItem value={null}>{allChangeTypesLabel}</SelectItem>
              {accessInventoryChangeTypes.map(option => (
                <SelectItem key={option} value={option}>
                  {t(`changes.changeTypes.${option}`)}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
        <div className={dateFilter()}>
          <DateField
            size={2}
            nullable
            value={observedAfter ?? ""}
            locale={i18n.language}
            max={observedBefore ?? undefined}
            aria-label={t("changes.filters.observedAfter")}
            placeholder={t("changes.filters.observedAfter")}
            onValueChange={(value) => {
              setObservedAfter(value === "" ? null : value);
            }}
          />
        </div>
        <div className={dateFilter()}>
          <DateField
            size={2}
            nullable
            value={observedBefore ?? ""}
            locale={i18n.language}
            min={observedAfter ?? undefined}
            aria-label={t("changes.filters.observedBefore")}
            placeholder={t("changes.filters.observedBefore")}
            onValueChange={(value) => {
              setObservedBefore(value === "" ? null : value);
            }}
          />
        </div>
      </div>
    </div>
  );
}
