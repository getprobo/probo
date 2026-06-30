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

import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";

const CARD_PLACEHOLDERS = ["a", "b", "c", "d", "e", "f"];

export function SubprocessorsPageSkeleton() {
  return (
    <>
      <HeaderBand>
        <div className="flex w-full flex-col gap-2">
          <HeadingSkeleton size={7} className="w-64" />
        </div>
      </HeaderBand>
      <div className="flex w-full flex-col items-center px-8 py-8">
        <div className="flex w-full max-w-5xl flex-col gap-4">
          <TextSkeleton size={3} className="w-48" />
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {CARD_PLACEHOLDERS.map(placeholder => (
              <div key={placeholder} className="h-56 animate-pulse rounded-5 bg-sand-3" />
            ))}
          </div>
        </div>
      </div>
    </>
  );
}
