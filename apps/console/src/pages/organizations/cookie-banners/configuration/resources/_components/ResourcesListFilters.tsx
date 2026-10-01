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

import { cookieBannerList } from "../../../variants";
import {
  isTrackerResourceType,
  type TrackerResourceType,
  trackerResourceTypes,
  useResourcesListFilters,
} from "../_lib/useResourcesListFilters";

import { ResourcesListSearch } from "./ResourcesListSearch";

const typeLabels = {
  SCRIPT: "script",
  IFRAME: "iframe",
  IMAGE: "image",
  STYLESHEET: "stylesheet",
  FONT: "font",
  BEACON: "beacon",
  FETCH: "fetch",
  MEDIA: "media",
  SERVICE_WORKER: "serviceWorker",
} as const;

export function ResourcesListFilters() {
  const { t } = useTranslation("organizations/cookie-banners");
  const { type, setType } = useResourcesListFilters();
  const { tools, filters, filter } = cookieBannerList();
  const allTypesLabel = t("resourcesPage.types.all");

  return (
    <div className={tools()}>
      <ResourcesListSearch />
      <div className={filters()}>
        <div className={filter()}>
          <Select
            value={type}
            onValueChange={(value: string | null) => {
              if (value == null) {
                setType(null);
                return;
              }
              if (isTrackerResourceType(value)) {
                setType(value);
              }
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allTypesLabel}
              aria-label={t("resourcesPage.filters.type")}
            >
              {(value: TrackerResourceType | null) => (
                value != null ? t(`resourcesPage.types.${typeLabels[value]}`) : allTypesLabel
              )}
            </SelectTrigger>
            <SelectPopup align="start">
              <SelectItem value={null}>{allTypesLabel}</SelectItem>
              {trackerResourceTypes.map(resourceType => (
                <SelectItem key={resourceType} value={resourceType}>
                  {t(`resourcesPage.types.${typeLabels[resourceType]}`)}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
      </div>
    </div>
  );
}
