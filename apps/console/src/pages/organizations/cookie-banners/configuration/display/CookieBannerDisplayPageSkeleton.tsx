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
import { CardSkeleton } from "@probo/ui/src/v2/Card/CardSkeleton";
import { ListSkeleton } from "@probo/ui/src/v2/List/ListSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { cookieBannerDisplaySection, cookieBannerPage, cookieBannerThemeSection } from "../../variants";

export function CookieBannerDisplayPageSkeleton() {
  const categories = cookieBannerDisplaySection();
  const theme = cookieBannerThemeSection();

  return (
    <div className={cookieBannerPage()}>
      <section className={categories.root()}>
        <div className={categories.intro()}>
          <div className={categories.heading()}>
            <HeadingSkeleton size={4} className="w-28" />
            <TextSkeleton size={2} className="w-96" />
          </div>
          <ButtonSkeleton size={2} className="w-32" />
        </div>
        <ListSkeleton count={4} />
      </section>
      <section className={theme.root()}>
        <div className={theme.intro()}>
          <div className={theme.heading()}>
            <HeadingSkeleton size={4} className="w-20" />
            <TextSkeleton size={2} className="w-full max-w-xl" />
          </div>
          <ButtonSkeleton size={2} className="w-36" />
        </div>
        <CardSkeleton size={2} className="h-56" />
        <CardSkeleton size={2} className="h-70" />
      </section>
    </div>
  );
}
