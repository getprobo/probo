import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { ContextPageLoaderQuery } from "#/__generated__/core/ContextPageLoaderQuery.graphql";

import ContextPage from "./ContextPage";

export const contextPageQuery = graphql`
  query ContextPageLoaderQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        ...ContextPageFragment
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<ContextPageLoaderQuery>;
};

export default function ContextPageLoader(props: Props) {
  const data = usePreloadedQuery(contextPageQuery, props.queryRef);

  return <ContextPage organization={data.organization} />;
}
