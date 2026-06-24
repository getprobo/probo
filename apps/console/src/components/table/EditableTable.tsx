// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { useToggle } from "@probo/hooks";
import {
  Button,
  Cell,
  CellHead,
  EditableRow,
  IconCheckmark1,
  Row,
  RowButton,
  Spinner,
} from "@probo/ui";
import { clsx } from "clsx";
import { type ReactNode } from "react";
import type { KeyType, KeyTypeData } from "react-relay/ReactRelayTypes";
import type { usePaginationFragmentHookType } from "react-relay/relay-hooks/usePaginationFragment";
import type { GraphQLTaggedNode, OperationType } from "relay-runtime";
import { z } from "zod";

import { defaultPageSize,
  SortableCellHead,
  SortableDataTable,
} from "#/components/table/SortableDataTable";
import { useMutateField } from "#/hooks/useMutateField";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useStateWithSchema } from "#/hooks/useStateWithSchema";

type ColumnDefinition = { label: string; field: string } | string;

type EditableTableRowProps<T, S extends z.ZodType<Record<string, unknown>>> = {
  item?: T;
  onUpdate: (key: keyof z.infer<S>, value: z.infer<S>[typeof key]) => void;
  errors: Record<string, string>;
};

/**
 * A "all-in-one" component to create a table with editable cells.
 */
export function EditableTable<
  T extends { id: string },
  S extends z.ZodType<Record<string, unknown>>,
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
  pageSize?: number;
}) {
  const { update } = useMutateField(props.updateMutation);
  const [showAdd, toggleAdd] = useToggle(false);

  return (
    <SortableDataTable
      columns={[...props.columns.map(() => "minmax(min-content, 1fr)"), "56px"]}
      refetch={props.pagination.refetch}
      hasNext={props.pagination.hasNext}
      isLoadingNext={props.pagination.isLoadingNext}
      loadNext={props.pagination.loadNext}
      pageSize={props.pageSize ?? defaultPageSize}
    >
      <Row>
        {props.columns.map((column, index) => (
          <EditableTableHead column={column} key={index} />
        ))}
        <CellHead />
      </Row>
      {props.items.map(item => (
        <EditableRow onUpdate={(k, v) => update(item.id, k, v)} key={item.id}>
          {props.row({
            item,
            onUpdate: (key, value) => update(item.id, key as string, value),
            errors: {},
          })}
          <Cell>{props.action({ item })}</Cell>
        </EditableRow>
      ))}
      {showAdd
        ? (
            <NewItemRow
              schema={props.schema}
              defaultValue={props.defaultValue}
              connectionId={props.connectionId}
              row={props.row}
              mutation={props.createMutation}
              onSuccess={toggleAdd}
            />
          )
        : (
            <RowButton onClick={toggleAdd} type="button">{props.addLabel}</RowButton>
          )}
    </SortableDataTable>
  );
}

function NewItemRow<
  T extends { id: string },
  S extends z.ZodType<Record<string, unknown>>,
>(props: {
  schema: S;
  defaultValue: z.infer<S>;
  connectionId: string;
  mutation: GraphQLTaggedNode;
  onSuccess: () => void;
  row: (props: EditableTableRowProps<T, S>) => ReactNode;
}) {
  const { update, errors, value } = useStateWithSchema(
    props.schema,
    props.defaultValue,
  );
  const [mutate, isMutating] = useMutationWithToasts(props.mutation);
  const isOk = Object.keys(errors ?? {}).length === 0;

  const onSubmit = async () => {
    // This should never happen, but we don't want to send bad data
    if (!isOk) {
      alert("Please fix the errors before submitting.");
      return;
    }
    await mutate({
      variables: {
        input: value,
        connections: [props.connectionId],
      },
      onSuccess: props.onSuccess,
    });
  };
  return (
    <EditableRow
      onUpdate={update as (key: string, value: unknown) => void}
      errors={errors}
    >
      {props.row({ errors, onUpdate: update })}
      <Cell>
        <Button
          disabled={!isOk || isMutating}
          variant="tertiary"
          className={clsx(isOk ? "text-txt-success" : "text-txt-secondary")}
          onClick={() => void onSubmit()}
        >
          {isMutating ? <Spinner size={16} /> : <IconCheckmark1 size={16} />}
        </Button>
      </Cell>
    </EditableRow>
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
