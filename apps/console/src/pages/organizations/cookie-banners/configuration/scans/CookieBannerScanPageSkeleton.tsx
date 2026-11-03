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

import { Card } from "@probo/ui/src/v2/Card/Card";

import { scanCounts, scanTimeline } from "../../variants";

export function CookieBannerScanPageSkeleton() {
  const { root, grid } = scanCounts();
  const { frame, columns } = scanTimeline();

  return (
    <div className="flex flex-col gap-6 animate-pulse">
      <div className="flex flex-col gap-4">
        <div className="h-4 w-16 rounded bg-sand-a4" />
        <div className="flex items-center gap-2">
          <div className="h-8 w-56 rounded bg-sand-a4" />
          <div className="h-5 w-16 rounded bg-sand-a4" />
        </div>
      </div>
      <div className="flex flex-col gap-3">
        <div className="h-5 w-24 rounded bg-sand-a4" />
        <Card variant="soft" size={3} padding={4} className={frame()}>
          <div className={columns()}>
            <div className="flex flex-col gap-3">
              <div className="h-8 rounded bg-sand-a3" />
              <div className="h-8 rounded bg-sand-a3" />
              <div className="h-8 rounded bg-sand-a3" />
            </div>
            <div className="flex flex-col gap-3">
              <div className="h-6 rounded bg-sand-a3" />
              <div className="h-6 rounded bg-sand-a3" />
            </div>
          </div>
        </Card>
      </div>
      <div className={root()}>
        <div className={grid()}>
          <div className="h-16 rounded-3 bg-sand-a3" />
          <div className="h-16 rounded-3 bg-sand-a3" />
          <div className="h-16 rounded-3 bg-sand-a3" />
          <div className="h-16 rounded-3 bg-sand-a3" />
        </div>
      </div>
    </div>
  );
}
