import type { PropsWithChildren, ReactNode } from "react";

type Props = PropsWithChildren<{
  title: ReactNode;
  description?: string | null;
}>;

export function PageHeader({ title, description, children }: Props) {
  return (
    <div className="flex justify-between items-start w-full">
      <div className=" space-y-1">
        <h1 className="text-2xl flex gap-4 font-semibold items-center">
          {title}
        </h1>
        {description && (
          <p className="text-sm text-txt-secondary">{description}</p>
        )}
      </div>
      <div className="flex gap-3">{children}</div>
    </div>
  );
}
