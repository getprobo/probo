import { type FC, type InputHTMLAttributes, useState } from "react";
import { tv } from "tailwind-variants";

import type { IconProps } from "../Icons/type";

type Props = {
  invalid?: boolean;
  disabled?: boolean;
  icon?: FC<IconProps>;
  onValueChange?: (value: string) => void;
  variant?: "bordered" | "ghost" | "title";
} & InputHTMLAttributes<HTMLInputElement>;

export const input = tv({
  base: "w-full disabled:bg-transparent",
  variants: {
    invalid: {
      true: "border-border-danger",
    },
    variant: {
      bordered:
                "py-[6px] bg-secondary border border-border-mid rounded-[10px] hover:border-border-strong text-sm px-3 text-txt-primary",
      ghost: "text-base text-txt-secondary outline-none",
      title: "text-2xl font-semibold text-txt-primary outline-none",
    },
  },
  defaultVariants: {
    variant: "bordered",
  },
});

export function Input({
  invalid,
  icon: IconComponent,
  onValueChange,
  ...props
}: Props) {
  const [focus, setFocus] = useState(false);

  if (IconComponent) {
    return (
      <div
        className={input({
          className: "flex items-center gap-2",
        })}
        data-focus={focus}
      >
        <IconComponent size={16} className="text-txt-secondary" />
        <input
          {...props}
          onFocus={(e) => {
            setFocus(true);
            props.onFocus?.(e);
          }}
          onBlur={(e) => {
            setFocus(false);
            props.onBlur?.(e);
          }}
          aria-invalid={invalid}
          className="w-full outline-none"
          onChange={(e) => {
            onValueChange?.(e.currentTarget.value);
            props.onChange?.(e);
          }}
        />
      </div>
    );
  }
  return (
    <input
      {...props}
      aria-invalid={invalid}
      className={input(props)}
      onChange={(e) => {
        onValueChange?.(e.currentTarget.value);
        props.onChange?.(e);
      }}
    />
  );
}
