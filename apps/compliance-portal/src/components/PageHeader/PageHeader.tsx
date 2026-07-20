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
  // Sit the toolbar flush on the band's bottom border (e.g. an underlined tab
  // bar whose divider doubles as the header divider).
  flushBottomSpace?: boolean;
}

// Page header for the Compliance Portal nav pages: a size-7 title in the shared white
// band, with an optional count, inline actions, and a toolbar slot below.
export function PageHeader({ title, count, actions, children, flushBottomSpace }: PageHeaderProps) {
  const { content, titleRow, count: countSlot } = pageHeader();

  return (
    <HeaderBand flushBottomSpace={flushBottomSpace}>
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
