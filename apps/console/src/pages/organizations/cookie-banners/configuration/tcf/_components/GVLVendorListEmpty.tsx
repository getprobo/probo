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

import { StorefrontIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { useGVLVendorFilters } from "../_lib/useGVLVendorFilters";
import { gvlVendorListEmpty } from "../variants";

export function GVLVendorListEmpty() {
  const { t } = useTranslation("organizations/cookie-banners");
  const { hasActiveFilters, clear } = useGVLVendorFilters();
  const { root, icon, body } = gvlVendorListEmpty();

  return (
    <div className={root()}>
      <span className={icon()}>
        <StorefrontIcon weight="light" />
      </span>
      <div className={body()}>
        <Text size={2} weight="medium" color="faint">
          {hasActiveFilters ? t("tcfPage.emptyFiltered") : t("tcfPage.emptyCatalog")}
        </Text>
        {hasActiveFilters && (
          <Text size={2} color="faint">
            {t("tcfPage.emptyFilteredDescription")}
          </Text>
        )}
      </div>
      {hasActiveFilters && (
        <Button variant="ghost" color="neutral" onClick={clear}>
          {t("tcfPage.actions.clearSearch")}
        </Button>
      )}
    </div>
  );
}
