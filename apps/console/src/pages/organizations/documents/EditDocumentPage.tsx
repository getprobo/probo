import { PageTemplateSkeleton } from "@/components/PageTemplate";
import { Suspense } from "react";
import { lazy } from "@probo/react-lazy";
import { useLocation } from "react-router";
import { ErrorBoundaryWithLocation } from "../ErrorBoundary";

const EditDocumentView = lazy(() => import("./EditDocumentView"));

export function EditDocumentViewSkeleton() {
  return (
    <PageTemplateSkeleton
      title="Edit Document"
      description="Edit an existing document"
    >
      <div className="bg-subtle-bg animate-pulse rounded-lg h-[600px]" />
    </PageTemplateSkeleton>
  );
}

export function EditDocumentPage() {
  const location = useLocation();

  return (
    <Suspense key={location.pathname} fallback={<EditDocumentViewSkeleton />}>
      <ErrorBoundaryWithLocation>
        <EditDocumentView />
      </ErrorBoundaryWithLocation>
    </Suspense>
  );
}
