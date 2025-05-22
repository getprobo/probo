import {
    Combobox as AriaKitCombobox,
    ComboboxItem as AriaKitComboboxItem,
    ComboboxPopover,
    ComboboxProvider,
} from "@ariakit/react";
import {
    type ComponentProps,
    type PropsWithChildren,
    type ReactNode,
} from "react";
import { input } from "../../Atoms/Input/Input";
import { dropdown, dropdownItem } from "../../Atoms/Dropdown/Dropdown";
import { isEmpty } from "@probo/helpers";

type Props = {
    children: ReactNode;
    loading?: boolean;
    onSearch: (query: string) => void;
    placeholder?: string;
};

export function Combobox({ children, onSearch, placeholder }: Props) {
    const showDropdown = !isEmpty(children);
    return (
        <ComboboxProvider setValue={onSearch}>
            <AriaKitCombobox placeholder={placeholder} className={input()} />
            {showDropdown && (
                <ComboboxPopover gutter={4} sameWidth className={dropdown()}>
                    {children}
                </ComboboxPopover>
            )}
        </ComboboxProvider>
    );
}

export function ComboboxItem({
    children,
    ...props
}: PropsWithChildren<ComponentProps<typeof AriaKitComboboxItem>>) {
    return (
        <AriaKitComboboxItem className={dropdownItem()} {...props}>
            {children}
        </AriaKitComboboxItem>
    );
}
