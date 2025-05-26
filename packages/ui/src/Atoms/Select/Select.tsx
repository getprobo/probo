import {
    Root,
    Trigger,
    Value,
    Icon,
    Portal,
    Content,
    Viewport,
    Item,
    ItemText,
} from "@radix-ui/react-select";
import * as ScrollArea from "@radix-ui/react-scroll-area";
import { input } from "../Input/Input.tsx";
import { IconChevronGrabberVertical } from "../Icons/IconChevronGrabberVertical.tsx";
import { tv } from "tailwind-variants";
import { Children, type ComponentProps, type PropsWithChildren } from "react";

type Props = PropsWithChildren<
    {
        id?: string;
        placeholder?: string;
        onValueChange?: (s: string) => void;
        variant?: "default" | "editor" | "dashed";
        invalid?: boolean;
        disabled?: boolean;
        className?: string;
    } & Omit<ComponentProps<typeof Root>, "onChange">
>;

const select = tv({
    slots: {
        trigger:
            "flex justify-between items-center data-placeholder:text-txt-tertiary whitespace-nowrap cursor-pointer *:first:contents",
        content:
            "z-100 shadow-mid rounded-[10px] bg-level-1 p-1 animate-in fade-in slide-in-from-top-2 overflow-y-auto overflow-y-auto",
        icon: "-mr-1",
    },
    variants: {
        invalid: {
            true: {
                trigger: "border-border-danger",
            },
        },
        variant: {
            dashed: {
                trigger: input({
                    class: "w-full gap-4 border-dashed pointer-events-none",
                }),
            },
            editor: {
                trigger:
                    "bg-highlight hover:bg-highlight-hover active:bg-highlight-pressed text-txt-primary text-sm px-[10px] py-[6px] rounded-lg w-max gap-2",
            },
            default: {
                trigger: input({ class: "w-full gap-4 " }),
            },
        },
    },
    compoundVariants: [
        {
            invalid: true,
            variant: "default",
            class: {
                trigger: "border-border-danger",
            },
        },
    ],
    defaultVariants: {
        variant: "default",
    },
});

export function Select({
    placeholder,
    children,
    onValueChange,
    value,
    ...props
}: Props) {
    const { trigger, content, icon } = select({
        ...props,
    });

    return (
        <Root onValueChange={onValueChange} value={value}>
            <Trigger {...props} className={trigger()}>
                <Value placeholder={placeholder} />
                <Icon className={icon()}>
                    <IconChevronGrabberVertical size={16} />
                </Icon>
            </Trigger>
            <Portal>
                <Content
                    className={content()}
                    position="popper"
                    sideOffset={5}
                    style={{
                        minWidth: "var(--radix-select-trigger-width)",
                        maxHeight:
                            "var(--radix-select-content-available-height)",
                    }}
                >
                    <ScrollArea.Root className="ScrollAreaRoot" type="auto">
                        <Viewport asChild>
                            <ScrollArea.Viewport className="ScrollAreaViewport">
                                {children}
                            </ScrollArea.Viewport>
                        </Viewport>
                    </ScrollArea.Root>
                </Content>
            </Portal>
        </Root>
    );
}

export function Option({ children, ...props }: ComponentProps<typeof Item>) {
    const hasSingleChildren = Children.count(children) <= 1;
    return (
        <Item
            {...props}
            className={
                "flex gap-2 items-center h-8 text-sm font-medium text-txt-primary hover:bg-tertiary-hover active:bg-tertiary-pressed cursor-pointer px-[10px]"
            }
        >
            <ItemText asChild>
                <span
                    className={
                        hasSingleChildren
                            ? "text-ellipsis overflow-hidden"
                            : "flex gap-2 items-center"
                    }
                >
                    {children}
                </span>
            </ItemText>
        </Item>
    );
}
