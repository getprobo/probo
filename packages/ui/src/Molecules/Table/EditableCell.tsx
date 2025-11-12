import {
    type FocusEventHandler,
    type KeyboardEventHandler,
    type ReactNode,
    useRef,
    useState,
} from "react";
import * as Popover from "@radix-ui/react-popover";
import { Select } from "../../Atoms/Select/Select.tsx";
import { Spinner } from "../../Atoms/Spinner/Spinner.tsx";
import { Cell } from "../../Atoms/DataTable/DataTable.tsx";

type Props = {
    type: "text" | "select";
    onValueChange: (value: string) => void;
    defaultValue?: ReactNode;
    children?: ReactNode;
    options?: ReactNode;
    isLoading?: boolean;
};

export function EditableCell({
    options,
    children,
    isLoading,
    onValueChange,
    defaultValue,
    type,
}: Props) {
    const td = useRef<HTMLTableCellElement>(null);
    const [height, setHeight] = useState(0);
    const [padding, setPadding] = useState("12px");

    const onOpenChange = (open: boolean) => {
        if (open) {
            setOpen(open);
            setHeight(td.current?.offsetHeight ?? 0);
            setPadding(getComputedStyle(td.current!).paddingLeft);
        }
    };

    const onInputBlur: FocusEventHandler<HTMLInputElement> = (e) => {
        setOpen(false);
        onValueChange(e.target.value);
    };

    const blurOnTab: KeyboardEventHandler<HTMLInputElement> = (e) => {
        if (e.key === "Tab") {
            setOpen(false);
            onValueChange(e.currentTarget.value);
        }
    };

    const [isOpen, setOpen] = useState(false);

    return (
        <Popover.Root onOpenChange={onOpenChange} open={isOpen}>
            <Popover.Trigger asChild>
                <Cell ref={td} asChild>
                    <button className="flex flex-row space-between gap-2 w-full items-center justify-start">
                        {" "}
                        {children ?? defaultValue}
                        {isLoading && <Spinner size={12} />}
                    </button>
                </Cell>
            </Popover.Trigger>
            <Popover.Portal>
                <Popover.Content
                    side="bottom"
                    align="start"
                    sideOffset={height * -1}
                    style={{
                        height,
                        paddingLeft: type === "select" ? 0 : padding,
                    }}
                    className="border border-border-low bg-level-2 min-w-[200px] min-h-[57px] flex flex-col justify-center rounded-sm"
                >
                    {type === "select" && (
                        <>
                            <Select
                                onValueChange={onValueChange}
                                onOpenChange={setOpen}
                                defaultOpen
                                variant="ghost"
                                style={{
                                    height: height,
                                }}
                                placeholder={children}
                                dropdownProps={{
                                    sideOffset: 0,
                                    className:
                                        "rounded-t-none border border-border-low",
                                }}
                            >
                                {options}
                            </Select>
                        </>
                    )}
                    {type === "text" && (
                        <input
                            type="text"
                            defaultValue={defaultValue?.toString() ?? ""}
                            className="text-txt-primary text-sm outline-none"
                            onKeyDown={blurOnTab}
                            onBlur={onInputBlur}
                        />
                    )}
                </Popover.Content>
            </Popover.Portal>
        </Popover.Root>
    );
}
