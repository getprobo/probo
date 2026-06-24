// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
