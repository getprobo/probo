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

export function IconArrowCornerDownLeft({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path d="M7.29297 10.293C7.68349 9.90244 8.31651 9.90244 8.70703 10.293C9.09756 10.6835 9.09756 11.3165 8.70703 11.707L6.41406 14H17C18.1046 14 19 13.1046 19 12V5C19 4.44772 19.4477 4 20 4C20.5523 4 21 4.44772 21 5V12C21 14.2092 19.2092 16 17 16H6.41406L8.70703 18.293L8.77539 18.3691C9.09574 18.7619 9.07315 19.3409 8.70703 19.707C8.34092 20.0731 7.76192 20.0957 7.36914 19.7754L7.29297 19.707L3.29297 15.707C2.90245 15.3165 2.90245 14.6835 3.29297 14.293L7.29297 10.293Z" fill="currentColor" />
    </svg>
  );
}
