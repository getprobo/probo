import { useState } from "react";
import { tv } from "tailwind-variants";
import { IconCheckmark1 } from "../Icons";

type Props = {
    checked: boolean;
    onChange: (checked: boolean) => void;
};

const checkbox = tv({
    base: "size-4 border border-border-mid relative rounded-sm flex items-center justify-center",
    variants: {
        isFocused: {
            true: "shadow shadow-focus",
        },
        checked: {
            true: "bg-accent text-invert",
        },
    },
});

export function Checkbox({ checked, onChange }: Props) {
    const [isFocused, setFocus] = useState(false);
    return (
        <div className={checkbox({ isFocused, checked })}>
            <input
                className="absolute inset-0 opacity-0"
                type="checkbox"
                checked={checked}
                onChange={(e) => onChange(e.target.checked)}
                onFocus={() => setFocus(true)}
                onBlur={() => setFocus(false)}
            />
            {checked && <IconCheckmark1 size={12} />}
        </div>
    );
}
