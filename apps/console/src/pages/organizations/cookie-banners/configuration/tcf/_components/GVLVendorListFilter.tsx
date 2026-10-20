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
  type GVLVendorMembershipOption,
  gvlVendorMembershipOptions,
  isGVLVendorMembershipOption,
  useGVLVendorFilters,
} from "../_lib/useGVLVendorFilters";
import { tcfSection } from "../variants";

export function GVLVendorListFilter() {
  const { t } = useTranslation("organizations/cookie-banners");
  const { membership, setMembership } = useGVLVendorFilters();
  const { filter } = tcfSection();
  const allLabel = t("tcfPage.filter.all");

  function optionLabel(option: GVLVendorMembershipOption): string {
    return t(`tcfPage.filter.${option}`);
  }

  return (
    <div className={filter()}>
      <Select
        value={membership === "all" ? null : membership}
        onValueChange={(value: string | null) => {
          if (value == null) {
            setMembership("all");
            return;
          }
          if (isGVLVendorMembershipOption(value)) {
            setMembership(value);
          }
        }}
      >
        <SelectTrigger
          size={2}
          placeholder={allLabel}
          aria-label={t("tcfPage.filter.label")}
        >
          {(value: GVLVendorMembershipOption | null) => (
            value != null ? optionLabel(value) : allLabel
          )}
        </SelectTrigger>
        <SelectPopup align="end">
          <SelectItem value={null}>{allLabel}</SelectItem>
          {gvlVendorMembershipOptions.map(option => (
            <SelectItem key={option} value={option}>
              {optionLabel(option)}
            </SelectItem>
          ))}
        </SelectPopup>
      </Select>
    </div>
  );
}
