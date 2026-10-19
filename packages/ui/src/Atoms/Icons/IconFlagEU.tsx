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

import type { IconProps } from "./type";

// Star polygons from the European Union's published colour SVG
// (EU-emblem_RGB.svg). The viewBox is that file's blue field; the
// published file adds a white margin around it.
const EU_STARS = [
  "141.53 46.86 147.51 42.5 153.5 46.86 151.22 39.8 157.3 35.44 149.8 35.44 147.51 28.29 145.22 35.44 137.73 35.44 143.8 39.8 141.53 46.86",
  "110.56 55.19 116.54 50.82 122.52 55.19 120.25 48.12 126.32 43.76 118.83 43.76 116.54 36.6 114.25 43.76 106.76 43.76 112.83 48.12 110.56 55.19",
  "93.88 59.31 91.59 66.47 84.1 66.47 90.18 70.83 87.9 77.9 93.88 73.53 99.87 77.9 97.59 70.83 103.66 66.47 96.18 66.47 93.88 59.31",
  "85.57 104.46 91.55 108.83 89.28 101.76 95.35 97.4 87.86 97.4 85.57 90.24 83.28 97.41 75.78 97.4 81.86 101.76 79.59 108.83 85.57 104.46",
  "96.18 128.42 93.88 121.26 91.59 128.43 84.1 128.42 90.18 132.79 87.9 139.85 93.88 135.47 99.87 139.85 97.59 132.79 103.66 128.42 96.18 128.42",
  "118.88 151.12 116.59 143.97 114.3 151.13 106.8 151.12 112.88 155.49 110.61 162.55 116.59 158.19 122.57 162.55 120.3 155.49 126.37 151.12 118.88 151.12",
  "149.81 159.35 147.52 152.19 145.23 159.36 137.73 159.35 143.81 163.72 141.53 170.78 147.52 166.41 153.5 170.78 151.22 163.72 157.3 159.35 149.81 159.35",
  "180.74 151.12 178.45 143.97 176.15 151.13 168.67 151.12 174.74 155.49 172.46 162.55 178.45 158.19 184.43 162.55 182.15 155.49 188.23 151.12 180.74 151.12",
  "203.45 128.42 201.15 121.26 198.86 128.43 191.37 128.42 197.45 132.79 195.17 139.85 201.15 135.47 207.14 139.85 204.86 132.79 210.93 128.42 203.45 128.42",
  "219.16 97.31 211.67 97.31 209.37 90.15 207.08 97.32 199.59 97.31 205.67 101.68 203.39 108.74 209.37 104.36 215.36 108.74 213.08 101.68 219.16 97.31",
  "195.17 77.8 201.15 73.44 207.14 77.8 204.86 70.74 210.93 66.38 203.45 66.38 201.15 59.22 198.86 66.39 191.37 66.38 197.45 70.74 195.17 77.8",
  "178.53 36.61 176.25 43.77 168.75 43.76 174.82 48.13 172.55 55.19 178.54 50.82 184.52 55.19 182.25 48.13 188.32 43.76 180.82 43.76 178.53 36.61",
];

const EU_VIEW_WIDTH = 283.6;
const EU_VIEW_HEIGHT = 189.07;

export function IconFlagEU({ size = 24, className }: IconProps) {
  return (
    <svg
      width={size * EU_VIEW_WIDTH / EU_VIEW_HEIGHT}
      height={size}
      viewBox={`5.67 5.67 ${EU_VIEW_WIDTH} ${EU_VIEW_HEIGHT}`}
      className={className}
      xmlns="http://www.w3.org/2000/svg"
    >
      <rect x="5.67" y="5.67" width={EU_VIEW_WIDTH} height={EU_VIEW_HEIGHT} fill="#039" />
      {EU_STARS.map(points => (
        <polygon key={points} points={points} fill="#fc0" />
      ))}
    </svg>
  );
}
