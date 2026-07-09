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

import type { ReactNode } from "react";

import { updatesList } from "./variants";

interface UpdatesListProps {
  // Dims the list while a page change is loading.
  busy?: boolean;
  // The rendered update rows.
  children: ReactNode;
}

// White card surface holding divider-separated update rows.
export function UpdatesList({ busy = false, children }: UpdatesListProps) {
  const { card, rows } = updatesList();

  return (
    <div className={card({ busy })} aria-busy={busy || undefined}>
      <div className={rows()}>{children}</div>
    </div>
  );
}
