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

import {
  Button,
  IconChevronDown,
  IconChevronTriangleDownSmall,
  Spinner,
  Table,
  Th,
} from "@probo/ui";
import { clsx } from "clsx";
import {
  type ComponentProps,
  createContext,
  startTransition,
  useContext,
  useState,
} from "react";
import { useTranslation } from "react-i18next";
import type { LoadMoreFn } from "react-relay";
import type { OperationType } from "relay-runtime";

export type Order = {
  direction: "ASC" | "DESC";
  field: string;
};

const defaultPageSize = 50;

export const SortableContext = createContext<{
  order: Order;
  changeOrder: (order: Order) => void;
}>({
  order: {
    direction: "DESC",
    field: "CREATED_AT",
  },
  changeOrder: () => {},
});

const defaultOrder = {
  direction: "DESC",
  field: "CREATED_AT",
} as Order;

export function SortableTable({
  refetch,
  hasNext,
  loadNext,
  isLoadingNext,
  pageSize = defaultPageSize,
  ...props
}: ComponentProps<typeof Table> & {
  refetch: (o: { order: Order }) => void;
  hasNext?: boolean;
  loadNext?: LoadMoreFn<OperationType>;
  isLoadingNext?: boolean;
  pageSize?: number;
}) {
  const { t } = useTranslation();
  const [order, setOrder] = useState(defaultOrder);
  const changeOrder = (o: Order) => {
    startTransition(() => {
      setOrder(o);
      refetch({ order: o });
    });
  };
  return (
    <SortableContext value={{ order, changeOrder }}>
      <div className="space-y-4">
        <Table {...props} />
        {hasNext && loadNext && (
          <Button
            variant="tertiary"
            onClick={() => loadNext(pageSize)}
            className="mt-3 mx-auto"
            disabled={isLoadingNext}
            icon={isLoadingNext ? Spinner : IconChevronDown}
          >
            {t("sortableTable.actions.showMore")}
          </Button>
        )}
      </div>
    </SortableContext>
  );
}

export function SortableTh({
  children,
  field,
  onOrderChange,
  ...props
}: ComponentProps<typeof Th> & {
  field: string;
  onOrderChange?: (order: { direction: "ASC" | "DESC"; field: string }) => void;
}) {
  const { order, changeOrder } = useContext(SortableContext);
  const isCurrentField = order.field === field;
  const isDesc = order.direction === "DESC";
  const handleChangeOrder = () => {
    const newOrder = {
      direction:
        isDesc && isCurrentField ? ("ASC" as const) : ("DESC" as const),
      field,
    };
    changeOrder(newOrder);
    onOrderChange?.(newOrder);
  };
  return (
    <Th {...props}>
      <button
        className="flex items-center cursor-pointer hover:text-txt-primary"
        onClick={handleChangeOrder}
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
    </Th>
  );
}
