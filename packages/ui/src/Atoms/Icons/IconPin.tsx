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
