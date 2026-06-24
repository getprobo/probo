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
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";

import { hero } from "./variants";

export interface HeroProps {
  title: ReactNode;
  description?: ReactNode;
  children?: ReactNode;
}

// Landing hero (home): a size-8 title (+ optional description) in the shared
// white band, plus an optional bottom slot for page-specific content (the org
// contact row). The slot content owns its own divider/spacing so it disappears
// cleanly when empty.
export function Hero({ title, description, children }: HeroProps) {
  const { content, section } = hero();

  return (
    <HeaderBand>
      <div className={content()}>
        <div className={section()}>
          <Heading level={1} size={8} weight="medium" highContrast>
            {title}
          </Heading>
          {description != null && description !== "" && (
            <Text size={2} color="neutral">
              {description}
            </Text>
          )}
        </div>
        {children}
      </div>
    </HeaderBand>
  );
}
