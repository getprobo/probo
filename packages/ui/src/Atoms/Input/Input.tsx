import type { InputHTMLAttributes } from "react";
import { tv } from "tailwind-variants";

type Props = InputHTMLAttributes<HTMLInputElement>;

export const input = tv({
    base: "py-[6px] bg-secondary border border-border-mid rounded-[10px] hover:border-border-strong focus:shadow-focus text-sm px-3",
});

export function Input(props: Props) {
    return <input className={input(props)} {...props} />;
}
