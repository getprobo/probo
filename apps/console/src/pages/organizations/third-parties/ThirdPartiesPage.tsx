// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
  IconPageTextLine,
  IconPlusLarge,
  IconUpload,
  PageHeader,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";

import type { ThirdPartiesPageFragment$key } from "#/__generated__/core/ThirdPartiesPageFragment.graphql";
import type { ThirdPartiesPageQuery } from "#/__generated__/core/ThirdPartiesPageQuery.graphql";
import type { ThirdPartiesPageRefetchQuery } from "#/__generated__/core/ThirdPartiesPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CreateThirdPartyDialog } from "./_components/CreateThirdPartyDialog";
import { PublishThirdPartyListDialog } from "./_components/PublishThirdPartyListDialog";
import { ThirdPartyRow } from "./_components/ThirdPartyRow";

export const thirdPartiesPageQuery = graphql`
  query ThirdPartiesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ...ThirdPartiesPageFragment
    }
  }
`;

const thirdPartiesFragment = graphql`
  fragment ThirdPartiesPageFragment on Organization
  @refetchable(queryName: "ThirdPartiesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ThirdPartyOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    filter: { type: "ThirdPartyFilter", defaultValue: { level: 1 } }
  ) {
    canCreateThirdParty: permission(action: "core:thirdParty:create")
    canPublishThirdParty: permission(action: "core:thirdParty:publish")
    thirdPartiesDocument {
      id
      defaultApprovers {
        id
      }
    }
    thirdParties(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "ThirdPartiesPage_thirdParties", filters: ["filter"]) {
      __id
      edges {
        node {
          id
          canDelete: permission(action: "core:thirdParty:delete")
          ...ThirdPartyRow_thirdParty
        }
      }
    }
  }
`;

export const ThirdPartiesConnectionKey = "ThirdPartiesPage_thirdParties";

// Must match the `filter` default of `ThirdPartiesPageFragment` above — the
// connection is keyed on this filter (`@connection(filters: ["filter"])`), so
// deriving its id elsewhere requires the same value.
export const ThirdPartiesConnectionFilter = { level: 1 };

interface ThirdPartiesPageProps {
  queryRef: PreloadedQuery<ThirdPartiesPageQuery>;
}

export default function ThirdPartiesPage(props: ThirdPartiesPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const queryData = usePreloadedQuery<ThirdPartiesPageQuery>(thirdPartiesPageQuery, props.queryRef);
  const { data: fragmentData, ...pagination } = usePaginationFragment<
    ThirdPartiesPageRefetchQuery,
    ThirdPartiesPageFragment$key
  >(thirdPartiesFragment, queryData.organization);

  const refetch = ({
    order,
  }: {
    order: { direction: string; field: string };
  }) => {
    pagination.refetch(
      {
        order: {
          direction: order.direction as "ASC" | "DESC",
          field: order.field as "NAME" | "CREATED_AT" | "UPDATED_AT",
        },
      },
      { fetchPolicy: "network-only" },
    );
  };

  const thirdParties = fragmentData.thirdParties?.edges.map(edge => edge.node) ?? [];
  const connectionId = fragmentData.thirdParties.__id;

  usePageTitle(t("thirdPartiesPage.title"));

  const hasAnyAction = thirdParties.some(({ canDelete }) => canDelete);

  const thirdPartiesDocument = fragmentData.thirdPartiesDocument;
  const defaultApproverIds
    = thirdPartiesDocument?.defaultApprovers?.map(a => a.id) ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("thirdPartiesPage.title")}
        description={t("thirdPartiesPage.description")}
      >
        <div className="flex gap-2">
          {thirdPartiesDocument && (
            <Button
              variant="secondary"
              icon={IconPageTextLine}
              onClick={() => void navigate(
                `/organizations/${organizationId}/documents/${thirdPartiesDocument.id}`,
              )}
            >
              {t("thirdPartiesPage.actions.document")}
            </Button>
          )}
          {fragmentData.canPublishThirdParty && (
            <PublishThirdPartyListDialog
              organizationId={organizationId}
              defaultApproverIds={defaultApproverIds}
              onPublished={documentId => void navigate(
                `/organizations/${organizationId}/documents/${documentId}`,
              )}
            >
              <Button variant="secondary" icon={IconUpload}>
                {t("thirdPartiesPage.actions.publish")}
              </Button>
            </PublishThirdPartyListDialog>
          )}
          {fragmentData.canCreateThirdParty && (
            <CreateThirdPartyDialog
              connection={connectionId}
              organizationId={organizationId}
            >
              <Button icon={IconPlusLarge}>{t("thirdPartiesPage.actions.add")}</Button>
            </CreateThirdPartyDialog>
          )}
        </div>
      </PageHeader>
      <SortableTable {...pagination} refetch={refetch}>
        <Thead>
          <Tr>
            <SortableTh field="NAME">{t("thirdPartiesPage.columns.thirdParty")}</SortableTh>
            <Th>{t("thirdPartiesPage.columns.assessedAt")}</Th>
            <Th>{t("thirdPartiesPage.columns.dataRisk")}</Th>
            <Th>{t("thirdPartiesPage.columns.businessRisk")}</Th>
            {hasAnyAction && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {thirdParties.map(thirdParty => (
            <ThirdPartyRow
              key={thirdParty.id}
              thirdPartyKey={thirdParty}
              connectionId={connectionId}
              hasAnyAction={hasAnyAction}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
