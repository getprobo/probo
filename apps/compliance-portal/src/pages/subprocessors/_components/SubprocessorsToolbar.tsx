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

import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { useCountryLabel } from "../_lib/useCountryLabel";
import { useSubprocessorFilters } from "../_lib/useSubprocessorFilters";
import { useSubprocessorSearch } from "../_lib/useSubprocessorSearch";

import type { SubprocessorsToolbar_query$key } from "./__generated__/SubprocessorsToolbar_query.graphql";

// Facet data: the distinct categories and countries actually present across the
// trust center's published subprocessors, used to populate the filter dropdowns
// (server-computed so the options never dead-end on an empty result).
const subprocessorsToolbarFragment = graphql`
  fragment SubprocessorsToolbar_query on Query {
    currentTrustCenter @required(action: THROW) {
      subprocessorCategories
      subprocessorCountries
    }
  }
`;

interface SubprocessorsToolbarProps {
  queryKey: SubprocessorsToolbar_query$key;
}

// Category / region / search controls. Writes filter state to the URL via the
// filter hook; the page reacts to those changes and refetches server-side.
export function SubprocessorsToolbar({ queryKey }: SubprocessorsToolbarProps) {
  const { t } = useTranslation("subprocessors");
  const data = useFragment(subprocessorsToolbarFragment, queryKey);
  const countryLabel = useCountryLabel();
  const { category, country, setCategory, setCountry } = useSubprocessorFilters();
  const [queryInput, setQueryInput] = useSubprocessorSearch();

  const { subprocessorCategories, subprocessorCountries } = data.currentTrustCenter;

  const categoryOptions = useMemo(() => {
    return [...subprocessorCategories].sort((a, b) => t(`categories.${a}.label`).localeCompare(t(`categories.${b}.label`)));
  }, [subprocessorCategories, t]);

  const countryOptions = useMemo(() => {
    return [...subprocessorCountries].sort((a, b) => countryLabel(a).localeCompare(countryLabel(b)));
  }, [subprocessorCountries, countryLabel]);

  return (
    // flushBottomSpace drops the band's bottom padding so the desktop toolbar
    // sits on the header edge; restore that padding on small screens where the
    // controls stack into a column.
    <div className="flex min-h-16 flex-wrap items-center gap-3 max-sm:pb-8">
      <div className="w-40 max-sm:w-full">
        <Select value={category || null} onValueChange={value => setCategory(value ?? "")}>
          <SelectTrigger placeholder={t("filters.allCategories")}>
            {(value: string | null) => (value ? t(`categories.${value}.label`) : t("filters.allCategories"))}
          </SelectTrigger>
          <SelectPopup>
            <SelectItem value={null}>{t("filters.allCategories")}</SelectItem>
            {categoryOptions.map(option => (
              <SelectItem key={option} value={option}>{t(`categories.${option}.label`)}</SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </div>
      <div className="w-40 max-sm:w-full">
        <Select value={country || null} onValueChange={value => setCountry(value ?? "")}>
          <SelectTrigger placeholder={t("filters.allRegions")}>
            {(value: string | null) => (value ? countryLabel(value) : t("filters.allRegions"))}
          </SelectTrigger>
          <SelectPopup>
            <SelectItem value={null}>{t("filters.allRegions")}</SelectItem>
            {countryOptions.map(option => (
              <SelectItem key={option} value={option}>{countryLabel(option)}</SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </div>
      <div className="min-w-60 flex-1 max-sm:w-full max-sm:min-w-0">
        <TextField
          value={queryInput}
          onValueChange={setQueryInput}
          placeholder={t("filters.searchPlaceholder")}
          aria-label={t("filters.searchPlaceholder")}
        />
      </div>
    </div>
  );
}
