import * as Popover from "@radix-ui/react-popover";
import {
    type KeyboardEventHandler,
    type ReactNode,
    Suspense,
    useRef,
    useState,
} from "react";
import { Cell } from "../../Atoms/DataTable/DataTable.tsx";
import { Command } from "cmdk";
import { useTranslate } from "@probo/i18n";
import { Spinner } from "../../Atoms/Spinner/Spinner.tsx";
import { focusSiblingElement } from "@probo/helpers";

type Props<T> =
    | {
          type: "text";
          value?: T;
          blink?: boolean;
          onValueChange: (value: string) => void;
          itemRenderer?: undefined;
      }
    | {
          type: "select";
          items: T[] | (() => T[]);
          itemRenderer: (v: { item: T }) => ReactNode;
          value?: T;
          blink?: boolean;
          onValueChange: (value: T) => void;
      }
    | {
          type: "multiple";
          items: T[] | (() => T[]);
          itemRenderer: (v: { item: T; onRemove?: () => void }) => ReactNode;
          value?: T[];
          blink?: boolean;
          onValueChange: (value: T[]) => void;
      };

type PropsField<Type, T> = Props<T> & {
    type: Type;
    onOpenChange: (open: boolean) => void;
    padding: string;
    height: number;
};

function getKey<T>(item: T): string {
    if (
        item &&
        typeof item === "object" &&
        "id" in item &&
        typeof item.id === "string"
    ) {
        return item.id.toString();
    }
    if (typeof item === "string" || typeof item === "number") {
        return item.toString();
    }
    if (item === undefined) {
        return "";
    }
    console.error("Cannot compute a key from item", item);
    return "";
}

export function EditableCell<T>(props: Props<T>) {
    const [isOpen, setOpen] = useState(false);
    const [value, setValueState] = useState(props.value);
    const [height, setHeight] = useState<number | undefined>(undefined);
    const [padding, setPadding] = useState("12px");
    const td = useRef<HTMLTableCellElement>(null);
    const valueRef = useRef(value);
    const setValue = (value: T) => {
        valueRef.current = value;
        setValueState(value);
    };

    // When opening the popover, remember the height and padding of the cell
    const onOpenChange = (open: boolean) => {
        if (open) {
            setHeight(td.current?.offsetHeight ?? undefined);
            setPadding(getComputedStyle(td.current!).paddingLeft);
        } else {
            setHeight(undefined);
        }
        // Send the value when closing the popover
        if (!open && valueRef.current !== props.value) {
            // @ts-expect-error - cannot unpack value type
            props.onValueChange(valueRef.current);
        }
        setOpen(open);
    };

    const fieldProps = {
        height,
        onValueChange: setValue,
        onOpenChange,
        padding,
        value,
    } as any;

    const children = (() => {
        if (!value) {
            return "";
        }
        if (props.type === "select") {
            // @ts-expect-error TS cannot understand the link between props.type and value
            return props.itemRenderer({ item: value });
        }
        if (Array.isArray(value) && props.type === "multiple") {
            return <>{value.map((v) => props.itemRenderer({ item: v }))}</>;
        }
        return value as ReactNode;
    })();

    // Handle keyboard navigation inside the cells
    const onKeyDown: KeyboardEventHandler<HTMLButtonElement> = (e) => {
        if (e.key === "ArrowRight") {
            focusSiblingElement(1);
        } else if (e.key === "ArrowLeft") {
            focusSiblingElement(-1);
        } else if (e.key === "ArrowDown") {
            e.preventDefault();
            focusSiblingElement(td.current?.parentNode?.children.length ?? 0);
        } else if (e.key === "ArrowUp") {
            e.preventDefault();
            focusSiblingElement(
                (td.current?.parentNode?.children.length ?? 0) * -1,
            );
        }
    };

    return (
        <Popover.Root onOpenChange={onOpenChange} open={isOpen}>
            <Popover.Trigger asChild>
                <Cell ref={td} asChild>
                    {/* Keep the height of the cell when the popover is open, so that the popover doesn't jump when it opens/closes. */}
                    <button
                        onKeyDown={onKeyDown}
                        className="flex flex-row justify-start hover:bg-level-2 flex-wrap gap-1 items-center relative"
                        style={{ height }}
                    >
                        {children}
                        {props.blink && (
                            <div className="size-2 bg-txt-accent rounded-full top-1/2 right-3 absolute -translate-y-1/2 animate-pulse" />
                        )}
                    </button>
                </Cell>
            </Popover.Trigger>

            <Popover.Portal>
                <Popover.Content
                    side="bottom"
                    align="start"
                    sideOffset={height ? height * -1 : 0}
                    style={{
                        minHeight: height,
                    }}
                    className="border border-border-low bg-level-2 min-w-[200px] flex flex-col justify-center rounded-sm"
                >
                    {props.type === "text" && (
                        <Input {...props} {...fieldProps} />
                    )}
                    {props.type === "select" && (
                        <Select {...props} {...fieldProps} />
                    )}
                    {props.type === "multiple" && (
                        <Multiple {...props} {...fieldProps} />
                    )}
                </Popover.Content>
            </Popover.Portal>
        </Popover.Root>
    );
}

function Input(props: PropsField<"text", string>) {
    const blurOnTab: KeyboardEventHandler<HTMLInputElement> = (e) => {
        if (e.key === "Tab") {
            props.onOpenChange(false);
        }
    };
    return (
        <input
            type="text"
            defaultValue={props.value}
            onKeyDown={blurOnTab}
            className="text-sm text-txt-primary outline-none"
            onChange={(e) => props.onValueChange(e.currentTarget.value)}
            style={{ paddingLeft: props.padding }}
        />
    );
}

function Select<T>(props: PropsField<"select", T>) {
    const { __ } = useTranslate();
    const showSearch = false;
    return (
        <Command className="text-txt-primary absolute left-0 top-0 right-0 bg-level-2 border border-border-low rounded-b-sm">
            <div
                style={{ height: props.height, paddingLeft: props.padding }}
                className="flex flex-col justify-center"
                onClick={() => props.onOpenChange(false)}
            >
                {props.value ? props.itemRenderer({ item: props.value }) : ""}
            </div>
            {showSearch && (
                <Command.Input
                    className="text-sm text-txt-secondary border-y border-border-low py-2 px-3 w-full focus:outline-txt-accent outline"
                    placeholder={__("Search")}
                />
            )}
            <Command.List>
                {Array.isArray(props.items) ? (
                    props.items
                        .filter((item) => item !== props.value)
                        .map((item) => (
                            <Command.Item
                                key={getKey(item)}
                                className="py-2 px-3 hover:bg-level-3 data-[selected]:bg-level-3"
                                onSelect={() => {
                                    props.onValueChange(item);
                                    props.onOpenChange(false);
                                }}
                            >
                                {props.itemRenderer({ item })}
                            </Command.Item>
                        ))
                ) : (
                    <Suspense
                        fallback={
                            <div className="py-2 px-3">
                                <Spinner />
                            </div>
                        }
                    >
                        <SelectItems
                            value={props.value}
                            itemRenderer={props.itemRenderer}
                            items={props.items}
                            onSelect={(item) => {
                                props.onValueChange(item);
                                props.onOpenChange(false);
                            }}
                        />
                    </Suspense>
                )}
            </Command.List>
        </Command>
    );
}

function Multiple<T>(props: PropsField<"multiple", T>) {
    const { __ } = useTranslate();
    const showSearch = true;

    const pushValue = (item: T) => {
        props.onValueChange([...(props.value ?? []), item]);
    };

    const removeValue = (item: T) => {
        props.onValueChange(
            props.value!.filter((v) => getKey(v) !== getKey(item)),
        );
    };

    return (
        <Command className="text-txt-primary absolute left-0 top-0 right-0 bg-level-2 border border-border-low rounded-b-sm">
            {props.value && props.value.length > 0 && (
                <div
                    className="flex flex-col gap-2 py-3"
                    style={{
                        paddingLeft: props.padding,
                    }}
                >
                    {props.value &&
                        props.value.map((item) =>
                            props.itemRenderer({
                                item,
                                onRemove: () => removeValue(item),
                            }),
                        )}
                </div>
            )}
            {showSearch && (
                <Command.Input
                    className="text-sm text-txt-secondary border-y border-border-low py-2 px-3 w-full focus:outline-txt-accent outline"
                    placeholder={__("Search")}
                />
            )}
            <Command.List>
                {Array.isArray(props.items) ? (
                    props.items
                        .filter((item) => item !== props.value)
                        .map((item, k) => (
                            <Command.Item
                                key={k}
                                className="py-2 px-3 hover:bg-level-3 data-[selected]:bg-level-3"
                                onSelect={() => {
                                    pushValue(item);
                                }}
                            >
                                {props.itemRenderer({
                                    item,
                                })}
                            </Command.Item>
                        ))
                ) : (
                    <Suspense
                        fallback={
                            <div className="py-2 px-3">
                                <Spinner />
                            </div>
                        }
                    >
                        <SelectItems
                            value={props.value}
                            itemRenderer={props.itemRenderer}
                            items={props.items}
                            onSelect={(item) => {
                                pushValue(item);
                            }}
                        />
                    </Suspense>
                )}
            </Command.List>
        </Command>
    );
}
/**
 * Resolve items with a suspense
 */
function SelectItems<T>(props: {
    value?: T | T[];
    itemRenderer: (v: { item: T }) => ReactNode;
    items: () => T[];
    onSelect: (item: T) => void;
}) {
    const items = props.items();
    const keys = Array.isArray(props.value)
        ? props.value.map(getKey)
        : [getKey(props.value)];
    return (
        <>
            {items
                .filter((item) => !keys.includes(getKey(item)))
                .map((item) => (
                    <Command.Item
                        key={getKey(item)}
                        className="py-2 px-3 hover:bg-level-3 data-[selected]:bg-level-3"
                        onSelect={() => props.onSelect(item)}
                    >
                        {props.itemRenderer({ item })}
                    </Command.Item>
                ))}
        </>
    );
}
