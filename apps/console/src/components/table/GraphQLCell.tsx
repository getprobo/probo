import { useStateWithRef } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { EditableCell, selectCell, SelectValue, Spinner } from "@probo/ui";
import { useEditableCellRef } from "@probo/ui/src/Molecules/Table/EditableCell";
import { useEditableRowContext } from "@probo/ui/src/Molecules/Table/EditableRow";
import { getKey } from "@probo/ui/src/Molecules/Table/utils";
import { Command } from "cmdk";
import { type ReactNode, Suspense } from "react";
import { useLazyLoadQuery } from "react-relay";
import type {
  GraphQLTaggedNode,
  OperationType,
  VariablesOf,
} from "relay-runtime";

type Props<Q extends OperationType, T> = {
  name: string;
  query: GraphQLTaggedNode;
  variables: VariablesOf<Q>;
  items: (v: ReturnType<typeof useLazyLoadQuery<Q>>) => T[];
  itemRenderer: (v: { item: T; onRemove?: (item: T) => void }) => ReactNode;
} & (
  | { defaultValue?: T; multiple?: undefined }
  | { defaultValue: T[]; multiple: true }
);

export function GraphQLCell<Q extends OperationType, T extends NonNullable<unknown>>(props: Props<Q, T>) {
  const [value, setValue, valueRef] = useStateWithRef<T | T[] | undefined>(
    props.defaultValue,
  );
  const cellRef = useEditableCellRef();
  const { __ } = useTranslate();
  const filteredValue = Array.isArray(value) ? value.filter(Boolean) : value ? [value] : [];
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
    // Only send ids when updating the value
    onUpdate(
      props.name,
      Array.isArray(valueRef.current)
        ? valueRef.current.map(getKey)
        : getKey(valueRef.current),
    );
  };

  const classNames = selectCell();

  return (
    <EditableCell
      name={props.name}
      label={<SelectValue value={value} itemRenderer={props.itemRenderer} />}
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
          <Suspense
            fallback={(
              <div className="py-2 px-3 flex items-center justify-center">
                <Spinner />
              </div>
            )}
          >
            <ItemList
              {...props}
              usedKeys={usedKeys}
              className={classNames.item()}
              onSelect={onSelect}
            />
          </Suspense>
        </Command.List>
      </Command>
    </EditableCell>
  );
}

function ItemList<Q extends OperationType, T>(
  props: Props<Q, T> & {
    className: string;
    onSelect: (item: T) => void;
    usedKeys: Set<string>;
  },
) {
  const data = useLazyLoadQuery<Q>(props.query, props.variables, {
    fetchPolicy: "network-only",
  });
  const items = props.items(data);
  return (
    <>
      {items
        .filter(item => !props.usedKeys.has(getKey(item) ?? ""))
        .map(item => (
          <Command.Item
            key={getKey(item)}
            className={props.className}
            onSelect={() => props.onSelect(item)}
          >
            {props.itemRenderer({ item })}
          </Command.Item>
        ))}
    </>
  );
}
