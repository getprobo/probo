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
import { Switch } from "@probo/ui/src/v2/Switch/Switch";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { accessReviewSourceLabel } from "../../_lib/sourceLabel";
import {
  type AccessInventoryGrantPresenceFilter,
  accessInventoryGrantPresences,
  useAccessInventoryGrantsFilters,
} from "../_lib/useAccessInventoryGrantsFilters";

import { accessInventoryGrantsList } from "./variants";

export interface AccessInventoryGrantsFilterSource {
  id: string;
  name: string;
  accountName?: string | null;
}

interface AccessInventoryGrantsFiltersProps {
  sources: readonly AccessInventoryGrantsFilterSource[];
}

export function AccessInventoryGrantsFilters({
  sources,
}: AccessInventoryGrantsFiltersProps) {
  const { t } = useTranslation("organizations/access-reviews");
  const {
    sourceId,
    presence,
    includeDisappeared,
    setSourceId,
    setPresence,
    setIncludeDisappeared,
  } = useAccessInventoryGrantsFilters();
  const { tools, filters, filter, include } = accessInventoryGrantsList();
  const allSourcesLabel = t("grants.filters.allSources");
  const allPresenceLabel = t("grants.filters.allPresence");
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
      <div className={include()}>
        <Switch
          size={2}
          checked={includeDisappeared}
          color="gold"
          aria-label={t("grants.filters.includeDisappeared")}
          onCheckedChange={setIncludeDisappeared}
        />
        <Text size={2} color="neutral">
          {t("grants.filters.includeDisappeared")}
        </Text>
      </div>
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
              aria-label={t("grants.filters.source")}
            >
              {(value: string | null) => (
                value != null
                  ? sourceById.get(value) ?? value
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
            value={presence}
            onValueChange={(value: string | null) => {
              if (value == null) {
                setPresence(null);
                return;
              }
              if ((accessInventoryGrantPresences as readonly string[]).includes(value)) {
                setPresence(value as AccessInventoryGrantPresenceFilter);
              }
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allPresenceLabel}
              aria-label={t("grants.filters.presence")}
            >
              {(value: AccessInventoryGrantPresenceFilter | null) => (
                value != null
                  ? t(`grants.presence.${value.toLowerCase()}`)
                  : allPresenceLabel
              )}
            </SelectTrigger>
            <SelectPopup align="end">
              <SelectItem value={null}>{allPresenceLabel}</SelectItem>
              {accessInventoryGrantPresences.map(option => (
                <SelectItem key={option} value={option}>
                  {t(`grants.presence.${option.toLowerCase()}`)}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
      </div>
    </div>
  );
}
