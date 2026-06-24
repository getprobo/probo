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

import { type ComponentProps, type ReactNode } from "react";
import { tv } from "tailwind-variants";

import { Input } from "../../Atoms/Input/Input";
import { Label } from "../../Atoms/Label/Label";
import { Select } from "../../Atoms/Select/Select";
import { Textarea } from "../../Atoms/Textarea/Textarea";

type BaseProps<T extends string, P> = {
  label?: string;
  help?: string;
  error?: string;
  onValueChange?: (s: string) => void;
  type?: T;
  children?: ReactNode;
} & P;

type FieldValue = string | readonly string[] | number;

type Props<T extends FieldValue = string>
  = | BaseProps<never, ComponentProps<typeof Input>>
    | BaseProps<"text", ComponentProps<typeof Input>>
    | BaseProps<"email", ComponentProps<typeof Input>>
    | BaseProps<"url", ComponentProps<typeof Input>>
    | BaseProps<"password", ComponentProps<typeof Input>>
    | BaseProps<"textarea", ComponentProps<typeof Textarea>>
    | BaseProps<"number", ComponentProps<typeof Input>>
    | BaseProps<"select", ComponentProps<typeof Select<T>>>;

const field = tv({
  slots: {
    base: "flex flex-col",
    label: "mb-[6px]",
    help: "text-xs text-txt-tertiary mt-1",
  },
});

const { base: baseClass, label: labelClass, help: helpClass } = field();

export function Field<T extends FieldValue = string>(props: Props<T>) {
  const showHelp = props.help && !props.error;
  const childrenAsInput = !props.type && props.children;
  return (
    <div className={baseClass()}>
      {props.label && (
        <Label htmlFor={props.name} className={labelClass()}>
          {props.label}
        </Label>
      )}
      {childrenAsInput ? props.children : getInput(props)}
      {showHelp && <span className={helpClass()}>{props.help}</span>}
      {props.error && (
        <span className="text-txt-danger text-sm mt-1">
          {props.error}
        </span>
      )}
    </div>
  );
}

function getInput<T extends FieldValue = string>(props: Props<T>) {
  const { error, onValueChange, type, ...restProps } = props;
  const baseProps = {
    ["aria-invalid"]: !!error,
    name: props.name,
    id: props.name,
    placeholder: props.placeholder,
  };
  switch (type) {
    case "select":
      return (
      // @ts-expect-error Select is too dynamic
        <Select<T>
          {...baseProps}
          {...restProps}
          onValueChange={onValueChange}
        />
      );
    case "textarea":
      return (
        <Textarea
          // @ts-expect-error Textarea is too dynamic
          onChange={e => onValueChange?.(e.target.value)}
          {...baseProps}
          {...restProps}
        />
      );
    default:
      return (
        <Input
          type={type}
          // @ts-expect-error Input is too dynamic
          onChange={e => onValueChange?.(e.target.value)}
          {...baseProps}
          {...restProps}
        />
      );
  }
}
