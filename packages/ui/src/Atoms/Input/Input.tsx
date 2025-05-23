import { type InputHTMLAttributes, type FC, useState } from "react";
import { tv } from "tailwind-variants";
import type { IconProps } from "../Icons/type";

type Props = {
    invalid?: boolean;
    disabled?: boolean;
    icon?: FC<IconProps>;
    onValueChange?: (value: string) => void;
} & InputHTMLAttributes<HTMLInputElement>;

export const input = tv({
    base: "py-[6px] bg-secondary border border-border-mid rounded-[10px] hover:border-border-strong focus:shadow-focus text-sm px-3 w-full bg-secondary disabled:bg-transparent focus:outline-none data-[focus=true]:shadow-focus",
    variants: {
        invalid: {
            true: "border-border-danger",
        },
    },
});

export function Input({
    invalid,
    icon: IconComponent,
    onValueChange,
    ...props
}: Props) {
    if (IconComponent) {
        const [focus, setFocus] = useState(false);
        return (
            <div
                className={input({
                    className: "flex items-center gap-2",
                })}
                data-focus={focus}
            >
                <IconComponent size={16} className="text-txt-secondary" />
                <input
                    onFocus={(e) => {
                        setFocus(true);
                        props.onFocus?.(e);
                    }}
                    onBlur={(e) => {
                        setFocus(false);
                        props.onBlur?.(e);
                    }}
                    aria-invalid={invalid}
                    className="w-full outline-none"
                    {...props}
                    onChange={(e) => {
                        onValueChange?.(e.currentTarget.value);
                        props.onChange?.(e);
                    }}
                />
            </div>
        );
    }
    return <input aria-invalid={invalid} className={input(props)} {...props} />;
}
