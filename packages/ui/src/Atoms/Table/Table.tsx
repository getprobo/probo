import { createContext, useContext, type PropsWithChildren } from "react";
import { Card } from "../Card/Card";
import { Link } from "react-router";
import clsx from "clsx";

export function Table({ children }: PropsWithChildren) {
    return (
        <Card className="relative w-full overflow-auto">
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
}: PropsWithChildren<{ className?: string }>) {
    return (
        <th className={clsx("first:pl-6 last:pr-6 py-3", className)}>
            {children}
        </th>
    );
}

const TrContext = createContext({} as { to?: string });

export function Tr({ children, to }: PropsWithChildren<{ to?: string }>) {
    return (
        <TrContext value={{ to }}>
            <tr
                className={clsx(
                    "border-border-low border-y first:border-none last:border-none",
                    to && "hover:bg-subtle",
                )}
            >
                {children}
            </tr>
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
}: PropsWithChildren<{ noLink?: boolean; className?: string }>) {
    const { to } = useContext(TrContext);
    if (!to || noLink) {
        return (
            <td className={clsx("first:pl-6 last:pr-6 py-3", className)}>
                {children}
            </td>
        );
    }
    return (
        <td
            className={clsx(
                "first:*:pl-6 *:block last:*:pr-6 *:py-3",
                className,
            )}
        >
            <Link to={to}>{children}</Link>
        </td>
    );
}
