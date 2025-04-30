import { PageTemplateSkeleton } from "@/components/PageTemplate";
import { Card, CardContent, CardFooter } from "@/components/ui/card";
import { Suspense } from "react";
import { useLocation } from "react-router";
import { ErrorBoundaryWithLocation } from "../ErrorBoundary";
import { lazy } from "@probo/react-lazy";

const PolicyListView = lazy(() => import("./PolicyListView"));

export function PolicyListViewSkeleton() {
  return (
    <PageTemplateSkeleton
      title="Policies"
      description="Manage your organization's policies"
      actions={
        <div className="bg-subtle-bg animate-pulse h-9 w-1/6 rounded-lg" />
      }
    >
      {/* Search and filter measures skeleton */}
      <div className="flex flex-col md:flex-row gap-4 mb-6">
        <div className="flex-1 h-10 bg-subtle-bg animate-pulse rounded" />
        <div className="flex gap-2">
          <div className="h-10 w-[180px] bg-subtle-bg animate-pulse rounded" />
          <div className="h-10 w-24 bg-subtle-bg animate-pulse rounded" />
        </div>
      </div>

      {/* Results summary skeleton */}
      <div className="mb-4 h-4 w-48 bg-subtle-bg animate-pulse rounded" />

      {/* Policy grid skeleton */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <Card key={i} className="bg-level-1/50 h-full flex flex-col">
            <CardContent className="p-6 flex-grow">
              <div className="flex justify-between items-start mb-3">
                <div className="h-7 w-48 bg-subtle-bg animate-pulse rounded" />
                <div className="h-6 w-20 bg-subtle-bg animate-pulse rounded-full" />
              </div>
              <div className="space-y-2 mb-4">
                <div className="h-4 w-full bg-subtle-bg animate-pulse rounded" />
                <div className="h-4 w-full bg-subtle-bg animate-pulse rounded" />
                <div className="h-4 w-2/3 bg-subtle-bg animate-pulse rounded" />
              </div>
              <div className="mt-auto space-y-2">
                <div className="h-4 w-40 bg-subtle-bg animate-pulse rounded" />
                <div className="h-4 w-40 bg-subtle-bg animate-pulse rounded" />
              </div>
            </CardContent>
            <CardFooter className="p-0 border-t">
              <div className="w-full grid grid-cols-2">
                <div className="h-12 border-r bg-subtle-bg/20 animate-pulse" />
                <div className="h-12 bg-subtle-bg/20 animate-pulse" />
              </div>
            </CardFooter>
          </Card>
        ))}
      </div>
    </PageTemplateSkeleton>
  );
}

export function PolicyListPage() {
  const location = useLocation();

  return (
    <Suspense key={location.pathname} fallback={<PolicyListViewSkeleton />}>
      <ErrorBoundaryWithLocation>
        <PolicyListView />
      </ErrorBoundaryWithLocation>
    </Suspense>
  );
}
