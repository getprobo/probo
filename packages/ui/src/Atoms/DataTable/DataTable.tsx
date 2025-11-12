import { type ComponentPropsWithRef, type PropsWithChildren } from "react";
import { Card } from "../Card/Card";
import clsx from "clsx";
import { type AsChildProps, Slot } from "../Slot.tsx";

export function DataTable({
    children,
    className,
    columns,
}: PropsWithChildren<{ className?: string; columns: number | string[] }>) {
    const style = () => {
        if (typeof columns === "number") {
            return {
                gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`,
            };
        }
        return {
            gridTemplateColumns: columns
                .map((col) => `minmax(0, ${col})`)
                .join(" "),
        };
    };

    return (
        <div className="overflow-auto relative w-full p-1 -m-1">
            <Card
                className={clsx(className, "w-full text-left grid")}
                style={style()}
            >
                {children}
            </Card>
        </div>
    );
}

export function CellHead({
    children,
    className,
    ...props
}: ComponentPropsWithRef<"div">) {
    return (
        <div
            {...props}
            className={clsx(
                "text-xs text-txt-tertiary font-semibold border-border-low border-b",
                "px-6 whitespace-nowrap py-3",
                className,
            )}
        >
            {children}
        </div>
    );
}

export function Cell({
    asChild,
    ...props
}: AsChildProps<ComponentPropsWithRef<"div">>) {
    const Component = asChild ? Slot : "div";
    return (
        <Component
            data-cell
            {...props}
            className={clsx(
                "py-3 px-6 text-sm text-txt-primary bg-tertiary border-t border-border-low flex flex-col justify-center",
                props.className,
            )}
        />
    );
}
