import type {
    AnchorHTMLAttributes,
    ButtonHTMLAttributes,
    FC,
    PropsWithChildren,
} from "react";
import { Link } from "react-router";
import { tv, type VariantProps } from "tailwind-variants";

export const button = tv({
    base: "flex items-center justify-center gap-[6px] px-3 py-2 rounded-full cursor-pointer text-sm font-medium h-8 focus:outline-none",
    variants: {
        variant: {
            primary:
                "bg-primary text-invert hover:bg-primary-hover shadow-base hover:shadow-hover active:bg-primary-pressed focus:shadow-focus",
            secondary:
                "bg-secondary text-txt-primary hover:bg-secondary-hover shadow-base hover:shadow-hover active:bg-secondary-pressed focus:shadow-focus border border-border-low",
            tertiary:
                "bg-tertiary text-txt-primary hover:bg-tertiary-hover active:bg-tertiary-pressed focus:shadow-focus",
            quaternary:
                "bg-highlight text-txt-primary hover:bg-highlight-hover active:bg-highlight-pressed",
            danger: "bg-danger-plain text-invert hover:bg-danger-hover shadow-base hover:shadow-hover active:bg-danger-pressed focus:shadow-focus border border-border-danger",
        },
        disabled: {
            true: "opacity-60 cursor-default",
        },
        empty: {
            true: "p-2 size-8",
        },
    },
    defaultVariants: {
        variant: "primary",
        icon: false,
        children: undefined,
    },
});

type Props = PropsWithChildren<
    {
        icon?: FC<{ size: number; className?: string }>;
        iconAfter?: FC<{ size: number; className?: string }>;
        disabled?: boolean;
        onClick?: () => void;
        variant?:
            | "primary"
            | "secondary"
            | "tertiary"
            | "quaternary"
            | "danger";
        to?: string;
    } & VariantProps<typeof button>
> &
    (
        | ButtonHTMLAttributes<HTMLButtonElement>
        | AnchorHTMLAttributes<HTMLAnchorElement>
    );

export const Button = (props: Props) => {
    const Component = props.to ? Link : "button";
    const {
        icon: IconComponent,
        iconAfter: IconAfterComponent,
        children,
        onClick,
        ...componentProps
    } = props;
    return (
        // @ts-expect-error Component is too dynamic
        <Component
            {...componentProps}
            onClick={onClick}
            className={button({ ...props, empty: !children })}
        >
            {IconComponent && <IconComponent size={16} className="flex-none" />}
            {children}
            {IconAfterComponent && (
                <IconAfterComponent size={16} className="flex-none" />
            )}
        </Component>
    );
};
