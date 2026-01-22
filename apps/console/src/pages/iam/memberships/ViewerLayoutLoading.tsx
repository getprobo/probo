import { Layout, Skeleton } from "@probo/ui";
import { Outlet } from "react-router";

export function ViewerLayoutLoading() {
  return (
    <Layout
      headerTrailing={(
        <div className="ml-auto">
          <Skeleton className="w-32 h-8" />
        </div>
      )}
    >
      <Outlet />
    </Layout>
  );
}
