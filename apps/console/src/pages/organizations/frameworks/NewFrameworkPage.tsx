import { PageTemplateSkeleton } from "@/components/PageTemplate";
import { Suspense } from "react";
import { lazy } from "@probo/react-lazy";
import { useLocation } from "react-router";
import { ErrorBoundaryWithLocation } from "../ErrorBoundary";

const NewFrameworkView = lazy(() => import("./NewFrameworkView"));

export function NewFrameworkViewSkeleton() {
  return (
    <PageTemplateSkeleton
      title="Create Framework"
      description="Create a new framework to organize your measures"
    >
      <div className="max-w-2xl aspect-square bg-subtle-bg rounded-xl animate-pulse" />
    </PageTemplateSkeleton>
  );
}

export function NewFrameworkPage() {
  const location = useLocation();

  return (
    <Suspense key={location.pathname} fallback={<NewFrameworkViewSkeleton />}>
      <ErrorBoundaryWithLocation>
        <NewFrameworkView />
      </ErrorBoundaryWithLocation>
    </Suspense>
  );
}
