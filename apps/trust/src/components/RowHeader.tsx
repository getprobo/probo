import type { PropsWithChildren } from "react";

export function RowHeader({ children }: PropsWithChildren) {
  return (
    <div className="bg-subtle text-xs font-medium text-txt-tertiary uppercase ">
      {children}
    </div>
  );
}
