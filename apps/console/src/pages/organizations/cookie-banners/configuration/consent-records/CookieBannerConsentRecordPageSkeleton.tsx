// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
