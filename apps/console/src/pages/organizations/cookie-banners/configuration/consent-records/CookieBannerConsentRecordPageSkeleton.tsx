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

export function CookieBannerConsentRecordPageSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <div className="rounded-2xl border border-border-low p-6 space-y-4">
        {Array.from({ length: 7 }).map((_, i) => (
          <div key={i} className="flex items-center justify-between py-3 border-b border-border-low last:border-b-0">
            <div className="h-4 w-28 rounded bg-bg-subtle" />
            <div className="h-4 w-48 rounded bg-bg-subtle" />
          </div>
        ))}
      </div>
      <div className="rounded-2xl border border-border-low p-6 space-y-4">
        <div className="h-5 w-32 rounded bg-bg-subtle" />
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="space-y-2 py-3 border-b border-border-low last:border-b-0">
            <div className="flex items-center justify-between">
              <div className="h-4 w-32 rounded bg-bg-subtle" />
              <div className="h-5 w-16 rounded bg-bg-subtle" />
            </div>
            <div className="ml-4 space-y-1">
              <div className="h-3 w-40 rounded bg-bg-subtle" />
              <div className="h-3 w-36 rounded bg-bg-subtle" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
