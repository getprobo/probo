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

import { Children, type ReactNode } from "react";

import { Label } from "../Label/Label";

type Props = {
  label: string;
  id?: string;
  children: ReactNode;
  error?: string;
};

export function PropertyRow({ id, label, children, error, ...props }: Props) {
  const [firstChild, ...restChildren] = Children.toArray(children);

  return (
    <div className="py-3 border-b border-border-low space-y-2" {...props}>
      <div className="flex items-center justify-between gap-4">
        <Label
          className="text-sm text-txt-secondary font-medium mb-0"
          htmlFor={id}
        >
          {label}
        </Label>
        <div className="min-w-0">{firstChild}</div>
      </div>
      {error && <div className="text-xs text-txt-danger">{error}</div>}
      {restChildren}
    </div>
  );
}
