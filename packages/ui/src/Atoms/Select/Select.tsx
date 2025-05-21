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
import type { ComponentProps, PropsWithChildren } from "react";

type Props = PropsWithChildren<
    {
        placeholder?: string;
        onChange?: (s: string) => void;
        empty?: boolean;
    } & Omit<ComponentProps<typeof Trigger>, "onChange">
>;

const select = tv({
    slots: {
        trigger: input({
            className:
                "flex justify-between items-center gap-4 data-placeholder:text-txt-tertiary w-full whitespace-nowrap",
        }),
        content:
            "z-100 shadow-mid rounded-[10px] bg-level-1 p-1 animate-in fade-in slide-in-from-top-2 overflow-y-auto overflow-y-auto",
        option: "flex items-center gap-2 h-8 text-sm font-medium text-txt-primary hover:bg-tertiary-hover active:bg-tertiary-pressed cursor-pointer px-[10px]",
    },
    variants: {
        empty: {
            true: {
                trigger: "border-dashed pointer-events-none",
            },
        },
    },
});

const { trigger, option, content } = select();

export function Select({
    placeholder,
    children,
    onChange,
    empty,
    ...props
}: Props) {
    return (
        <Root onValueChange={onChange}>
            <Trigger className={trigger({ ...props, empty })}>
                <div className="text-ellipsis overflow-hidden">
                    <Value placeholder={placeholder} />
                </div>
                <Icon className="SelectIcon">
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
    return (
        <Item {...props} className={option(props)}>
            <ItemText>{children}</ItemText>
        </Item>
    );
}
