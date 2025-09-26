import { Skeleton } from "@probo/ui";

export function TabSkeleton() {
  return (
    <div className="h-64">
      <Skeleton className="h-6 w-[110px] mb-1" />
      <Skeleton className="h-6 w-[250px] mb-4" />
      <Skeleton className="w-full h-[180px]" />
    </div>
  );
}
