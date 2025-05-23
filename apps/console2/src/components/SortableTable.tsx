import { IconChevronTriangleDownSmall, Table, Th } from "@probo/ui";
import clsx from "clsx";
import {
  createContext,
  startTransition,
  useContext,
  useState,
  type ComponentProps,
} from "react";

type Order = {
  direction: string;
  field: string;
};

const SortableContext = createContext({
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

export function SortableTable({
  refetch,
  ...props
}: ComponentProps<typeof Table> & {
  refetch: (o: { order: Order }) => void;
}) {
  const [order, setOrder] = useState(defaultOrder);
  const onOrderChange = (o: Order) => {
    startTransition(() => {
      setOrder(o);
      refetch({ order: o });
    });
  };
  return (
    <SortableContext value={{ order, onOrderChange }}>
      <Table {...props} />
    </SortableContext>
  );
}

export function SortableTh({
  children,
  field,
  ...props
}: ComponentProps<typeof Th> & { field: string }) {
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
    <Th {...props}>
      <button
        className="flex items-center cursor-pointer hover:text-txt-primary"
        onClick={changeOrder}
      >
        {children}
        <IconChevronTriangleDownSmall
          size={16}
          className={clsx(
            isCurrentField && "text-txt-primary",
            isCurrentField && !isDesc && "rotate-180"
          )}
        />
      </button>
    </Th>
  );
}
