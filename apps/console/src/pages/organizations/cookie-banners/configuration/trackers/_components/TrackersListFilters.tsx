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
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { TrackersListFilters_cookieBanner$key } from "#/__generated__/core/TrackersListFilters_cookieBanner.graphql";

import { cookieBannerList } from "../../../variants";
import {
  type CookieSource,
  cookieSources,
  isCookieSource,
  isTrackerType,
  type TrackerType,
  trackerTypes,
  useTrackersListFilters,
} from "../_lib/useTrackersListFilters";

import { TrackersListSearch } from "./TrackersListSearch";

const trackersListFiltersFragment = graphql`
  fragment TrackersListFilters_cookieBanner on CookieBanner {
    linkedThirdParties {
      id
      name
    }
    categories(first: 50, orderBy: { field: RANK, direction: ASC })
      @required(action: THROW) {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`;

const sourceLabels = {
  SCRIPT: "script",
  PRE_EXISTING: "preExisting",
  HTTP: "http",
  EXTENSION: "extension",
} as const;

const typeLabels = {
  COOKIE: "cookie",
  LOCAL_STORAGE: "localStorage",
  SESSION_STORAGE: "sessionStorage",
  INDEXED_DB: "indexedDb",
  CACHE_STORAGE: "cacheStorage",
} as const;

interface TrackersListFiltersProps {
  cookieBannerKey: TrackersListFilters_cookieBanner$key;
}

export function TrackersListFilters({ cookieBannerKey }: TrackersListFiltersProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const cookieBanner = useFragment(trackersListFiltersFragment, cookieBannerKey);
  const {
    source,
    type,
    category,
    party,
    setSource,
    setType,
    setCategory,
    setParty,
  } = useTrackersListFilters();
  const { tools, filters, filter } = cookieBannerList();
  const allSourcesLabel = t("trackersPage.sources.all");
  const allTypesLabel = t("trackersPage.types.all");
  const allCategoriesLabel = t("trackersPage.filters.allCategories");
  const allThirdPartiesLabel = t("trackersPage.filters.allThirdParties");
  const categories = cookieBanner.categories.edges.map(edge => edge.node);

  return (
    <div className={tools()}>
      <TrackersListSearch />
      <div className={filters()}>
        <div className={filter()}>
          <Select
            value={party}
            onValueChange={(value: string | null) => {
              setParty(value);
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allThirdPartiesLabel}
              aria-label={t("trackersPage.filters.thirdParty")}
            >
              {(value: string | null) => (
                value != null
                  ? cookieBanner.linkedThirdParties.find(item => item.id === value)?.name
                  ?? allThirdPartiesLabel
                  : allThirdPartiesLabel
              )}
            </SelectTrigger>
            <SelectPopup align="start">
              <SelectItem value={null}>{allThirdPartiesLabel}</SelectItem>
              {cookieBanner.linkedThirdParties.map(item => (
                <SelectItem key={item.id} value={item.id}>
                  {item.name}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
        <div className={filter()}>
          <Select
            value={type}
            onValueChange={(value: string | null) => {
              if (value == null) {
                setType(null);
                return;
              }
              if (isTrackerType(value)) {
                setType(value);
              }
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allTypesLabel}
              aria-label={t("trackersPage.filters.type")}
            >
              {(value: TrackerType | null) => (
                value != null ? t(`trackersPage.types.${typeLabels[value]}`) : allTypesLabel
              )}
            </SelectTrigger>
            <SelectPopup align="start">
              <SelectItem value={null}>{allTypesLabel}</SelectItem>
              {trackerTypes.map(trackerType => (
                <SelectItem key={trackerType} value={trackerType}>
                  {t(`trackersPage.types.${typeLabels[trackerType]}`)}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
        <div className={filter()}>
          <Select
            value={source}
            onValueChange={(value: string | null) => {
              if (value == null) {
                setSource(null);
                return;
              }
              if (isCookieSource(value)) {
                setSource(value);
              }
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allSourcesLabel}
              aria-label={t("trackersPage.filters.source")}
            >
              {(value: CookieSource | null) => (
                value != null ? t(`trackersPage.sources.${sourceLabels[value]}`) : allSourcesLabel
              )}
            </SelectTrigger>
            <SelectPopup align="start">
              <SelectItem value={null}>{allSourcesLabel}</SelectItem>
              {cookieSources.map(cookieSource => (
                <SelectItem key={cookieSource} value={cookieSource}>
                  {t(`trackersPage.sources.${sourceLabels[cookieSource]}`)}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
        <div className={filter()}>
          <Select
            value={category}
            onValueChange={(value: string | null) => {
              setCategory(value);
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allCategoriesLabel}
              aria-label={t("trackersPage.filters.category")}
            >
              {(value: string | null) => (
                value != null
                  ? categories.find(item => item.id === value)?.name ?? allCategoriesLabel
                  : allCategoriesLabel
              )}
            </SelectTrigger>
            <SelectPopup align="start">
              <SelectItem value={null}>{allCategoriesLabel}</SelectItem>
              {categories.map(item => (
                <SelectItem key={item.id} value={item.id}>
                  {item.name}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
      </div>
    </div>
  );
}
