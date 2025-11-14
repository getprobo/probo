import {
  SortableCellHead,
  SortableDataTable,
} from "/components/table/SortableDataTable.tsx";
import type { GraphQLTaggedNode, OperationType } from "relay-runtime";
import type { KeyType, KeyTypeData } from "react-relay/relay-hooks/helpers";
import type { usePaginationFragmentHookType } from "react-relay/relay-hooks/usePaginationFragment";
import {
  Button,
  Cell,
  CellHead,
  IconCheckmark1,
  Row,
  RowButton,
  Spinner,
} from "@probo/ui";
import { z } from "zod";
import { useMutateField } from "/hooks/useMutateField.tsx";
import { type ReactNode } from "react";
import { useToggle } from "@probo/hooks";
import { useStateWithSchema } from "/hooks/useStateWithSchema.ts";
import { useMutation } from "react-relay";
import clsx from "clsx";

type ColumnDefinition = { label: string; field: string } | string;

type EditableTableRowProps<T, S extends z.ZodSchema> = {
  item?: T;
  onUpdate: (key: keyof z.infer<S>, value: z.infer<S>[typeof key]) => void;
  errors: Record<string, string>;
};

/**
 * A "all-in-one" component to create a table with editable cells.
 */
export function EditableTable<
  T extends { id: string },
  S extends z.ZodSchema,
>(props: {
  // Schema to create a new item
  schema: S;
  // GraphQL related props
  connectionId: string;
  pagination: usePaginationFragmentHookType<
    OperationType,
    KeyType,
    KeyTypeData<KeyType>
  >;
  updateMutation: GraphQLTaggedNode;
  createMutation: GraphQLTaggedNode;
  items: T[];
  // List of the columns
  columns: ColumnDefinition[];
  // Render a row for each item and to create a new item
  row: (props: EditableTableRowProps<T, S>) => ReactNode;
  // Render the content of the last cell
  action: (props: { item: T }) => ReactNode;
  // Label used when adding a new item
  addLabel: string;
  // Default value used when creating a new item
  defaultValue: z.infer<S>;
}) {
  const { update } = useMutateField(props.updateMutation);
  const [showAdd, toggleAdd] = useToggle(false);

  return (
    <SortableDataTable
      columns={[...props.columns.map(() => "1fr"), "56px"]}
      refetch={props.pagination.refetch}
      hasNext={props.pagination.hasNext}
      isLoadingNext={props.pagination.isLoadingNext}
      loadNext={props.pagination.loadNext}
    >
      <Row>
        {props.columns.map((column, index) => (
          <EditableTableHead column={column} key={index} />
        ))}
        <CellHead />
      </Row>
      {props.items.map((item) => (
        <Row key={item.id}>
          {props.row({
            item,
            onUpdate: (key, value) => update(item.id, key as string, value),
            errors: {},
          })}
          <Cell>{props.action({ item })}</Cell>
        </Row>
      ))}
      {showAdd ? (
        <NewItemRow
          schema={props.schema}
          defaultValue={props.defaultValue}
          connectionId={props.connectionId}
          row={props.row}
          mutation={props.createMutation}
        />
      ) : (
        <RowButton onClick={toggleAdd}>{props.addLabel}</RowButton>
      )}
    </SortableDataTable>
  );
}

function NewItemRow<S extends z.ZodSchema>(props: {
  schema: S;
  defaultValue: z.infer<S>;
  connectionId: string;
  mutation: GraphQLTaggedNode;
  row: (props: EditableTableRowProps<any, S>) => ReactNode;
}) {
  const { update, errors, value } = useStateWithSchema(
    props.schema,
    props.defaultValue,
  );
  const [mutate, isMutating] = useMutation(props.mutation);
  const isOk = Object.keys(errors ?? {}).length === 0;

  const onSubmit = async () => {
    // This should never happen, but we don't want to send bad data
    if (!isOk) {
      alert("Please fix the errors before submitting.");
      return;
    }
    mutate({
      variables: {
        input: value,
        connections: [props.connectionId],
      },
    });
  };
  return (
    <Row>
      {props.row({ errors, onUpdate: update })}
      <Cell>
        <Button
          disabled={!isOk || isMutating}
          variant="tertiary"
          className={clsx(isOk ? "text-txt-success" : "text-txt-secondary")}
          onClick={onSubmit}
        >
          {isMutating ? <Spinner size={16} /> : <IconCheckmark1 size={16} />}
        </Button>
      </Cell>
    </Row>
  );
}

function EditableTableHead(props: { column: ColumnDefinition }) {
  if (typeof props.column === "string") {
    return <CellHead>{props.column}</CellHead>;
  }
  return (
    <SortableCellHead field={props.column.field}>
      {props.column.label}
    </SortableCellHead>
  );
}
