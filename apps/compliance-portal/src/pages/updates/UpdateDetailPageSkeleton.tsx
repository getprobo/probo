// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
