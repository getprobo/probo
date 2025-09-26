import type { PropsWithChildren } from "react";
import clsx from "clsx";

export function Rows({
  children,
  className,
}: PropsWithChildren<{ className?: string }>) {
  return (
    <div
      className={clsx(
        "*:first:rounded-t-lg *:last:rounded-b-lg *:px-6 *:py-3 *:-mt-[1px] *:border *:border-border-solid",
        className,
      )}
    >
      {children}
    </div>
  );
}
