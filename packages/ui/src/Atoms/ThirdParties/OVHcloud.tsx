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

// The mark only, lifted from OVHcloud's 2024 wordmark. The viewBox is the
// glyph's bounding box squared off around its centre, because the mark is
// wider than it is tall and the sibling logos are square.
export function OVHcloud(props: ComponentProps<"svg">) {
  return (
    <svg
      viewBox="33.99 21.63 78.15 78.15"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        fill="#000E9C"
        fillRule="evenodd"
        clipRule="evenodd"
        d="M107.1,39.8c7.9,14.2,6.4,31.8-3.7,44.5H81.9l6.6-11.7h-8.7l10.3-18.1h8.8L107.1,39.8z M64.8,84.3H42.9c-10.3-12.7-11.8-30.4-3.8-44.6l14.2,24.6l15.6-27.2h23L64.8,84.3z"
      />
    </svg>
  );
}
