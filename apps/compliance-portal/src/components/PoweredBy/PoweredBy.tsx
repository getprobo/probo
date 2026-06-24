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
