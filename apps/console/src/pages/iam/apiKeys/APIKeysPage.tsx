import { useTranslate } from "@probo/i18n";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { APIKeysPageQuery } from "#/__generated__/iam/APIKeysPageQuery.graphql";

import { PersonalAPIKeyList } from "./_components/PersonalAPIKeyList";

export const apiKeysPageQuery = graphql`
  query APIKeysPageQuery {
    viewer {
      ...PersonalAPIKeyListFragment
    }
  }
`;

export function APIKeysPage(props: {
  queryRef: PreloadedQuery<APIKeysPageQuery>;
}) {
  const { queryRef } = props;
  const { __ } = useTranslate();

  const data = usePreloadedQuery<APIKeysPageQuery>(apiKeysPageQuery, queryRef);

  return (
    <div className="space-y-6 w-full py-6">
      <h1 className="text-3xl font-bold text-center">{__("API Keys")}</h1>
      {data.viewer && <PersonalAPIKeyList fKey={data.viewer} />}
    </div>
  );
}
