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

export const SortableContext = createContext({
  order: {
    direction: "DESC",
    field: "CREATED_AT",
  },
  onOrderChange: (() => {}) as (order: Order) => void,
});

const defaultOrder = {
  direction: "DESC",
  field: "CREATED_AT",
} as Order;

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
