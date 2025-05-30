import { Skeleton } from "@/components/ui/skeleton";
import { PageTemplate } from "@/components/PageTemplate";
import { useLocation } from "react-router";
import { Suspense } from "react";
import { ErrorBoundaryWithLocation } from "../ErrorBoundary";
import { lazy } from "@probo/react-lazy";

const DataListView = lazy(() => import("./DataListView"));

export function DataListViewSkeleton() {
  return (
    <PageTemplate
      title="Data"
      description="Keep track of your organization's data and their classification levels."
    >
      <div className="space-y-2">
        {Array.from({ length: 5 }).map((_, i) => (
          <div
            key={i}
            className="flex items-center justify-between p-4 rounded-xl border bg-level-1"
          >
            <div className="flex items-center gap-3">
              <Skeleton className="h-8 w-8 rounded-full" />
              <div className="space-y-1">
                <Skeleton className="h-4 w-48" />
                <Skeleton className="h-3 w-32" />
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Skeleton className="h-5 w-16" />
              <Skeleton className="h-5 w-16" />
              <Skeleton className="h-8 w-8" />
              <Skeleton className="h-4 w-4" />
            </div>
          </div>
        ))}
      </div>
    </PageTemplate>
  );
}

export function DataListPage() {
  const location = useLocation();

  return (
    <Suspense key={location.pathname} fallback={<DataListViewSkeleton />}>
      <ErrorBoundaryWithLocation>
        <DataListView />
      </ErrorBoundaryWithLocation>
    </Suspense>
  );
}
