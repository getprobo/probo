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

import { Card } from "@probo/ui/src/v2/Card/Card";
import type { ReactNode } from "react";

import { documentReviewItem } from "../../variants";

import { DocumentReviewItemContext } from "./documentReviewItemContext";

interface DocumentReviewItemProps {
  kind: "comment" | "suggestion";
  active: boolean;
  children: ReactNode;
  onActivate: () => void;
}

// The primary author row renders the CardButton around the author name; see
// DocumentReviewAuthor.
export function DocumentReviewItem({
  kind,
  active,
  children,
  onActivate,
}: DocumentReviewItemProps) {
  const { root, body } = documentReviewItem();
  const context = { kind, active, onActivate };

  return (
    <Card
      variant={active ? "soft" : "gold"}
      padding={1}
      interactive
      className={root()}
      data-active={active || undefined}
    >
      <DocumentReviewItemContext value={context}>
        <div className={body()}>{children}</div>
      </DocumentReviewItemContext>
    </Card>
  );
}
