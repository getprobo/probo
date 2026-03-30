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

export function IconRadioUnchecked({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 25 25" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M13 4.875C8.51269 4.875 4.875 8.51269 4.875 13C4.875 17.4873 8.51269 21.125 13 21.125C17.4873 21.125 21.125 17.4873 21.125 13C21.125 8.51269 17.4873 4.875 13 4.875ZM3 13C3 7.47715 7.47715 3 13 3C18.5228 3 23 7.47715 23 13C23 18.5228 18.5228 23 13 23C7.47715 23 3 18.5228 3 13Z" fill="currentColor" />
    </svg>
  );
}
