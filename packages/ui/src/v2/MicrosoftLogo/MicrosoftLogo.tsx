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

import type { ComponentProps } from "react";

export type MicrosoftLogoProps = ComponentProps<"svg">;

// Official Microsoft four-square mark. Fills are Microsoft's brand palette
// and do not follow currentColor — size it with width/height or a size class.
export function MicrosoftLogo(props: MicrosoftLogoProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 256 256"
      preserveAspectRatio="xMidYMid"
      {...props}
    >
      <rect x="0" y="0" width="121" height="121" fill="#F25022" />
      <rect x="135" y="0" width="121" height="121" fill="#7FBA00" />
      <rect x="0" y="135" width="121" height="121" fill="#00A4EF" />
      <rect x="135" y="135" width="121" height="121" fill="#FFB900" />
    </svg>
  );
}
