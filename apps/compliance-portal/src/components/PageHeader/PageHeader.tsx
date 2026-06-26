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
import type { ReactNode } from "react";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";

import { pageHeader } from "./variants";

export interface PageHeaderProps {
  title: ReactNode;
  // Optional muted item count appended to the title (e.g. "Documents (3)").
  count?: number;
  // Right-aligned controls on the title line (e.g. a "New Request" button).
  actions?: ReactNode;
  // Optional toolbar (tabs / filters / search) rendered below the title.
  children?: ReactNode;
}

// Page header for the Trust Center nav pages: a size-7 title in the shared white
// band, with an optional count, inline actions, and a toolbar slot below.
export function PageHeader({ title, count, actions, children }: PageHeaderProps) {
  const { content, titleRow, count: countSlot } = pageHeader();

  return (
    <HeaderBand>
      <div className={content()}>
        <div className={titleRow()}>
          <Heading level={1} size={7} weight="medium" highContrast>
            {title}
            {count != null && <span className={countSlot()}>{` (${count})`}</span>}
          </Heading>
          {actions}
        </div>
        {children}
      </div>
    </HeaderBand>
  );
}
