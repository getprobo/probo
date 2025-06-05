import type { ComponentProps } from "react";
import { tv } from "tailwind-variants";
import { Label as RadixLabel } from "@radix-ui/react-label";

type Props = ComponentProps<typeof RadixLabel>;

const label = tv({
    base: "block text-sm font-medium text-txt-primary mb-[6px]",
});

export function Label({ ...props }: Props) {
    return <RadixLabel {...props} className={label(props)} />;
}
