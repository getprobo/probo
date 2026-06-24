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

export function IconCollapse({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M8.33235 2.66667H5C2.97496 2.66667 1.33333 4.3083 1.33333 6.33334V17.6667C1.33333 19.6917 2.97496 21.3333 5 21.3333H19C21.025 21.3333 22.6667 19.6917 22.6667 17.6667V6.33334C22.6667 4.3083 21.025 2.66667 19 2.66667H8.33432C8.33399 2.66667 8.33267 2.66667 8.33235 2.66667C8.33267 2.66667 8.33202 2.66667 8.33235 2.66667ZM9.33333 19.3333V4.66667H19C19.9205 4.66667 20.6667 5.41286 20.6667 6.33334V17.6667C20.6667 18.5871 19.9205 19.3333 19 19.3333H9.33333ZM7.33333 19.3333H5C4.07953 19.3333 3.33333 18.5871 3.33333 17.6667V6.33334C3.33333 5.41286 4.07953 4.66667 5 4.66667H7.33333V19.3333ZM17.0404 7.95956C17.431 8.35009 17.431 8.98325 17.0404 9.37378L14.4142 12L17.0404 14.6262C17.431 15.0168 17.431 15.6499 17.0404 16.0404C16.6499 16.431 16.0168 16.431 15.6262 16.0404L12.2929 12.7071C11.9024 12.3166 11.9024 11.6834 12.2929 11.2929L15.6262 7.95956C16.0168 7.56904 16.6499 7.56904 17.0404 7.95956Z" fill="currentColor" />
    </svg>
  );
}
