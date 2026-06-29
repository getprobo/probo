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

import { ProboLogo } from "@probo/ui/src/v2/ProboLogo/ProboLogo";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";

import { dotPatternStyle } from "#/components/MediaTile/variants";

import { poweredBy } from "./variants";

export interface PoweredByProps {
  // Localized "Powered by" label. Defaults to the English string.
  label?: ReactNode;
  // Destination for the Probo logo link.
  href?: string;
}

// "Powered by Probo" footer: a borderless attribution on the body surface, with
// a dotted texture fading in from the bottom and the full Probo logo (picto +
// wordmark).
export function PoweredBy({ label = "Powered by", href = "https://www.probo.com/" }: PoweredByProps) {
  const slots = poweredBy();

  return (
    <footer className={slots.root()}>
      <div className={slots.backdrop()} style={dotPatternStyle} />
      <div className={slots.backdropFade()} />
      <div className={slots.content()}>
        <Text size={1} color="neutral">
          {label}
        </Text>
        <a href={href} target="_blank" rel="noopener noreferrer" aria-label="Probo">
          <ProboLogo className={slots.logo()} aria-hidden />
        </a>
      </div>
    </footer>
  );
}
