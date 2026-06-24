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

export function IconRecKeyframe({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M9.17113 3.16424C10.7332 1.60215 13.2659 1.60215 14.828 3.16424L20.8353 9.17158C22.3974 10.7337 22.3974 13.2663 20.8353 14.8284L14.828 20.8358C13.2659 22.3979 10.7332 22.3979 9.17111 20.8358L3.16375 14.8284C1.60166 13.2663 1.60167 10.7337 3.16377 9.17158L9.17113 3.16424ZM8.24951 12C8.24951 9.92893 9.92844 8.25 11.9995 8.25C14.0706 8.25 15.7495 9.92893 15.7495 12C15.7495 14.0711 14.0706 15.75 11.9995 15.75C9.92844 15.75 8.24951 14.0711 8.24951 12Z" fill="currentColor" />
    </svg>
  );
}
