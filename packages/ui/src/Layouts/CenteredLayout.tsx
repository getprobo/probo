import type { PropsWithChildren } from "react";
import { Outlet } from "react-router";

import { Skeleton } from "../Atoms/Skeleton/Skeleton";

export function CenteredLayout({ children }: PropsWithChildren) {
  return (
    <div className="grid place-items-center min-h-[calc(100vh-48px-(2*48px))] text-txt-primary bg-level-0">
      <div className="w-full max-w-2xl flex flex-col items-center">
        {children ?? <Outlet />}
      </div>
    </div>
  );
}

export function CenteredLayoutSkeleton() {
  return (
    <CenteredLayout>
      <div className="w-full max-w-2xl flex flex-col items-center space-y-6">
        <Skeleton className="w-77 h-9" />
        <Skeleton className="w-full h-20" />
      </div>
    </CenteredLayout>
  );
}
