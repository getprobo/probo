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

import { ndaPage } from "./variants";

export function NDAPageSkeleton() {
  const slots = ndaPage();

  return (
    <div className={slots.root()}>
      <HeaderBand flushBottomSpace>
        <div className={slots.header()}>
          <div className={slots.text()}>
            <HeadingSkeleton size={7} className="w-80" />
            <TextSkeleton size={2} className="w-96" />
            <TextSkeleton size={1} className="w-full max-w-2xl" />
          </div>
          <div className={slots.toolbar()}>
            <div className={slots.toolbarStart()}>
              <ButtonSkeleton size={2} className="w-40" />
            </div>
            <div className={slots.actions()}>
              <ButtonSkeleton size={2} className="w-32" />
            </div>
          </div>
        </div>
      </HeaderBand>
      <div className={slots.body()}>
        <div className={slots.stage()} />
      </div>
    </div>
  );
}
