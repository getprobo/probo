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

import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";

import { documentSection } from "./variants";

interface DocumentSectionProps {
  title: ReactNode;
  description?: ReactNode;
  // The list rows (document / file / report list items).
  children: ReactNode;
}

// One category group: a localized header above a bordered list of rows.
export function DocumentSection({ title, description, children }: DocumentSectionProps) {
  const { root, header, list } = documentSection();

  return (
    <section className={root()}>
      <div className={header()}>
        <Text size={3} weight="medium" color="neutral" highContrast role="heading" aria-level={2}>
          {title}
        </Text>
        {description != null && (
          <Text size={2} color="neutral">
            {description}
          </Text>
        )}
      </div>
      <div className={list()}>{children}</div>
    </section>
  );
}
