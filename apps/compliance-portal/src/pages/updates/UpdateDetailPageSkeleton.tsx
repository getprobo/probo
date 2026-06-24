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
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";

import { updateArticle } from "./_components/variants";

const BODY_LINE_COUNT = 5;

export function UpdateDetailPageSkeleton() {
  const { toolbar, content, article } = updateArticle();

  return (
    <>
      <HeaderBand>
        <div className={toolbar()}>
          <ButtonSkeleton size={2} />
          <ButtonSkeleton size={2} />
        </div>
      </HeaderBand>
      <div className={content()}>
        <div className={article()} aria-hidden>
          <TextSkeleton size={1} className="w-28" />
          <HeadingSkeleton size={7} className="w-96" />
          <div className="flex flex-col gap-2">
            {Array.from({ length: BODY_LINE_COUNT }, (_, index) => (
              <TextSkeleton key={index} size={3} className="w-full" />
            ))}
          </div>
        </div>
      </div>
    </>
  );
}
