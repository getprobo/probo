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

import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { CommitmentCardSkeleton } from "#/components/CommitmentCard/CommitmentCardSkeleton";

import { securityCommitments } from "./variants";

export function SecurityCommitmentsSectionSkeleton() {
  const slots = securityCommitments();

  return (
    <section className={slots.root()} aria-hidden>
      <div className={slots.group()}>
        <div className={slots.groupHeader()}>
          <TextSkeleton size={1} className="w-32" />
          <TextSkeleton size={2} className="w-28" />
          <TextSkeleton size={2} className="w-full max-w-2xl" />
        </div>
        <div className={slots.grid()}>
          {Array.from({ length: 3 }, (_, index) => (
            <CommitmentCardSkeleton key={index} />
          ))}
        </div>
      </div>
    </section>
  );
}
