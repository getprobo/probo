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
import {
    Children,
    isValidElement,
    type ComponentProps,
    type PropsWithChildren,
    type ReactNode,
} from "react";

type Props = PropsWithChildren<
    {
        id?: string;
        placeholder?: string;
        onValueChange?: (s: string) => void;
        variant?: "default" | "editor" | "dashed";
        invalid?: boolean;
        disabled?: boolean;
    } & Omit<ComponentProps<typeof Root>, "onChange">
>;

const select = tv({
    slots: {
        trigger:
            "flex justify-between items-center data-placeholder:text-txt-tertiary whitespace-nowrap cursor-pointer",
        content:
            "z-100 shadow-mid rounded-[10px] bg-level-1 p-1 animate-in fade-in slide-in-from-top-2 overflow-y-auto overflow-y-auto",
        option: "flex items-center h-8 text-sm font-medium text-txt-primary hover:bg-tertiary-hover active:bg-tertiary-pressed cursor-pointer px-[10px]",
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

const { trigger, option, content, icon } = select();

/**
 * To display the selected value we need to find the selected option among children
 */
const findSelectedOption = (
    children: unknown[],
    condition: (c: { props: Record<string, unknown> }) => boolean,
): ReactNode => {
    if (children.length === 1) {
        debugger;
    }
    const selectedOptions = children.find(
        // @ts-expect-error We know that the children are ReactElements with props
        (c) => isValidElement(c) && condition(c),
    );
    if (!isValidElement(selectedOptions)) {
        return null;
    }
    return (selectedOptions.props as { children: ReactNode }).children;
};

export function Select({
    placeholder,
    children,
    onValueChange,
    value,
    ...props
}: Props) {
    const childrenArr = Children.toArray(children);
    const valueNode = value
        ? findSelectedOption(
              childrenArr,
              (c) => c.props.value?.toString() === value?.toString(),
          )
        : undefined;
    console.log(childrenArr, valueNode);
    return (
        <Root onValueChange={onValueChange} value={value}>
            <Trigger className={trigger({ ...props })} {...props}>
                <div className="text-ellipsis overflow-hidden">
                    <Value placeholder={placeholder}>
                        {valueNode ? (
                            <span className="flex items-center gap-2">
                                {valueNode}
                            </span>
                        ) : null}
                    </Value>
                </div>
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
    return (
        <Item {...props} className={option(props)}>
            <ItemText asChild>
                <span className="flex items-center gap-2">{children}</span>
            </ItemText>
        </Item>
    );
}
