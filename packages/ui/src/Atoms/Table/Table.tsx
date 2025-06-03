import {
    createContext,
    useContext,
    type HTMLAttributes,
    type PropsWithChildren,
} from "react";
import { Card } from "../Card/Card";
import { Link } from "react-router";
import clsx from "clsx";

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

export function Thead({ children }: PropsWithChildren) {
    return (
        <thead className="text-xs text-txt-tertiary font-semibold border-border-low border-b">
            {children}
        </thead>
    );
}

export function Th({
    children,
    className,
    width,
}: PropsWithChildren<{ className?: string; width?: number }>) {
    return (
        <th
            className={clsx("first:pl-6 last:pr-6 py-3", className)}
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
            className={clsx(
                "first:*:pl-6 *:block last:*:pr-6 *:py-3",
                className,
            )}
        >
            <Link to={to}>{children}</Link>
        </td>
    );
}
