import { clsx } from "clsx";

type Props = { size?: number; className?: string; centered?: boolean };

export function Spinner({ size = 16, className, centered }: Props) {
  return (
    <div
      className={clsx(
        "animate-spin rounded-full border-b-2",
        centered && "my-4 mx-auto",
        className,
      )}
      style={{ width: size, height: size, borderColor: "currentColor" }}
    />
  );
}
