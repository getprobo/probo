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

import { ButtonSkeleton } from "@probo/ui/src/v2/Button/ButtonSkeleton";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextFieldSkeleton } from "@probo/ui/src/v2/form/TextFieldSkeleton";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { CookieBannerPageHeaderSkeleton } from "../../_components/CookieBannerPageHeaderSkeleton";
import {
  cookieBannerPage,
  cookieBannerSettingsSection,
  trackerMaxAgeField,
} from "../../variants";

export function NewTrackerPatternPageSkeleton() {
  const { card, fields, field, pair, actions } = cookieBannerSettingsSection();
  const maxAge = trackerMaxAgeField();

  return (
    <div className={cookieBannerPage()}>
      <CookieBannerPageHeaderSkeleton titleClassName="w-40" />
      <Card size={2} variant="soft" className={card()}>
        <div className={fields()}>
          <div className={field()}>
            <TextFieldSkeleton size={2} />
            <TextSkeleton size={1} className="w-72" />
          </div>
          <div className={pair()}>
            <SelectSkeleton size={2} />
            <SelectSkeleton size={2} />
          </div>
          <TextFieldSkeleton size={2} className="h-20" />
          <div className={maxAge.root()}>
            <TextFieldSkeleton size={2} className={maxAge.value()} />
            <SelectSkeleton size={2} className={maxAge.unit()} />
          </div>
          <div className={actions({ className: "items-center gap-4" })}>
            <TextSkeleton size={2} className="w-14" />
            <ButtonSkeleton size={2} className="w-32" />
          </div>
        </div>
      </Card>
    </div>
  );
}
