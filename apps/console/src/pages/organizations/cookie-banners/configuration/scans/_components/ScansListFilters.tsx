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
  type CookieBannerScanState,
  cookieBannerScanStates,
  isCookieBannerScanState,
  scanStateBadges,
} from "../_lib/scanStates";
import { useScansListFilters } from "../_lib/useScansListFilters";

export function ScansListFilters() {
  const { t } = useTranslation("organizations/cookie-banners");
  const { state, setState } = useScansListFilters();
  const { tools, filters, filter } = cookieBannerList();
  const anyStateLabel = t("scansPage.states.any");

  return (
    <div className={tools()}>
      <div className={filters()}>
        <div className={filter()}>
          <Select
            value={state}
            onValueChange={(value: string | null) => {
              if (value == null) {
                setState(null);
                return;
              }
              if (isCookieBannerScanState(value)) {
                setState(value);
              }
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={anyStateLabel}
              aria-label={t("scansPage.filters.state")}
            >
              {(value: CookieBannerScanState | null) => (
                value != null
                  ? t(`scansPage.states.${scanStateBadges[value].labelKey}`)
                  : anyStateLabel
              )}
            </SelectTrigger>
            <SelectPopup align="start">
              <SelectItem value={null}>{anyStateLabel}</SelectItem>
              {cookieBannerScanStates.map(scanState => (
                <SelectItem key={scanState} value={scanState}>
                  {t(`scansPage.states.${scanStateBadges[scanState].labelKey}`)}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
      </div>
    </div>
  );
}
