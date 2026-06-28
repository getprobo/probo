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

import { AvatarSkeleton } from "@probo/ui/src/v2/Avatar/AvatarSkeleton";
import { ButtonSkeleton } from "@probo/ui/src/v2/Button/ButtonSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { topBar } from "./variants";

// Width per nav item, sized to its English label so the placeholder row reads
// like the real nav. See .cursor/rules/skeleton-width-sync.mdc.
const NAV_ITEMS = [
  { key: "documents", width: "w-16" },
  { key: "subprocessors", width: "w-24" },
  { key: "updates", width: "w-12" },
  { key: "requests", width: "w-14" },
] as const;

// Loading placeholder paired with TopBar: reuses the same layout slots with
// skeleton primitives. Imports no Relay / Base UI, so it renders instantly.
export function TopBarSkeleton() {
  const slots = topBar();

  return (
    <div className={slots.bar()}>
      <div className={slots.inner()}>
        <div className={slots.brand()}>
          <AvatarSkeleton size={1} radius="small" />
          <TextSkeleton size={2} className="w-24" />
        </div>

        <div className={slots.spacer()} />

        <nav className={slots.nav()}>
          {NAV_ITEMS.map(item => (
            <TextSkeleton key={item.key} size={2} className={`mx-3 ${item.width}`} />
          ))}
          <ButtonSkeleton size={2} />
        </nav>
      </div>
    </div>
  );
}
