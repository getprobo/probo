import type { HTMLAttributes } from "react";
import { tv } from "tailwind-variants";

type Props = {
    variant?: "success" | "warning" | "danger" | "info" | "neutral" | "outline";
} & HTMLAttributes<HTMLDivElement>;

const badge = tv({
    base: "text-xs font-medium py-[2px] px-[6px] rounded-lg w-max",
    variants: {
        variant: {
            success: "bg-success text-txt-success",
            warning: "bg-warning text-txt-warning",
            danger: "bg-danger text-txt-danger",
            info: "bg-info text-txt-info",
            neutral: "bg-subtle text-txt-secondary",
            outline: "text-txt-tertiary border border-border-low",
        },
    },
});

export function Badge(props: Props) {
    return <div className={badge(props)} {...props} />;
}
