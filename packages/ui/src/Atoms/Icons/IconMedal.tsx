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

export function IconMedal({ size = 24, className }: IconProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 16 16"
      fill="transparent"
      className={className}
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M5.33331 10.1673V14.0249C5.33331 14.3965 5.72447 14.6383 6.05692 14.4721L7.77638 13.6123C7.91711 13.5419 8.08285 13.5419 8.22358 13.6123L9.94305 14.4721C10.2755 14.6383 10.6666 14.3965 10.6666 14.0249V10.1673M12.6666 6.00065C12.6666 8.57798 10.5773 10.6673 7.99998 10.6673C5.42265 10.6673 3.33331 8.57798 3.33331 6.00065C3.33331 3.42332 5.42265 1.33398 7.99998 1.33398C10.5773 1.33398 12.6666 3.42332 12.6666 6.00065Z"
        stroke="currentColor"
        strokeWidth="1.33333"
        strokeLinejoin="round"
      />
    </svg>
  );
}
