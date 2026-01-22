import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { clsx } from "clsx";
import type { ComponentProps, FC, PropsWithChildren, ReactNode } from "react";
import { tv } from "tailwind-variants";

import { Button } from "../Button/Button";
import { IconDotGrid1x3Horizontal } from "../Icons";
import type { IconProps } from "../Icons/type";

type Props = PropsWithChildren<{
  toggle?: ReactNode;
  className?: string;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}>;

export const dropdown = tv({
  base: "z-50 p-2 shadow-mid min-w-[8rem] bg-level-1 overflow-y-auto overflow-x-hidden rounded-2xl border-border-low",
});

export function Dropdown({ children, toggle, className, open, onOpenChange }: Props) {
  return (
    <DropdownMenu.Root open={open} onOpenChange={onOpenChange}>
      {toggle && (
        <DropdownMenu.Trigger asChild>{toggle}</DropdownMenu.Trigger>
      )}
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          style={{
            maxHeight:
                            "var(--radix-dropdown-menu-content-available-height)",
          }}
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
}>
& ComponentProps<typeof DropdownMenu.Item>;

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
      toggle={(
        <Button
          className="inline-flex isolate z-5"
          variant={props.variant ?? "tertiary"}
          icon={IconDotGrid1x3Horizontal}
        />
      )}
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
      {asChild
        ? (
            children
          )
        : (
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
