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

import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { useTranslation } from "react-i18next";

import { cookieBannerList } from "../../../variants";

interface FindingsPaginationProps {
  hasPrevious: boolean;
  hasNext: boolean;
  disabled?: boolean;
  onPrevious: () => void;
  onNext: () => void;
}

export function FindingsPagination({
  hasPrevious,
  hasNext,
  disabled,
  onPrevious,
  onNext,
}: FindingsPaginationProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { pager } = cookieBannerList();

  return (
    <div className={pager()}>
      <Pagination
        hasPrevious={hasPrevious}
        hasNext={hasNext}
        previousLabel={t("scansPage.actions.previous")}
        nextLabel={t("scansPage.actions.next")}
        showLabels
        disabled={disabled}
        onPrevious={onPrevious}
        onNext={onNext}
      />
    </div>
  );
}
