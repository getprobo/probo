import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { useParams } from "react-router";
import { type FrameworkViewQuery } from "./__generated__/FrameworkViewQuery.graphql";
import { Suspense, useEffect } from "react";
import { Control } from "./controls/Control";
import { FrameworkViewSkeleton } from "./FrameworkPage";

const frameworkViewQuery = graphql`
  query FrameworkViewQuery($frameworkId: ID!) {
    node(id: $frameworkId) {
      id
      ... on Framework {
        name
        description
        firstControl: controls(
          first: 1
          orderBy: { field: CREATED_AT, direction: ASC }
        ) @connection(key: "FrameworkView_firstControl") {
          edges {
            node {
              ...ControlFragment_Control
            }
          }
        }
      }
    }
  }
`;

function FrameworkViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<FrameworkViewQuery>;
}) {
  const data = usePreloadedQuery<FrameworkViewQuery>(
    frameworkViewQuery,
    queryRef,
  );

  if (!data?.node) {
    return <div>Framework not found</div>;
  }

  if (!data.node.firstControl?.edges?.length) {
    return <div>No controls found for this framework</div>;
  }

  return <Control controlKey={data.node.firstControl.edges[0].node} />;
}

export default function FrameworkView() {
  const { frameworkId } = useParams();
  const [queryRef, loadQuery] =
    useQueryLoader<FrameworkViewQuery>(frameworkViewQuery);

  useEffect(() => {
    loadQuery({ frameworkId: frameworkId! });
  }, [loadQuery, frameworkId]);

  if (!queryRef) {
    return <FrameworkViewSkeleton />;
  }

  return (
    <Suspense fallback={<FrameworkViewSkeleton />}>
      <FrameworkViewContent queryRef={queryRef} />
    </Suspense>
  );
}
