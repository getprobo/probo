// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { useCountryLabel } from "../_lib/useCountryLabel";
import { useSubprocessorFilters } from "../_lib/useSubprocessorFilters";
import { useSubprocessorSearch } from "../_lib/useSubprocessorSearch";

import type { SubprocessorsToolbar_query$key } from "./__generated__/SubprocessorsToolbar_query.graphql";

// Unfiltered facet data: the distinct categories and countries actually present,
// used to populate the filter dropdowns (aliased so it does not collide with the
// filtered list selection on the same trust center).
const subprocessorsToolbarFragment = graphql`
  fragment SubprocessorsToolbar_query on Query {
    currentTrustCenter @required(action: THROW) {
      allSubprocessors: subprocessors(first: 250) {
        edges {
          node {
            id
            category
            countries
          }
        }
      }
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

  const nodes = data.currentTrustCenter.allSubprocessors.edges.map(edge => edge.node);

  const categoryOptions = useMemo(() => {
    const present = [...new Set(nodes.map(node => node.category))];
    return present.sort((a, b) => t(`categories.${a}.label`).localeCompare(t(`categories.${b}.label`)));
  }, [nodes, t]);

  const countryOptions = useMemo(() => {
    const present = [...new Set(nodes.flatMap(node => node.countries))];
    return present.sort((a, b) => countryLabel(a).localeCompare(countryLabel(b)));
  }, [nodes, countryLabel]);

  return (
    <div className="flex flex-wrap items-center gap-3">
      <div className="w-40">
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
      <div className="w-40">
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
      <div className="w-60">
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
