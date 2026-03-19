import { useEffect } from "react";
import { type PreloadedQuery, usePreloadedQuery, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";

import type { ContextPageLoaderQuery } from "#/__generated__/core/ContextPageLoaderQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import ContextPage from "./ContextPage";

const contextPageQuery = graphql`
  query ContextPageLoaderQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        ...ContextPageFragment
      }
    }
  }
`;

function ContextPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<ContextPageLoaderQuery>(contextPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <LinkCardSkeleton />;

  return <ContextPageInner queryRef={queryRef} />;
}

function ContextPageInner({ queryRef }: { queryRef: PreloadedQuery<ContextPageLoaderQuery> }) {
  const data = usePreloadedQuery(contextPageQuery, queryRef);

  return <ContextPage organization={data.organization} />;
}

export default function ContextPageLoader() {
  return (
    <CoreRelayProvider>
      <ContextPageQueryLoader />
    </CoreRelayProvider>
  );
}
