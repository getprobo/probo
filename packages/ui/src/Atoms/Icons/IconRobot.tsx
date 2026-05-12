// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

export function IconRobot({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M13 3V5H17C18.6569 5 20 6.34315 20 8V19C20 20.6569 18.6569 22 17 22H7C5.34315 22 4 20.6569 4 19V8C4 6.34315 5.34315 5 7 5H11V3H13ZM7 7C6.44772 7 6 7.44772 6 8V19C6 19.5523 6.44772 20 7 20H17C17.5523 20 18 19.5523 18 19V8C18 7.44772 17.5523 7 17 7H7Z" />
      <circle cx="9.5" cy="12" r="1.5" />
      <circle cx="14.5" cy="12" r="1.5" />
      <path d="M9 16H15V17.5H9V16Z" />
      <path d="M12 1C12.5523 1 13 1.44772 13 2V3H11V2C11 1.44772 11.4477 1 12 1Z" />
      <path d="M2 11C2 10.4477 2.44772 10 3 10V14C2.44772 14 2 13.5523 2 13V11Z" />
      <path d="M21 10C21.5523 10 22 10.4477 22 11V13C22 13.5523 21.5523 14 21 14V10Z" />
    </svg>
  );
}
