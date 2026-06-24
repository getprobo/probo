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

export function CookieBannerConsentRecordsPageSkeleton() {
  return (
    <div className="space-y-4 animate-pulse">
      <div className="flex items-center gap-4">
        <div className="h-9 w-36 rounded bg-bg-subtle" />
        <div className="h-9 w-48 rounded bg-bg-subtle" />
        <div className="h-9 w-40 rounded bg-bg-subtle" />
      </div>
      <div className="rounded-lg border border-border-low">
        <div className="h-10 border-b border-border-low bg-bg-subtle" />
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="h-12 border-b border-border-low last:border-b-0" />
        ))}
      </div>
    </div>
  );
}
