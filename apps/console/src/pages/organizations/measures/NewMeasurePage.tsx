import { PageTemplateSkeleton } from "@/components/PageTemplate";
import { Suspense } from "react";
import { lazy } from "@probo/react-lazy";
import { useLocation } from "react-router";
import { ErrorBoundaryWithLocation } from "../ErrorBoundary";

const NewMeasureView = lazy(() => import("./NewMeasureView"));

export function NewMeasureViewSkeleton() {
  return (
    <PageTemplateSkeleton
      title="New Measure"
      description="Add a new measure. You will be able to link it to control and risks"
    >
      <div className="max-w-2xl aspect-square bg-subtle-bg rounded-xl animate-pulse" />
    </PageTemplateSkeleton>
  );
}

export function NewMeasurePage() {
  const location = useLocation();

  return (
    <Suspense key={location.pathname} fallback={<NewMeasureViewSkeleton />}>
      <ErrorBoundaryWithLocation>
        <NewMeasureView />
      </ErrorBoundaryWithLocation>
    </Suspense>
  );
}
