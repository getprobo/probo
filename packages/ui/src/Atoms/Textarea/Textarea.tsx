import type { TextareaHTMLAttributes } from "react";
import { tv } from "tailwind-variants";
import { input } from "../Input/Input";

type Props = TextareaHTMLAttributes<HTMLTextAreaElement>;

export const textarea = tv({
    base: input({ class: "min-h-20" }),
});

export function Textarea(props: Props) {
    return <textarea {...props} className={textarea(props)} />;
}
