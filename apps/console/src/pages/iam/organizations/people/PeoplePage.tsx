import { useTranslate } from "@probo/i18n";
import { PageHeader } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { PeoplePageQuery } from "#/__generated__/iam/PeoplePageQuery.graphql";

import { PeopleList } from "./_components/PeopleList";

export const peoplePageQuery = graphql`
  query PeoplePageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ...PeopleListFragment
        @arguments(first: 20, order: { direction: ASC, field: FULL_NAME })
    }
  }
`;

export function PeoplePage(props: {
  queryRef: PreloadedQuery<PeoplePageQuery>;
}) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<PeoplePageQuery>(
    peoplePageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("node is of invalid type");
  }

  return (
    <div className="space-y-6">
      <PageHeader title={__("People")} />

      <div className="pb-6 pt-6">
        <PeopleList fKey={organization} />
      </div>
    </div>
  );
}
