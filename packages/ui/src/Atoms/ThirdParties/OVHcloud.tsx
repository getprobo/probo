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

// The mark only, lifted from OVHcloud's own wordmark. The viewBox is the
// glyph's bounding box squared off around its centre, because the mark is
// wider than it is tall and every sibling logo here is square.
export function OVHcloud(props: ComponentProps<"svg">) {
  return (
    <svg
      viewBox="36.97 22.12 78.18 78.18"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        fill="#000E9C"
        d="M110.11,40.31a40.1,40.1,0,0,1-3.74,44.49H84.93l6.6-11.67H82.8L93.09,55h8.78l8.24-14.67ZM67.77,84.8H45.91a39.59,39.59,0,0,1-3.79-44.59L56.3,64.84,71.93,37.62h23L67.78,84.78h0v0Z"
      />
    </svg>
  );
}
