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

import { TableSkeleton } from "@probo/ui/src/v2/Table/TableSkeleton";

export function TrackerPatternDetailPageSkeleton() {
  return (
    <div className="flex flex-col gap-6 animate-pulse">
      <div className="flex flex-col gap-4">
        <div className="h-4 w-20 rounded bg-sand-a4" />
        <div className="flex items-start justify-between gap-4">
          <div className="flex min-w-0 flex-col gap-1">
            <div className="flex items-center gap-2">
              <div className="h-8 w-48 rounded bg-sand-a4" />
              <div className="h-5 w-16 rounded bg-sand-a4" />
              <div className="h-5 w-12 rounded bg-sand-a4" />
            </div>
            <div className="h-4 w-32 rounded bg-sand-a4" />
          </div>
          <div className="h-8 w-24 rounded bg-sand-a4" />
        </div>
      </div>
      <div className="flex flex-col gap-4">
        <div className="h-5 w-24 rounded bg-sand-a4" />
        <div className="flex flex-col gap-4 rounded-4 bg-sand-a2 p-4">
          <div className="h-16 rounded bg-sand-a4" />
          <div className="grid grid-cols-2 gap-3 max-sm:grid-cols-1">
            <div className="h-8 rounded bg-sand-a4" />
            <div className="h-8 rounded bg-sand-a4" />
          </div>
        </div>
      </div>
      <div className="flex flex-col gap-4">
        <div className="h-5 w-28 rounded bg-sand-a4" />
        <div className="h-4 w-56 rounded bg-sand-a4" />
        <TableSkeleton variant="surface" columns={5} count={4} />
      </div>
    </div>
  );
}
