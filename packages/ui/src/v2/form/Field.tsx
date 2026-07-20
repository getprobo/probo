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

import { cloneElement, isValidElement, type ReactElement, type ReactNode, useId } from "react";

import { field } from "./variants";

export type FieldProps = {
  // Text shown above the control, associated with it via `htmlFor`/`id`.
  label?: ReactNode;
  // Validation / server error shown below the control and linked to it via
  // `aria-describedby` so assistive technology announces it.
  error?: ReactNode;
  className?: string;
  // A single form control (TextField, Textarea, …). It receives an injected
  // `id`, plus `aria-describedby`/`aria-invalid` when an error is present.
  children: ReactNode;
};

// Vertical label + control + error grouping for form dialogs. Unlike a raw
// <label> wrapper, the label associates with the control by id, so controls
// whose root is a <div> (and multi-element controls) remain valid and clicks
// never activate an unintended descendant.
export function Field(props: FieldProps) {
  const { label, error, className, children } = props;
  const { root, labelText, error: errorSlot } = field();

  const generatedId = useId();
  const errorId = useId();

  const child = isValidElement(children)
    ? (children as ReactElement<Record<string, unknown>>)
    : null;
  const existingId = typeof child?.props.id === "string" ? child.props.id : undefined;
  const controlId = existingId ?? generatedId;

  const control = child
    ? cloneElement(child, {
        "id": controlId,
        "aria-describedby": error != null ? errorId : undefined,
        "aria-invalid": error != null ? true : undefined,
      })
    : children;

  return (
    <div className={root({ className })}>
      {label != null && (
        <label htmlFor={controlId} className={labelText()}>
          {label}
        </label>
      )}
      {control}
      {error != null && (
        <p id={errorId} className={errorSlot()}>
          {error}
        </p>
      )}
    </div>
  );
}
