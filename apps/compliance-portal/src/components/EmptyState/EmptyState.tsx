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

import { emptyState } from "./variants";

interface EmptyStateProps {
  // Leading phosphor icon, toned and sized by the icon slot.
  icon: ReactNode;
  title: ReactNode;
  description?: ReactNode;
  // Optional call to action (e.g. a Button) below the copy.
  action?: ReactNode;
}

// Generic empty-state placeholder shared across list/collection regions: a
// faint icon, a title, optional guidance, and an optional call to action.
// Purely presentational; consumers supply their own copy and action.
export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  const slots = emptyState();

  return (
    <div className={slots.root()}>
      <span className={slots.icon()}>{icon}</span>
      <div className={slots.body()}>
        <Text size={2} weight="medium" color="faint">
          {title}
        </Text>
        {description != null && (
          <Text size={2} color="faint">
            {description}
          </Text>
        )}
      </div>
      {action}
    </div>
  );
}
