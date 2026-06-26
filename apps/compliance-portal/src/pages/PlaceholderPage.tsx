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

import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useLocation } from "react-router";

// Temporary stub for the Trust Center sections so the top-bar nav links resolve
// and the active state renders. Real pages replace these per section.
export default function PlaceholderPage() {
  const { pathname } = useLocation();
  const segment = pathname.replace(/^\//, "").split("/")[0] ?? "";
  const title = segment.charAt(0).toUpperCase() + segment.slice(1);

  return (
    <main className="mx-auto w-full max-w-[1024px] px-8 py-10">
      <Heading>{title}</Heading>
    </main>
  );
}
