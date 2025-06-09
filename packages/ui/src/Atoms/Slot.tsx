import clsx from "clsx";
import React from "react";
import { cloneElement, isValidElement, Children } from "react";
import type { ReactNode } from "react";

export type AsChildProps<DefaultElementProps> =
    | ({ asChild?: false } & DefaultElementProps)
    | ({ asChild: true; children: ReactNode } & DefaultElementProps);

export function Slot({
    children,
    ...props
}: React.HTMLAttributes<HTMLElement> & {
    children?: ReactNode;
}) {
    if (isValidElement(children)) {
        return cloneElement(children, {
            ...props,
            // @ts-expect-error children.props can be undefined it's ok
            ...children.props,
            // @ts-expect-error children.props.className can be undefined it's ok
            className: clsx(props.className, children.props.className),
        });
    }
    if (Children.count(children) > 1) {
        Children.only(null);
    }
    return null;
}
