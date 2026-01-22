import { graphql, usePreloadedQuery, type PreloadedQuery } from "react-relay";
import { Outlet } from "react-router";
import { Layout, Skeleton } from "@probo/ui";
import type { ViewerLayoutQuery } from "/__generated__/iam/ViewerLayoutQuery.graphql";
import { Suspense } from "react";
import { ViewerDropdown } from "./_components/ViewerDropdown";

export const viewerLayoutQuery = graphql`
  query ViewerLayoutQuery {
    viewer @required(action: THROW) {
      ...ViewerDropdownFragment
    }
  }
`;

export function ViewerLayout(props: {
  hideSidebar?: boolean;
  queryRef: PreloadedQuery<ViewerLayoutQuery>;
}) {
  const { queryRef } = props;

  const { viewer } = usePreloadedQuery<ViewerLayoutQuery>(
    viewerLayoutQuery,
    queryRef,
  );

  return (
    <Layout
      headerTrailing={(
        <div className="ml-auto">
          <Suspense fallback={<Skeleton className="w-32 h-8" />}>
            <ViewerDropdown fKey={viewer} />
          </Suspense>
        </div>
      )}
    >
      <Outlet />
    </Layout>
  );
}
