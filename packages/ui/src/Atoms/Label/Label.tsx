import type { HTMLAttributes } from "react";
import { tv } from "tailwind-variants";

type Props = HTMLAttributes<HTMLLabelElement> & {
    htmlFor?: string;
};

const label = tv({
    base: "text-sm font-medium text-txt-primary",
});

export function Label({ ...props }: Props) {
    return <label {...props} className={label(props)} />;
}
