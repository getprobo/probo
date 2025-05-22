import clsx from "clsx";

type Props = { size: number; className?: string };

export function Spinner({ size = 16, className }: Props) {
    return (
        <div
            className={clsx("animate-spin rounded-full border-b-2", className)}
            style={{ width: size, height: size, borderColor: "currentColor" }}
        />
    );
}
