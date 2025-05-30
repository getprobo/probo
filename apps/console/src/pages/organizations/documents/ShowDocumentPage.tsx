import { PageTemplateSkeleton } from "@/components/PageTemplate";
import { Card, CardContent } from "@/components/ui/card";
import { Suspense } from "react";
import { useLocation } from "react-router";
import { ErrorBoundaryWithLocation } from "../ErrorBoundary";
import { lazy } from "@probo/react-lazy";

const ShowDocumentView = lazy(() => import("./ShowDocumentView"));

export function ShowDocumentViewSkeleton() {
  return (
    <PageTemplateSkeleton
      withDescription
      actions={
        <div className="flex gap-2 w-1/3">
          <div className="bg-subtle-bg animate-pulse h-8 w-1/2 rounded-lg" />
          <div className="bg-subtle-bg animate-pulse h-8 w-1/2 rounded-lg" />
        </div>
      }
    >
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <Card className="border shadow-sm">
            <CardContent className="p-6">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 rounded-md bg-slate-100/50 h-10 w-10 animate-pulse" />
                <div className="flex gap-2">
                  <div className="h-6 w-16 bg-subtle-bg animate-pulse rounded-md" />
                  <div className="h-6 w-20 bg-subtle-bg animate-pulse rounded-md" />
                  <div className="h-6 w-16 bg-subtle-bg animate-pulse rounded-md" />
                </div>
                <div className="ml-auto">
                  <div className="h-9 w-9 bg-subtle-bg animate-pulse rounded-full" />
                </div>
              </div>

              <div className="h-8 w-3/4 bg-subtle-bg animate-pulse rounded mb-3" />

              <div className="space-y-2 mb-6">
                <div className="h-4 w-full bg-subtle-bg animate-pulse rounded" />
                <div className="h-4 w-full bg-subtle-bg animate-pulse rounded" />
                <div className="h-4 w-2/3 bg-subtle-bg animate-pulse rounded" />
              </div>

              <div className="flex border-b mb-6">
                <div className="px-4 py-2 h-8 w-28 bg-subtle-bg animate-pulse rounded" />
                <div className="px-4 py-2 h-8 w-28 bg-subtle-bg animate-pulse rounded opacity-50" />
                <div className="px-4 py-2 h-8 w-28 bg-subtle-bg animate-pulse rounded opacity-50" />
              </div>

              <div className="space-y-4">
                <div className="h-4 w-full bg-subtle-bg animate-pulse rounded" />
                <div className="h-4 w-full bg-subtle-bg animate-pulse rounded" />
                <div className="h-4 w-full bg-subtle-bg animate-pulse rounded" />
                <div className="h-4 w-full bg-subtle-bg animate-pulse rounded" />
                <div className="h-4 w-3/4 bg-subtle-bg animate-pulse rounded" />
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="lg:col-span-1">
          <Card className="border shadow-sm mb-6">
            <CardContent className="p-6">
              <div className="h-6 w-32 bg-subtle-bg animate-pulse rounded mb-6" />

              <div className="space-y-6">
                <div>
                  <div className="h-4 w-28 bg-subtle-bg animate-pulse rounded mb-1" />
                  <div className="h-5 w-24 bg-subtle-bg animate-pulse rounded" />
                </div>

                <div>
                  <div className="h-4 w-28 bg-subtle-bg animate-pulse rounded mb-1" />
                  <div className="h-5 w-24 bg-subtle-bg animate-pulse rounded" />
                </div>

                <div>
                  <div className="h-4 w-28 bg-subtle-bg animate-pulse rounded mb-1" />
                  <div className="h-5 w-36 bg-subtle-bg animate-pulse rounded" />
                </div>

                <hr className="my-4" />

                <div>
                  <div className="h-5 w-32 bg-subtle-bg animate-pulse rounded mb-3" />
                  <div className="space-y-2">
                    <div className="h-4 w-48 bg-subtle-bg animate-pulse rounded" />
                    <div className="h-4 w-48 bg-subtle-bg animate-pulse rounded" />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <div className="h-10 w-full bg-subtle-bg animate-pulse rounded mb-6" />

          <Card className="border shadow-sm bg-red-50/30">
            <CardContent className="p-6">
              <div className="h-5 w-28 bg-subtle-bg animate-pulse rounded mb-3" />
              <div className="h-16 w-full bg-subtle-bg animate-pulse rounded mb-4" />
              <div className="h-10 w-full bg-subtle-bg animate-pulse rounded" />
            </CardContent>
          </Card>
        </div>
      </div>
    </PageTemplateSkeleton>
  );
}

export function ShowDocumentPage() {
  const location = useLocation();

  return (
    <Suspense key={location.pathname} fallback={<ShowDocumentViewSkeleton />}>
      <ErrorBoundaryWithLocation>
        <ShowDocumentView />
      </ErrorBoundaryWithLocation>
    </Suspense>
  );
}
