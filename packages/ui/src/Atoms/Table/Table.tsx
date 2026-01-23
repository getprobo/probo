import { clsx } from "clsx";
import {
  createContext,
  type FC,
  type HTMLAttributes,
  type PropsWithChildren,
  type ReactNode,
  type ThHTMLAttributes,
  useContext,
} from "react";
import { Link } from "react-router";

import { Card } from "../Card/Card";
import { IconPlusLarge } from "../Icons";

export function Table({
  children,
  className,
}: PropsWithChildren<{ className?: string }>) {
  return (
    <Card className={clsx("relative w-full overflow-auto", className)}>
      <table className="w-full text-left">{children}</table>
    </Card>
  );
}

export function Thead({
  children,
  className,
}: PropsWithChildren<{ className?: string }>) {
  return (
    <thead
      className={clsx(
        "text-xs text-txt-tertiary font-semibold border-border-low border-b",
        className,
      )}
    >
      {children}
    </thead>
  );
}

export function Th({
  children,
  className,
  width,
  compact,
  ...props
}: {
  className?: string;
  width?: number;
  colspan?: number;
  compact?: boolean;
} & ThHTMLAttributes<HTMLTableCellElement>) {
  return (
    <th
      {...props}
      className={clsx(
        "first:pl-6 last:pr-6 whitespace-nowrap",
        compact ? "py-1" : "py-3",
        className,
      )}
      style={{ width }}
    >
      {children}
    </th>
  );
}

const TrContext = createContext({} as { to?: string });

export function Tr({
  to,
  className,
  ...props
}: { to?: string } & HTMLAttributes<HTMLTableRowElement>) {
  return (
    <TrContext value={{ to }}>
      <tr
        {...props}
        className={clsx(
          "border-border-low border-y first:border-none last:border-none",
          (to || props.onClick) && "hover:bg-subtle",
          className,
        )}
      />
    </TrContext>
  );
}

export function Tbody({ children }: PropsWithChildren) {
  return (
    <tbody className="text-sm text-txt-primary bg-tertiary">
      {children}
    </tbody>
  );
}

export function Td({
  children,
  noLink,
  className,
  width,
  ...props
}: {
  noLink?: boolean;
  width?: number;
  colSpan?: number;
} & HTMLAttributes<HTMLTableCellElement>) {
  const { to } = useContext(TrContext);
  if (!to || noLink) {
    return (
      <td
        {...props}
        width={width}
        className={clsx("first:pl-6 last:pr-6 py-3", className)}
      >
        {children}
      </td>
    );
  }
  return (
    <td
      {...props}
      width={width}
      className={clsx("first:*:pl-6 *:pr-6 *:block *:py-3", className)}
    >
      <Link to={to} className="select-text" draggable={false}>
        {children}
      </Link>
    </td>
  );
}

export function TrButton({
  icon = IconPlusLarge,
  children,
  colspan,
  ...props
}: {
  colspan?: number;
  children: ReactNode;
  icon?: FC<{ size: number; className?: string }>;
} & HTMLAttributes<HTMLButtonElement>) {
  const IconComponent = icon;
  return (
    <tr>
      <td colSpan={colspan}>
        <button
          {...props}
          className="py-2 bg-highlight hover:bg-highlight-hover active:bg-highlight-pressed cursor-pointer w-full flex gap-2 items-center justify-center"
        >
          <IconComponent size={16} />
          {children}
        </button>
      </td>
    </tr>
  );
}
