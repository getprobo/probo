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

import type { VariantProps } from "tailwind-variants/lite";

import { Button } from "../Button/Button";
import { Text } from "../typography/Text";

import { inlineError } from "./variants";

export type InlineErrorProps
  = VariantProps<typeof inlineError>
    & {
      message: string;
      // Retry handler. When omitted, the retry action is hidden.
      onRetry?: () => void;
      retryLabel?: string;
      className?: string;
    };

// Presentational inline error (Figma "Error message / Inline"). Copy and the
// retry handler come from the caller; this only lays them out.
export function InlineError({ layout, message, onRetry, retryLabel = "Retry", className }: InlineErrorProps) {
  const slots = inlineError({ layout });

  return (
    <div className={slots.root({ className })}>
      <Text size={2} color="neutral" className={slots.message()}>{message}</Text>
      {onRetry && (
        <Button size={2} variant="soft" color="neutral" onClick={onRetry}>
          {retryLabel}
        </Button>
      )}
    </div>
  );
}
