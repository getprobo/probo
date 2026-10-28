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

export function Daytona(props: ComponentProps<"svg">) {
  return (
    <svg
      viewBox="0 -0.45 26.64 26.64"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M1.412 17.375h9.666v3.048H1.412ZM14.39 6.707h11.046v3.048H14.39Zm-5.799.94L16.792.103l2.343 2.156-8.2 7.543Z"
        fill="#000000"
      />
      <path
        d="M8.642 15.267 2.343 9.473 0 11.628l6.298 5.794Zm8.285 4.263-6.639 6.107-2.344-2.155 6.64-6.107Zm-.051-7.619 7.42 6.825 2.343-2.156-7.42-6.825Zm-8.285-7.49v7.366H5.277V4.421Z"
        fill="#000000"
      />
      <path d="M20.19 15.089v9.144h-3.313v-9.144Z" fill="#000000" />
    </svg>
  );
}
