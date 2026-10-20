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

import { CardSkeleton } from "@probo/ui/src/v2/Card/CardSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { CookieBannerPageHeaderSkeleton } from "../../_components/CookieBannerPageHeaderSkeleton";
import { consentRecordPage } from "../../variants";

export function CookieBannerConsentRecordPageSkeleton() {
  const {
    root,
    header,
    body,
    column,
    actionCard,
    locationCard,
    requestCard,
    tcfCard,
    categoriesCard,
  } = consentRecordPage();

  return (
    <div className={root()}>
      <div className={header()}>
        <TextSkeleton size={2} className="w-16" />
        <CookieBannerPageHeaderSkeleton titleClassName="w-40" />
      </div>
      <div className={body()}>
        <div className={column()}>
          <CardSkeleton size={2} className={actionCard({ className: "h-56" })} />
          <CardSkeleton size={2} className={requestCard({ className: "h-64" })} />
          <CardSkeleton size={2} className={tcfCard({ className: "h-72" })} />
        </div>
        <div className={column()}>
          <CardSkeleton size={2} className={categoriesCard({ className: "h-40" })} />
          <CardSkeleton size={2} className={locationCard({ className: "h-56" })} />
        </div>
      </div>
    </div>
  );
}
