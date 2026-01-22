import { useTranslate } from "@probo/i18n";
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
import type { LoadMoreFn } from "react-relay";
import type { OperationType } from "relay-runtime";

export type Order = {
  direction: "ASC" | "DESC";
  field: string;
};

const defaultPageSize = 50;

export const SortableContext = createContext({
  order: {
    direction: "DESC",
    field: "CREATED_AT",
  },
  changeOrder: (() => {}) as (order: Order) => void,
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
  const { __ } = useTranslate();
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
            {__("Show more")}
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
