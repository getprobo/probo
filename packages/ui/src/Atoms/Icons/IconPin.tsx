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

import type { IconProps } from "./type";

export function IconPin({ size = 24, className }: IconProps) {
  return (
    <svg
      width={size}
      height={size}
      className={className}
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 17 16"
    >
      <path
        stroke="currentColor"
        strokeLinejoin="round"
        strokeWidth="1.33"
        d="M10 6.67a1.67 1.67 0 1 1-3.33 0 1.67 1.67 0 0 1 3.33 0Z"
      />
      <path
        stroke="currentColor"
        strokeLinejoin="round"
        strokeWidth="1.33"
        d="M13 6.67c0 2.91-2.6 5.55-3.91 6.71-.44.38-1.07.38-1.51 0-1.32-1.16-3.91-3.8-3.91-6.71a4.67 4.67 0 1 1 9.33 0Z"
      />
    </svg>
  );
}
