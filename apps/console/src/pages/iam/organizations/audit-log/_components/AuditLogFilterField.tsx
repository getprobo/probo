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

import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { useState } from "react";

import { auditLogList } from "../variants";

interface AuditLogFilterFieldProps {
  label: string;
  value: string;
  placeholder?: string;
  onCommit: (value: string) => void;
}

export function AuditLogFilterField({
  label,
  value,
  placeholder,
  onCommit,
}: AuditLogFilterFieldProps) {
  const [draft, setDraft] = useState(value);
  const { filter } = auditLogList();

  function commit() {
    const next = draft.trim();
    if (next !== draft) {
      setDraft(next);
    }
    if (next !== value) {
      onCommit(next);
    }
  }

  return (
    <div className={filter()}>
      <Field label={label}>
        <TextField
          size={2}
          value={draft}
          placeholder={placeholder}
          aria-label={label}
          onValueChange={setDraft}
          onBlur={commit}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              event.preventDefault();
              commit();
            }
          }}
        />
      </Field>
    </div>
  );
}
