import type { InputHTMLAttributes } from "react";
import { tv } from "tailwind-variants";
import { input } from "../Input/Input";

type Props = InputHTMLAttributes<HTMLTextAreaElement>;

export const textarea = tv({
    base: input({ class: "min-h-20" }),
});

console.log(input({ class: "min-h-20" }));

export function Textarea(props: Props) {
    return <textarea className={textarea(props)} {...props} />;
}
