import type { PropsWithChildren } from "react";

import { IconCircleInfo } from "../Atoms/Icons";

type Props = PropsWithChildren<{
  title?: string;
  description?: string;
}>;

export function ErrorLayout({ title, description, children }: Props) {
  return (
    <div className="min-h-screen w-full gap-2 flex flex-col items-center justify-center">
      <IconCircleInfo className="text-txt-danger" size={40} />
      <h1 className="text-4xl font-bold">{title}</h1>
      <p className="text-txt-secondary">{description}</p>
      {children}
    </div>
  );
}
