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

const BODY_PLACEHOLDERS = ["a", "b", "c", "d", "e"];

export function UpdateDetailPageSkeleton() {
  return (
    <>
      <HeaderBand>
        <div className="flex w-full items-center justify-between gap-4">
          <ButtonSkeleton size={2} />
          <ButtonSkeleton size={2} />
        </div>
      </HeaderBand>
      <div className="flex w-full flex-col items-center px-8 py-8">
        <div className="flex w-full max-w-2xl flex-col gap-4" aria-hidden>
          <TextSkeleton size={1} className="w-28" />
          <HeadingSkeleton size={7} className="w-96" />
          <div className="flex flex-col gap-2">
            {BODY_PLACEHOLDERS.map(placeholder => (
              <TextSkeleton key={placeholder} size={3} className="w-full" />
            ))}
          </div>
        </div>
      </div>
    </>
  );
}
