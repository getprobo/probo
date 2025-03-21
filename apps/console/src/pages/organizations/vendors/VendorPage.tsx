import { PageTemplateSkeleton } from "@/components/PageTemplate";
import { lazy, Suspense } from "react";
import { useLocation } from "react-router";
import { ErrorBoundaryWithLocation } from "../ErrorBoundary";

const VendorView = lazy(() => import("./VendorView"));

export function VendorViewSkeleton() {
  return (
    <PageTemplateSkeleton>
      <div className="space-y-2">
        {[1, 2].map((i) => (
          <div key={i} className="h-20 bg-muted animate-pulse rounded-lg" />
        ))}
      </div>
    </PageTemplateSkeleton>
  );
}

export function VendorPage() {
  const location = useLocation();

  return (
    <Suspense key={location.pathname} fallback={<VendorViewSkeleton />}>
      <ErrorBoundaryWithLocation>
        <VendorView />
      </ErrorBoundaryWithLocation>
    </Suspense>
  );
}
