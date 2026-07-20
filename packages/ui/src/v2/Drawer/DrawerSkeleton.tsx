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

import { TextSkeleton } from "../typography/TextSkeleton";

import { drawer, drawerSkeleton } from "./variants";

// Loading placeholder matching the drawer frame (header + body + footer),
// importing only variants and skeleton primitives — never Base UI.
export function DrawerSkeleton() {
  const { header, body, footer } = drawer();

  return (
    <div className={drawerSkeleton()} aria-hidden>
      <div className={header()}>
        <TextSkeleton size={4} className="w-24" />
        <TextSkeleton size={2} className="w-8" />
      </div>
      <div className={body()}>
        <TextSkeleton size={2} className="w-full" />
        <TextSkeleton size={2} className="w-3/4" />
        <TextSkeleton size={2} className="w-5/6" />
      </div>
      <div className={footer()}>
        <TextSkeleton size={2} className="w-full" />
      </div>
    </div>
  );
}
