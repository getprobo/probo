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

import { CopyIcon } from "@phosphor-icons/react";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";

import type { WorkloadIdentitySetupRow } from "../_lib/workloadIdentitySetup";

interface WorkloadIdentitySetupValuesProps {
  rows: WorkloadIdentitySetupRow[];
}

export function WorkloadIdentitySetupValues({
  rows,
}: WorkloadIdentitySetupValuesProps) {
  return (
    <div className="flex flex-col gap-2 rounded-2 bg-sand-3 p-3">
      {rows.map(row => (
        <div key={row.label} className="flex flex-col gap-1">
          <Text size={1} color="faint">{row.label}</Text>
          <div className="flex min-w-0 items-center gap-1">
            <Code size={2} className="min-w-0 flex-1 break-all">{row.value}</Code>
            <IconButton
              type="button"
              size={1}
              variant="soft"
              color="neutral"
              aria-label={row.copyLabel}
              onClick={row.onCopy}
            >
              <CopyIcon />
            </IconButton>
          </div>
        </div>
      ))}
    </div>
  );
}
