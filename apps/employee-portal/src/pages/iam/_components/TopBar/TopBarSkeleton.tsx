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

import { AvatarSkeleton } from "@probo/ui/src/v2/Avatar/AvatarSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { useTopBarWide } from "./useTopBarWide";
import { topBar, topBarUserMenuTrigger } from "./variants";

// Loading placeholder paired with TopBar: reuses the same layout slots with
// skeleton primitives. Imports no Relay / Base UI, so it renders instantly.
export function TopBarSkeleton() {
  const slots = topBar({ wide: useTopBarWide() });

  return (
    <div className={slots.bar()}>
      <div className={slots.inner()}>
        <div className={slots.brand()}>
          <AvatarSkeleton size={2} radius="small" />
          <span className={slots.brandText()}>
            <TextSkeleton size={2} className={`w-24 ${slots.brandName()}`} />
            <TextSkeleton size={1} className={`w-28 ${slots.tagline()}`} />
          </span>
        </div>

        <div className={topBarUserMenuTrigger()}>
          <AvatarSkeleton size={1} radius="small" />
          <TextSkeleton size={2} className="w-20" />
        </div>
      </div>
    </div>
  );
}
