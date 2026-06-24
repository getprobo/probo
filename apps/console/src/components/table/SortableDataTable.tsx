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

import { useTranslate } from "@probo/i18n";
import {
  Button,
  CellHead,
  DataTable,
  IconChevronDown,
  IconChevronTriangleDownSmall,
  Spinner,
} from "@probo/ui";
import { clsx } from "clsx";
import {
  type ComponentProps,
  createContext,
  startTransition,
  useContext,
  useState,
} from "react";
import type { LoadMoreFn } from "react-relay";
import type { OperationType } from "relay-runtime";

type Order = {
  direction: string;
  field: string;
};

export const defaultPageSize = 50;

export const SortableContext = createContext<{
  order: Order;
  onOrderChange: (order: Order) => void;
}>({
  order: {
    direction: "DESC",
    field: "CREATED_AT",
  },
  onOrderChange: () => {},
});

const defaultOrder = {
  direction: "DESC",
  field: "CREATED_AT",
};

export function SortableDataTable({
  refetch,
  hasNext,
  loadNext,
  isLoadingNext,
  pageSize = defaultPageSize,
  ...props
}: ComponentProps<typeof DataTable> & {
  refetch: (o: { order: Order }) => void;
  hasNext?: boolean;
  loadNext?: LoadMoreFn<OperationType>;
  isLoadingNext?: boolean;
  pageSize?: number;
}) {
  const { __ } = useTranslate();
  const [order, setOrder] = useState(defaultOrder);
  const onOrderChange = (o: Order) => {
    startTransition(() => {
      setOrder(o);
      refetch({ order: o });
    });
  };
  return (
    <SortableContext.Provider value={{ order, onOrderChange }}>
      <div className="space-y-4">
        <DataTable {...props} />
        {hasNext && loadNext && (
          <Button
            variant="tertiary"
            onClick={() => loadNext(pageSize)}
            className="mt-3 mx-auto"
            disabled={isLoadingNext}
            icon={isLoadingNext ? Spinner : IconChevronDown}
          >
            {__("Show more")}
          </Button>
        )}
      </div>
    </SortableContext.Provider>
  );
}

export function SortableCellHead({
  children,
  field,
  ...props
}: ComponentProps<typeof CellHead> & { field: string }) {
  const { order, onOrderChange } = useContext(SortableContext);
  const isCurrentField = order.field === field;
  const isDesc = order.direction === "DESC";
  const changeOrder = () => {
    onOrderChange({
      direction: isDesc && isCurrentField ? "ASC" : "DESC",
      field,
    });
  };
  return (
    <CellHead {...props}>
      <button
        className="flex items-center cursor-pointer hover:text-txt-primary"
        onClick={changeOrder}
        type="button"
      >
        {children}
        <IconChevronTriangleDownSmall
          size={16}
          className={clsx(
            isCurrentField && "text-txt-primary",
            isCurrentField && !isDesc && "rotate-180",
          )}
        />
      </button>
    </CellHead>
  );
}
