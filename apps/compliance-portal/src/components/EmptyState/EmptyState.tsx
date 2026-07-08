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
