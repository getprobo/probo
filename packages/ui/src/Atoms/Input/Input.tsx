import type { InputHTMLAttributes } from "react";
import { tv } from "tailwind-variants";

type Props = {
    invalid?: boolean;
    disabled?: boolean;
} & InputHTMLAttributes<HTMLInputElement>;

export const input = tv({
    base: "py-[6px] bg-secondary border border-border-mid rounded-[10px] hover:border-border-strong focus:shadow-focus text-sm px-3 w-full bg-secondary disabled:bg-transparent",
    variants: {
        invalid: {
            true: "border-border-danger",
        },
    },
});

export function Input({ invalid, ...props }: Props) {
    return <input aria-invalid={invalid} className={input(props)} {...props} />;
}
