import type { HTMLAttributes } from "react";
import { tv } from "tailwind-variants";

import { Slot } from "../Slot";

type Props = {
  asChild?: boolean;
  variant?:
    | "success"
    | "warning"
    | "danger"
    | "info"
    | "neutral"
    | "outline"
    | "highlight";
  size?: "sm" | "md";
} & HTMLAttributes<HTMLElement>;

const badge = tv({
  base: "font-medium rounded-lg w-max flex gap-1 items-center group whitespace-nowrap",
  variants: {
    variant: {
      success: "bg-success text-txt-success",
      warning: "bg-warning text-txt-warning",
      danger: "bg-danger text-txt-danger",
      info: "bg-info text-txt-info",
      neutral: "bg-subtle text-txt-secondary",
      outline: "text-txt-tertiary border border-border-low",
      highlight: "bg-highlight text-txt-primary",
    },
    size: {
      sm: "text-xs py-[2px] px-[6px]",
      md: "text-sm py-[6px] px-2",
    },
  },
  defaultVariants: {
    variant: "neutral",
    size: "sm",
  },
});

export function Badge(props: Props) {
  const Component = props.asChild ? Slot : "div";
  const { className, size, variant, ...restProps } = props;
  return (
    <Component {...restProps} className={badge({ className, size, variant })} />
  );
}
