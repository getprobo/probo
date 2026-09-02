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

import { CheckIcon } from "@phosphor-icons/react";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";

import { documentEmpty } from "./variants";

export interface DocumentEmptyProps {
  title: string;
  description?: string;
  icon?: ReactNode;
}

export function DocumentEmpty({ title, description, icon }: DocumentEmptyProps) {
  const slots = documentEmpty();

  return (
    <Card variant="soft" size={3} padding="none" className={slots.frame()}>
      <span className={slots.icon()}>
        {icon ?? <CheckIcon />}
      </span>
      <div className={slots.copy()}>
        <Text size={2} weight="medium" color="current" className={slots.label()}>
          {title}
        </Text>
        {description != null && description !== "" && (
          <Text size={2} color="current" className={slots.label()}>
            {description}
          </Text>
        )}
      </div>
    </Card>
  );
}
