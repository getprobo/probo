import { PageHeader, Skeleton } from "@probo/ui";

export function PageSkeleton() {
  return (
    <div className="space-y-6">
      <PageHeader title={<Skeleton className="w-40 h-8" />} />
      <div>
        <Skeleton style={{ aspectRatio: "1280/280" }} />
      </div>
    </div>
  );
}
