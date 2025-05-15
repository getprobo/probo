import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { ControlViewSkeleton } from "./ControlPage";
import { Suspense, useEffect } from "react";
import { type ControlViewQuery } from "./__generated__/ControlViewQuery.graphql";
import { useParams } from "react-router";
import { Control } from "./Control";

const controlViewQuery = graphql`
  query ControlViewQuery($controlId: ID!) {
    node(id: $controlId) {
      id
      ...ControlFragment_Control
    }
  }
`;

function ControlViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<ControlViewQuery>;
}) {
  const { node } = usePreloadedQuery<ControlViewQuery>(
    controlViewQuery,
    queryRef,
  );

  return <Control controlKey={node} />;
}

export default function ControlView({ controlId }: { controlId?: string }) {
  const { controlId: controlIdParam } = useParams();
  const [queryRef, loadQuery] =
    useQueryLoader<ControlViewQuery>(controlViewQuery);

  useEffect(() => {
    loadQuery({ controlId: (controlId ?? controlIdParam)! });
  }, [loadQuery, controlId, controlIdParam]);

  if (!queryRef) return <ControlViewSkeleton />;

  return (
    <Suspense>
      <ControlViewContent queryRef={queryRef} />
    </Suspense>
  );
}
