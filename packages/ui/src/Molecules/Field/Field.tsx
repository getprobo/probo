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
