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

import { useId } from "react";

import type { IconProps } from "./type";

// Public-domain specification drawing (Wikimedia, Flag of the United
// States.svg): 1.9:1, Old Glory red and blue, fifty stars from one path.
const US_VIEW_WIDTH = 7410;
const US_VIEW_HEIGHT = 3900;

export function IconFlagUS({ size = 24, className }: IconProps) {
  const id = useId().replace(/:/g, "");
  const star = `${id}-s`;
  const row = `${id}-s4`;
  const column = `${id}-s5`;
  const block = `${id}-s9`;
  const field = `${id}-s18`;

  return (
    <svg
      width={size * US_VIEW_WIDTH / US_VIEW_HEIGHT}
      height={size}
      viewBox={`0 0 ${US_VIEW_WIDTH} ${US_VIEW_HEIGHT}`}
      className={className}
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M0,0h7410v3900H0" fill="#b31942" />
      <path d="M0,450H7410m0,600H0m0,600H7410m0,600H0m0,600H7410m0,600H0" stroke="#FFF" strokeWidth="300" />
      <path d="M0,0h2964v2100H0" fill="#0a3161" />
      <g fill="#FFF">
        <g id={field}>
          <g id={block}>
            <g id={column}>
              <g id={row}>
                <path id={star} d="M247,90 317.534230,307.082039 132.873218,172.917961H361.126782L176.465770,307.082039z" />
                <use href={`#${star}`} y="420" />
                <use href={`#${star}`} y="840" />
                <use href={`#${star}`} y="1260" />
              </g>
              <use href={`#${star}`} y="1680" />
            </g>
            <use href={`#${row}`} x="247" y="210" />
          </g>
          <use href={`#${block}`} x="494" />
        </g>
        <use href={`#${field}`} x="988" />
        <use href={`#${block}`} x="1976" />
        <use href={`#${column}`} x="2470" />
      </g>
    </svg>
  );
}
