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

import { Option } from "../../Atoms/Select/Select";

export function SentitivityOptions() {
  const { __ } = useTranslate();

  const descriptions = {
    NONE: {
      label: __("None"),
      description: __("No sensitive data"),
    },
    LOW: {
      label: __("Low"),
      description: __("Public or non-sensitive data"),
    },
    MEDIUM: {
      label: __("Medium"),
      description: __("Internal/restricted data"),
    },
    HIGH: {
      label: __("High"),
      description: __("Confidential data"),
    },
    CRITICAL: {
      label: __("Critical"),
      description: __("Regulated/PII/financial data"),
    },
  } as const;

  return (
    <>
      {Object.entries(descriptions).map(([key, description]) => (
        <Option
          key={key}
          value={key}
          className="border-b border-border-low"
        >
          <span>
            <span className="text-sm font-bold">
              {description.label}
            </span>
            ,
            {" "}
            <span className="text-sm text-txt-secondary">
              {description.description}
            </span>
          </span>
        </Option>
      ))}
    </>
  );
}
