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

import { DocumentAccessAction } from "./DocumentAccessAction";
import { documentListItem } from "./variants";

interface DocumentEntryProps {
  // Primary line (document title, file name, or framework name).
  title: ReactNode;
  // Accent sub-label (document type, file category, or report file name).
  meta: ReactNode;
  // Whether the viewer may open the entry (public or granted access).
  isAuthorized: boolean;
  // Whether an access request is already pending for the entry.
  requested: boolean;
  // Route to the document viewer, used when authorized.
  viewHref: string;
}

// Presentational row shared by the document / file / report list items: a title
// with accent metadata and the trailing access action. The connection-item
// wrappers own their fragments and supply these values.
export function DocumentEntry({ title, meta, isAuthorized, requested, viewHref }: DocumentEntryProps) {
  const { root, content } = documentListItem();

  return (
    <div className={root()}>
      <div className={content()}>
        <Text size={2} weight="medium" color="neutral" highContrast className="truncate">
          {title}
        </Text>
        <Text size={1} color="gold" className="truncate">
          {meta}
        </Text>
      </div>
      <DocumentAccessAction isAuthorized={isAuthorized} requested={requested} viewHref={viewHref} />
    </div>
  );
}
