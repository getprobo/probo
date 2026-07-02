// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();
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

  usePageTitle(__("Third parties"));

  const hasAnyAction = thirdParties.some(({ canDelete }) => canDelete);

  const thirdPartiesDocument = fragmentData.thirdPartiesDocument;
  const defaultApproverIds
    = thirdPartiesDocument?.defaultApprovers?.map(a => a.id) ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Third parties")}
        description={__(
          "Third parties are external services and providers that your company uses. Add them to keep track of their risk and compliance status.",
        )}
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
              {__("Document")}
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
                {__("Publish")}
              </Button>
            </PublishThirdPartyListDialog>
          )}
          {fragmentData.canCreateThirdParty && (
            <CreateThirdPartyDialog
              connection={connectionId}
              organizationId={organizationId}
            >
              <Button icon={IconPlusLarge}>{__("Add third party")}</Button>
            </CreateThirdPartyDialog>
          )}
        </div>
      </PageHeader>
      <SortableTable {...pagination} refetch={refetch}>
        <Thead>
          <Tr>
            <SortableTh field="NAME">{__("Third party")}</SortableTh>
            <Th>{__("Assessed At")}</Th>
            <Th>{__("Data Risk")}</Th>
            <Th>{__("Business Risk")}</Th>
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
