// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { useTranslate } from "@probo/i18n";
import { ActionDropdown, Button, IconPencil, Skeleton } from "@probo/ui";

/**
 * Skeleton state for the framework control panel
 */
export function ControlSkeleton() {
  const { __ } = useTranslate();
  return (
    <div className="space-y-6">
      <div className="flex justify-between">
        <Skeleton style={{ width: 72, height: 34 }} className="mb-3" />
        <div className="flex gap-2">
          <Button icon={IconPencil} variant="secondary" disabled>
            {__("Edit control")}
          </Button>
          <ActionDropdown variant="secondary" />
        </div>
      </div>
      <Skeleton style={{ width: "80%", height: 24 }} />
      <Skeleton style={{ height: 160 }} />
      <Skeleton style={{ height: 160 }} />
    </div>
  );
}
