import { Suspense } from "react";
import { ErrorBoundaryWithLocation } from "../ErrorBoundary";
import { lazy } from "@probo/react-lazy";
import { PageTemplate } from "@/components/PageTemplate";
import { Skeleton } from "@/components/ui/skeleton";

const NewDatumView = lazy(() => import("./NewDatumView"));

export function NewDatumViewSkeleton() {
  return (
    <PageTemplate
      title="New Data"
      description="Add new data to your organization."
    >
      <div className="space-y-6 max-w-2xl">
        <div className="space-y-2">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-20 w-full" />
        </div>
        <div className="space-y-2">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-10 w-full" />
        </div>
        <div className="space-y-2">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-10 w-full" />
        </div>
        <Skeleton className="h-10 w-32" />
      </div>
    </PageTemplate>
  );
}
export function NewDatumPage() {
  return (
    <Suspense fallback={<NewDatumViewSkeleton />}>
      <ErrorBoundaryWithLocation>
        <NewDatumView />
      </ErrorBoundaryWithLocation>
    </Suspense>
  );
}
