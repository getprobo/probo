// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { usePageTitle } from "@probo/hooks";
import {
  Button,
  Card,
  IconChevronDown,
  IconPlusLarge,
  PageHeader,
  Spinner,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { ThirdPartyAgreementsPageFragment$key } from "#/__generated__/core/ThirdPartyAgreementsPageFragment.graphql";
import type { ThirdPartyAgreementsPageQuery } from "#/__generated__/core/ThirdPartyAgreementsPageQuery.graphql";
import type { ThirdPartyAgreementsPageRefetchQuery } from "#/__generated__/core/ThirdPartyAgreementsPageRefetchQuery.graphql";

import { ThirdPartyAgreementListItem } from "./_components/ThirdPartyAgreementListItem";
import { UploadThirdPartyAgreementDialog } from "./_components/UploadThirdPartyAgreementDialog";

const thirdPartyAgreementsFragment = graphql`
  fragment ThirdPartyAgreementsPageFragment on ThirdParty
  @refetchable(queryName: "ThirdPartyAgreementsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    id
    name
    canUpload: permission(action: "core:thirdParty-agreement:upload")
    agreements(first: $first, after: $after, last: $last, before: $before)
      @connection(key: "ThirdPartyAgreementsPage_agreements") {
      __id
      edges {
        node {
          id
          ...ThirdPartyAgreementListItem_agreement
        }
      }
    }
  }
`;

export const thirdPartyAgreementsPageQuery = graphql`
  query ThirdPartyAgreementsPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        ...ThirdPartyAgreementsPageFragment
      }
    }
  }
`;

interface ThirdPartyAgreementsPageProps {
  queryRef: PreloadedQuery<ThirdPartyAgreementsPageQuery>;
}

export function ThirdPartyAgreementsPage({ queryRef }: ThirdPartyAgreementsPageProps) {
  const { t } = useTranslation();
  const queryData = usePreloadedQuery<ThirdPartyAgreementsPageQuery>(
    thirdPartyAgreementsPageQuery,
    queryRef,
  );
  if (queryData.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }

  const { data, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    ThirdPartyAgreementsPageRefetchQuery,
    ThirdPartyAgreementsPageFragment$key
  >(thirdPartyAgreementsFragment, queryData.node);

  const connectionId = data.agreements.__id;
  const agreements = data.agreements.edges.map(edge => edge.node);

  usePageTitle(t("thirdPartyAgreementsPage.pageTitle", { name: data.name }));

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("thirdPartyAgreementsPage.title")}
        description={t("thirdPartyAgreementsPage.description")}
      >
        {data.canUpload && (
          <UploadThirdPartyAgreementDialog
            thirdPartyId={data.id}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>
              {t("thirdPartyAgreementsPage.actions.add")}
            </Button>
          </UploadThirdPartyAgreementDialog>
        )}
      </PageHeader>
      <Card className="space-y-4" padded>
        {agreements.length === 0 && (
          <p className="text-sm text-txt-secondary">
            {t("thirdPartyAgreementsPage.empty")}
          </p>
        )}
        {agreements.map(agreement => (
          <ThirdPartyAgreementListItem
            key={agreement.id}
            agreementKey={agreement}
            connectionId={connectionId}
          />
        ))}
        {hasNext && (
          <Button
            variant="tertiary"
            onClick={() => loadNext(50)}
            className="mx-auto"
            disabled={isLoadingNext}
            icon={isLoadingNext ? Spinner : IconChevronDown}
          >
            {t("thirdPartyAgreementsPage.actions.showMore")}
          </Button>
        )}
      </Card>
    </div>
  );
}
