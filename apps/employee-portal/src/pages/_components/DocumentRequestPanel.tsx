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

import { CaretLeftIcon } from "@phosphor-icons/react";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import type { ReactNode } from "react";

import { documentRequestPanel } from "./variants";

interface DocumentRequestPanelProps {
  backLabel: string;
  backTo: string;
  title: string;
  // Instruction or signed/approved/rejected status, stacked with the title.
  detail?: ReactNode;
  children?: ReactNode;
}

// Shared request-column chrome: back to the list, title, optional detail,
// then actions.
export function DocumentRequestPanel({
  backLabel,
  backTo,
  title,
  detail,
  children,
}: DocumentRequestPanelProps) {
  const slots = documentRequestPanel();

  return (
    <div className={slots.root()}>
      <div className={slots.copy()}>
        <div className={slots.back()}>
          <Link
            to={backTo}
            size={2}
            color="neutral"
            underline={false}
            iconStart={<CaretLeftIcon />}
          >
            {backLabel}
          </Link>
        </div>
        <Heading level={1} size={6} weight="medium" highContrast>
          {title}
        </Heading>
        {detail}
      </div>
      {children}
    </div>
  );
}
