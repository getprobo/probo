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
import { useId } from "react";

export type ChromeLogoProps = Omit<ComponentProps<"svg">, "width" | "height"> & {
  size?: string | number;
};

// Official Google Chrome mark. Fills are Chrome's brand palette and do not
// follow currentColor. Size it with `size`, matching Phosphor icons.
export function ChromeLogo({ size = "1em", ...props }: ChromeLogoProps) {
  const uid = useId();
  const green = `${uid}-green`;
  const yellow = `${uid}-yellow`;
  const red = `${uid}-red`;

  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="-10 -10 276 276"
      width={size}
      height={size}
      role="img"
      aria-label="Google Chrome"
      {...props}
    >
      <linearGradient id={green} x1="145" x2="34" y1="253" y2="61" gradientUnits="userSpaceOnUse">
        <stop offset="0" stopColor="#1e8e3e" />
        <stop offset="1" stopColor="#34a853" />
      </linearGradient>
      <linearGradient id={yellow} x1="111" x2="222" y1="254" y2="62" gradientUnits="userSpaceOnUse">
        <stop offset="0" stopColor="#fcc934" />
        <stop offset="1" stopColor="#fbbc04" />
      </linearGradient>
      <linearGradient id={red} x1="17" x2="239" y1="80" y2="80" gradientUnits="userSpaceOnUse">
        <stop offset="0" stopColor="#d93025" />
        <stop offset="1" stopColor="#ea4335" />
      </linearGradient>
      <circle cx="128" cy="128" r="64" fill="#fff" />
      <path fill={`url(#${green})`} d="M96 183.4A63.7 63.7 0 0 1 72.6 160L17.2 64A128 128 0 0 0 128 256l55.4-96A64 64 0 0 1 96 183.4Z" />
      <path fill={`url(#${yellow})`} d="M192 128a63.7 63.7 0 0 1-8.6 32L128 256A128 128 0 0 0 238.9 64h-111a64 64 0 0 1 64 64Z" />
      <circle cx="128" cy="128" r="52" fill="#1a73e8" />
      <path fill={`url(#${red})`} d="M96 72.6a63.7 63.7 0 0 1 32-8.6h110.8a128 128 0 0 0-221.7 0l55.5 96A64 64 0 0 1 96 72.6Z" />
    </svg>
  );
}
