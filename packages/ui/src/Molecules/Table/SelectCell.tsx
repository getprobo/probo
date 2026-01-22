import { useStateWithRef } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Command } from "cmdk";
import { Fragment, type ReactNode } from "react";
import { tv } from "tailwind-variants";

import { Badge } from "../../Atoms/Badge/Badge";

import { EditableCell, useEditableCellRef } from "./EditableCell";
import { useEditableRowContext } from "./EditableRow";
import { getKey } from "./utils";

type Props<T> = {
  name: string;
  items: T[];
  itemRenderer: (v: { item: T; onRemove?: (item: T) => void }) => ReactNode;
} & (
  | { defaultValue: T; multiple?: undefined }
  | { defaultValue: T[]; multiple: true }
);

export const selectCell = tv({
  slots: {
    command:
            "text-txt-primary absolute left-0 top-0 right-0 bg-level-2 border border-border-low rounded-b-sm",
    value: "flex flex-col gap-2 py-3 justify-center",
    input: "text-sm text-txt-secondary border-y border-border-low py-2 px-3 w-full focus:outline-txt-accent outline",
    item: "py-2 px-3 hover:bg-level-3 data-[selected=true]:bg-level-3",
  },
});

export function SelectCell<T extends NonNullable<unknown>>(props: Props<T>) {
  const [value, setValue, valueRef] = useStateWithRef<T | T[]>(
    props.defaultValue,
  );
  const cellRef = useEditableCellRef();
  const { __ } = useTranslate();
  const filteredValue = Array.isArray(value) ? value.filter(v => v !== undefined) : value ? [value] : [];
  const usedKeys = new Set<string>(filteredValue.map(getKey).filter(Boolean) as string[]);
  const { onUpdate } = useEditableRowContext();

  const onSelect = (item: T) => {
    if (props.multiple) {
      setValue([...((valueRef.current as T[]) ?? []), item]);
      return;
    }
    setValue(item);
    cellRef.current?.close();
  };

  const onClose = () => {
    if (valueRef.current === props.defaultValue) {
      return;
    }
    onUpdate(props.name, valueRef.current);
  };

  const classNames = selectCell();

  return (
    <EditableCell
      name={props.name}
      label={
        <SelectValue value={value} itemRenderer={props.itemRenderer} />
      }
      onClose={onClose}
      ref={cellRef}
    >
      <Command className={classNames.command()}>
        <div
          className={classNames.value()}
          style={{
            paddingLeft: "var(--padding)",
            minHeight: "var(--height)",
          }}
        >
          {" "}
          <SelectValue
            onValueChange={setValue}
            value={value}
            itemRenderer={props.itemRenderer}
          />
        </div>
        {" "}
        {props.multiple && (
          <Command.Input
            className={classNames.input()}
            placeholder={__("Search")}
          />
        )}
        <Command.List>
          {props.items
            .filter(item => !usedKeys.has(getKey(item) ?? ""))
            .map(item => (
              <Command.Item
                key={getKey(item)}
                className={classNames.item()}
                onSelect={() => onSelect(item)}
              >
                {props.itemRenderer({ item })}
              </Command.Item>
            ))}
        </Command.List>
      </Command>
    </EditableCell>
  );
}

export function SelectValue<T extends NonNullable<unknown>>(props: {
  itemRenderer: Props<T>["itemRenderer"];
  onValueChange?: (value: T | T[]) => void;
  value: T | T[] | undefined;
}) {
  if (props.value === undefined) {
    return "";
  }
  if (!Array.isArray(props.value)) {
    return props.value ? props.itemRenderer({ item: props.value }) : "";
  }

  const removeValue = (item: T) => {
    if (!Array.isArray(props.value) || !props.onValueChange) {
      return;
    }
    props.onValueChange(
      props.value.filter(v => getKey(v) !== getKey(item)),
    );
  };

  if (!props.onValueChange && props.value.length > 0) {
    return (
      <>
        {props.value.slice(0, 3).map(item => (
          <Fragment key={getKey(item)}>
            {props.itemRenderer({
              item,
              onRemove: props.onValueChange
                ? () => removeValue(item)
                : undefined,
            })}
          </Fragment>
        ))}
        {props.value.length > 3 && (
          <Badge className="text-txt-secondary">
            +
            {props.value.length - 3}
          </Badge>
        )}
      </>
    );
  }

  return (
    <>
      {props.value.map(item => (
        <Fragment key={getKey(item)}>
          {props.itemRenderer({
            item,
            onRemove: props.onValueChange
              ? () => removeValue(item)
              : undefined,
          })}
        </Fragment>
      ))}
    </>
  );
}
