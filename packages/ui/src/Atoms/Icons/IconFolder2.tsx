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

export function IconFolder2({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M6 5C4.89543 5 4 5.89543 4 7V9.53512C4.58835 9.19478 5.27143 9 6 9H18C18.7286 9 19.4117 9.19478 20 9.53511V9C20 7.89542 19.1046 7 18 7H13.2426C12.1818 7 11.1643 6.57858 10.4142 5.82842L10.1716 5.58578C9.79654 5.21075 9.28778 5 8.75736 5H6ZM20 13C20 11.8954 19.1046 11 18 11H6C4.89542 11 4 11.8954 4 13V16C4 17.1046 4.89542 18 6 18H18C19.1046 18 20 17.1046 20 16V13ZM2 16C2 18.2092 3.79088 20 6 20H18C20.2092 20 22 18.2092 22 16V9C22 6.79088 20.2092 5 18 5H13.2426C12.7122 5 12.2035 4.78928 11.8284 4.41422L11.5859 4.17163C10.8357 3.42147 9.81822 3 8.75736 3H6C3.79087 3 2 4.79087 2 7V16Z" fill="currentColor" />
    </svg>
  );
}
