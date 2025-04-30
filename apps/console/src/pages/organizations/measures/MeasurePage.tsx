import { PageTemplateSkeleton } from "@/components/PageTemplate";
import { Card, CardContent } from "@/components/ui/card";
import { Suspense } from "react";
import { lazy } from "@probo/react-lazy";
import { useLocation } from "react-router";
import { ErrorBoundaryWithLocation } from "../ErrorBoundary";

const MeasureView = lazy(() => import("./MeasureView"));

export function MeasureViewSkeleton() {
  return (
    <PageTemplateSkeleton
      actions={
        <div className="flex items-center gap-2 w-1/3">
          <div className="bg-subtle-bg animate-pulse h-8 w-1/3 rounded-lg" />
          <div className="bg-subtle-bg animate-pulse h-8 w-1/3 rounded-lg" />
          <div className="bg-subtle-bg animate-pulse h-8 w-1/3 rounded-full" />
        </div>
      }
    >
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <Card key={i}>
            <CardContent className="p-6">
              <div className="h-6 w-48 bg-subtle-bg animate-pulse rounded mb-2" />
              <div className="h-4 w-full bg-subtle-bg animate-pulse rounded" />
            </CardContent>
          </Card>
        ))}
      </div>
    </PageTemplateSkeleton>
  );
}

export function MeasurePage() {
  const location = useLocation();

  return (
    <Suspense key={location.pathname} fallback={<MeasureViewSkeleton />}>
      <ErrorBoundaryWithLocation>
        <MeasureView />
      </ErrorBoundaryWithLocation>
    </Suspense>
  );
}
