import { clsx } from "clsx";
import type { PropsWithChildren, ReactNode } from "react";
import { Children, cloneElement, isValidElement } from "react";

export type AsChildProps<DefaultElementProps>
  = | ({ asChild?: false } & DefaultElementProps)
    | ({ asChild: true; children: ReactNode } & DefaultElementProps);

export function Slot({
  children,
  ...props
}: PropsWithChildren<{ className?: string | undefined } & Record<string, unknown>>) {
  if (isValidElement<{ className?: string | undefined } & Record<string, unknown>>(children)) {
    return cloneElement(children, {
      ...props,
      className: clsx(props.className, children.props.className),
    });
  }
  if (Children.count(children) > 1) {
    Children.only(null);
  }
  return null;
}
