import type {
  AnchorHTMLAttributes,
  ButtonHTMLAttributes,
  FC,
  PropsWithChildren,
} from "react";
import { Link, type To } from "react-router";
import { tv, type VariantProps } from "tailwind-variants";

import { Slot } from "../Slot";

export const button = tv({
  base: "flex items-center justify-center gap-[6px] px-3 py-2 rounded-full cursor-pointer text-sm font-medium h-8 focus:outline-none whitespace-nowrap w-max",
  variants: {
    variant: {
      primary:
                "bg-primary text-invert hover:bg-primary-hover shadow-base hover:shadow-hover active:bg-primary-pressed",
      secondary:
                "bg-secondary text-txt-primary hover:bg-secondary-hover shadow-base hover:shadow-hover active:bg-secondary-pressed border border-border-low",
      tertiary:
                "bg-tertiary text-txt-primary hover:bg-tertiary-hover active:bg-tertiary-pressed",
      quaternary:
                "bg-highlight text-txt-primary hover:bg-highlight-hover active:bg-highlight-pressed",
      danger: "bg-danger-plain text-txt-invert hover:bg-danger-hover shadow-base hover:shadow-hover active:bg-danger-pressed border border-border-danger",
    },
    disabled: {
      true: "opacity-60 cursor-default",
    },
    empty: {
      true: "p-2 size-8",
    },
  },
  defaultVariants: {
    variant: "primary",
    icon: false,
    children: undefined,
  },
});

type Props = PropsWithChildren<
  {
    icon?: FC<{ size: number; className?: string }>;
    iconAfter?: FC<{ size: number; className?: string }>;
    disabled?: boolean;
    onClick?: () => void;
    variant?:
      | "primary"
      | "secondary"
      | "tertiary"
      | "quaternary"
      | "danger";
    to?: To;
    asChild?: boolean;
  } & VariantProps<typeof button>
>
& (
  | ButtonHTMLAttributes<HTMLButtonElement>
  | AnchorHTMLAttributes<HTMLAnchorElement>
    );

export const Button = (props: Props) => {
  const {
    icon: IconComponent,
    iconAfter: IconAfterComponent,
    children,
    onClick,
    variant,
    asChild,
    ...componentProps
  } = props;
  if (asChild) {
    return (
      <Slot
        {...componentProps}
        className={button({ ...props, variant, empty: !children })}
      >
        {children}
      </Slot>
    );
  }

  const Component = props.to ? Link : "button";
  return (
  // @ts-expect-error Component is too dynamic
    <Component
      {...componentProps}
      onClick={onClick}
      className={button({ ...props, variant, empty: !children })}
    >
      {IconComponent && <IconComponent size={16} className="flex-none" />}
      {children}
      {IconAfterComponent && (
        <IconAfterComponent size={16} className="flex-none" />
      )}
    </Component>
  );
};
