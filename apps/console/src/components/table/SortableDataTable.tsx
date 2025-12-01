import { useTranslate } from "@probo/i18n";
import {
  Button,
  CellHead,
  DataTable,
  IconChevronDown,
  IconChevronTriangleDownSmall,
  Spinner,
} from "@probo/ui";
import clsx from "clsx";
import {
  type ComponentProps,
  createContext,
  startTransition,
  useContext,
  useState,
} from "react";

type Order = {
  direction: string;
  field: string;
};

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
  ...props
}: ComponentProps<typeof DataTable> & {
  refetch: (o: { order: Order }) => void;
  hasNext?: boolean;
  loadNext?: (...args: any[]) => void;
  isLoadingNext?: boolean;
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
            onClick={() => loadNext()}
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
