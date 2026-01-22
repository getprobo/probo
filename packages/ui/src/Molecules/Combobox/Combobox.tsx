import {
  Combobox as AriaKitCombobox,
  ComboboxItem as AriaKitComboboxItem,
  ComboboxPopover,
  ComboboxProvider,
} from "@ariakit/react";
import { isEmpty } from "@probo/helpers";
import {
  type ComponentProps,
  type PropsWithChildren,
  type ReactNode,
} from "react";

import { dropdown, dropdownItem } from "../../Atoms/Dropdown/Dropdown";
import { input } from "../../Atoms/Input/Input";

type Props = {
  children: ReactNode;
  loading?: boolean;
  onSearch: (query: string) => void;
  placeholder?: string;
  autoSelect?: boolean;
  onSelect?: (value: string) => void;
  resetValueOnHide?: boolean;
  value?: string;
} & Omit<ComponentProps<typeof AriaKitCombobox>, "onSelect" | "value">;

export function Combobox({
  children,
  onSearch,
  placeholder,
  autoSelect,
  onSelect,
  resetValueOnHide,
  value,
  ...props
}: Props) {
  const showDropdown = !isEmpty(children);
  return (
    <ComboboxProvider
      value={value}
      setValue={onSearch}
      setSelectedValue={v => onSelect?.(v as string)}
      resetValueOnHide={resetValueOnHide}
    >
      <AriaKitCombobox
        {...props}
        autoSelect={autoSelect}
        placeholder={placeholder}
        className={input()}
      />
      {showDropdown && (
        <ComboboxPopover
          gutter={4}
          sameWidth
          className={dropdown()}
          style={{ maxHeight: "var(--popover-available-height)" }}
        >
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
    <AriaKitComboboxItem hideOnClick className={dropdownItem()} {...props}>
      {children}
    </AriaKitComboboxItem>
  );
}
