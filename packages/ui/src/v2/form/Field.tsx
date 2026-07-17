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

import type { ReactNode } from "react";

import { field } from "./variants";

export type FieldProps = {
  // Text shown above the control. The control is nested inside the <label> so
  // the association is implicit (no htmlFor / id threading required).
  label?: ReactNode;
  // Validation / server error shown below the control.
  error?: ReactNode;
  className?: string;
  children: ReactNode;
};

// Vertical label + control + error grouping for form dialogs.
export function Field(props: FieldProps) {
  const { label, error, className, children } = props;
  const { root, label: labelSlot, labelText, error: errorSlot } = field();

  return (
    <div className={root({ className })}>
      <label className={labelSlot()}>
        {label != null && <span className={labelText()}>{label}</span>}
        {children}
      </label>
      {error != null && <p className={errorSlot()}>{error}</p>}
    </div>
  );
}
