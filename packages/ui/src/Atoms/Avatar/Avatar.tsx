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

import { tv } from "tailwind-variants";

type Props = {
  name: string;
  src?: string | null;
  size?: "s" | "m" | "l" | "xl";
  className?: string;
};

const avatar = tv({
  base: "bg-border-mid text-txt-invert! rounded-full font-semibold flex items-center justify-center flex-none",
  variants: {
    size: {
      s: "size-5 text-xss",
      m: "size-6 text-xxs",
      l: "size-8 text-sm",
      xl: "size-16 text-3xl",
    },
  },
  defaultVariants: {
    size: "m",
  },
});

export function Avatar(props: Props) {
  const className = avatar(props);
  if (props.src) {
    return <img className={className} src={props.src} alt={props.name} />;
  }
  return <div className={className}>{extractInitials(props.name ?? "")}</div>;
}

function extractInitials(name: string) {
  const words = name.split(" ");
  if (words.length === 2) {
    return words
      .map(word => word[0])
      .join("")
      .toUpperCase();
  }
  return name.substring(0, 2).toUpperCase();
}
