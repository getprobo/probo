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

import { clsx } from "clsx";

type Props = {
  level: "LOW" | "MEDIUM" | "HIGH" | "URGENT";
};

export function PriorityLevel({ level }: Props) {
  if (level === "URGENT") {
    return (
      <div className="w-max flex items-center justify-center text-txt-danger">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path
            d="M7 1.75v5.25M7 10.5h.005"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </div>
    );
  }

  const bars = level === "HIGH" ? 3 : level === "MEDIUM" ? 2 : 1;

  return (
    <div className="w-max p-[2px] flex gap-[2px] items-end">
      <div
        className={clsx(
          "h-1 w-[3px] rounded",
          bars >= 1 ? "bg-txt-secondary" : "bg-txt-quaternary",
        )}
      />
      <div
        className={clsx(
          "h-2 w-[3px] rounded",
          bars >= 2 ? "bg-txt-secondary" : "bg-txt-quaternary",
        )}
      />
      <div
        className={clsx(
          "h-3 w-[3px] rounded",
          bars >= 3 ? "bg-txt-secondary" : "bg-txt-quaternary",
        )}
      />
    </div>
  );
}
