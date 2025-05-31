import type { PropsWithChildren } from "react";
import { tv } from "tailwind-variants";
import { Slot } from "../Slot";

type Props = PropsWithChildren<{
    padded?: boolean;
    className?: string;
    asChild?: boolean;
}>;

const card = tv({
    base: "border border-border-low bg-level-1 rounded-2xl",
    variants: {
        padded: {
            true: "p-4",
        },
    },
});

export function Card({ padded = false, children, className, asChild }: Props) {
    const Component = asChild ? Slot : "div";
    return (
        <Component className={card({ padded, className })}>
            {children}
        </Component>
    );
}
