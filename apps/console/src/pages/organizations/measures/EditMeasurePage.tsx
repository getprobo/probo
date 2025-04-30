import { PageTemplateSkeleton } from "@/components/PageTemplate";
import { Suspense } from "react";
import { lazy } from "@probo/react-lazy";
import { useLocation } from "react-router";
import { ErrorBoundaryWithLocation } from "../ErrorBoundary";

const EditMeasureView = lazy(() => import("./EditMeasureView"));

export function EditMeasureViewSkeleton() {
  return (
    <PageTemplateSkeleton
      title="Edit Measure"
      description="Edit a Measure. You will be able to link it to control and risks"
    >
      <div className="max-w-2xl aspect-square bg-subtle-bg rounded-xl animate-pulse" />
    </PageTemplateSkeleton>
  );
}

export function EditMeasurePage() {
  const location = useLocation();

  return (
    <Suspense key={location.pathname} fallback={<EditMeasureViewSkeleton />}>
      <ErrorBoundaryWithLocation>
        <EditMeasureView />
      </ErrorBoundaryWithLocation>
    </Suspense>
  );
}
