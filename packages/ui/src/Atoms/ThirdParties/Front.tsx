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

export function Front(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 65 65" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        d="M0 63.6152H21.3918V21.3008H64.0428V-0.00341797H0V63.6152Z"
        fill="#A857F1"
      />
      <path
        d="M45.5945 64.8878C56.2523 64.8878 64.8922 56.2479 64.8922 45.5901C64.8922 34.9323 56.2523 26.2925 45.5945 26.2925C34.9367 26.2925 26.2969 34.9323 26.2969 45.5901C26.2969 56.2479 34.9367 64.8878 45.5945 64.8878Z"
        fill="#A857F1"
      />
    </svg>
  );
}
