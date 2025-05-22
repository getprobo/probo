import type { ComponentProps, FC, PropsWithChildren, ReactNode } from "react";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { clsx } from "clsx";
import type { IconProps } from "../Icons/type.ts";
import { tv } from "tailwind-variants";
import { IconDotGrid1x3Horizontal } from "../Icons/IconDotGrid1x3Horizontal.tsx";
import { Button } from "../Button/Button.tsx";

type Props = PropsWithChildren<{
    toggle?: ReactNode;
    className?: string;
    open?: boolean;
}>;

export const dropdown = tv({
    base: "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 origin-[--radix-dropdown-menu-content-transform-origin] z-50 p-2 shadow-mid min-w-[8rem] bg-level-1 overflow-y-auto overflow-x-hidden rounded-2xl border-border-low",
});

export function Dropdown({ children, toggle, className, open }: Props) {
    return (
        <DropdownMenu.Root open={open}>
            {toggle && (
                <DropdownMenu.Trigger asChild>{toggle}</DropdownMenu.Trigger>
            )}
            <DropdownMenu.Portal>
                <DropdownMenu.Content
                    className={dropdown({ className })}
                    sideOffset={5}
                >
                    {children}
                </DropdownMenu.Content>
            </DropdownMenu.Portal>
        </DropdownMenu.Root>
    );
}

export function DropdownSeparator({ className }: { className?: string }) {
    return (
        <DropdownMenu.Separator
            className={clsx("h-[1px] bg-border-low my-2", className)}
        />
    );
}

type DropdownItemProps = PropsWithChildren<{
    className?: string;
    icon?: FC<IconProps>;
    asChild?: boolean;
    variant?: "primary" | "tertiary" | "danger";
}> &
    ComponentProps<typeof DropdownMenu.Item>;

export const dropdownItem = tv({
    base: "text-txt-primary flex items-center gap-2 hover:bg-tertiary-hover active:bg-tertiary-pressed data-active-item:bg-tertiary-pressed cursor-pointer p-2",
    variants: {
        variant: {
            primary: "text-txt-primary",
            tertiary: "text-txt-tertiary",
            danger: "text-txt-danger",
        },
    },
    defaultVariants: {
        variant: "primary",
    },
});

export function ActionDropdown(
    props: Omit<Props, "toggle"> & {
        variant?: ComponentProps<typeof Button>["variant"];
    },
) {
    return (
        <Dropdown
            {...props}
            toggle={
                <Button
                    variant={props.variant ?? "tertiary"}
                    icon={IconDotGrid1x3Horizontal}
                />
            }
        />
    );
}

export function DropdownItem({
    icon: IconComponent,
    variant,
    className,
    children,
    asChild,
    ...props
}: DropdownItemProps) {
    return (
        <DropdownMenu.Item
            {...props}
            className={clsx(dropdownItem({ variant }), className)}
            asChild={asChild}
        >
            {asChild ? (
                children
            ) : (
                <>
                    {IconComponent && (
                        <IconComponent size={16} className="flex-none" />
                    )}
                    {children}
                </>
            )}
        </DropdownMenu.Item>
    );
}
