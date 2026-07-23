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
  IconPlusLarge,
  PageHeader,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { ThirdPartyContactsPageFragment$key } from "#/__generated__/core/ThirdPartyContactsPageFragment.graphql";
import type { ThirdPartyContactsPageQuery } from "#/__generated__/core/ThirdPartyContactsPageQuery.graphql";
import type { ThirdPartyContactsPageRefetchQuery } from "#/__generated__/core/ThirdPartyContactsPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { CreateContactDialog } from "../_components/CreateContactDialog";
import { EditContactDialog } from "../_components/EditContactDialog";

import { ThirdPartyContactRow } from "./_components/ThirdPartyContactRow";

const thirdPartyContactsFragment = graphql`
  fragment ThirdPartyContactsPageFragment on ThirdParty
  @refetchable(queryName: "ThirdPartyContactsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ThirdPartyContactOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    id
    name
    canCreateContact: permission(action: "core:thirdParty-contact:create")
    contacts(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "ThirdPartyContactsPage_contacts") {
      __id
      edges {
        node {
          id
          canUpdate: permission(action: "core:thirdParty-contact:update")
          canDelete: permission(action: "core:thirdParty-contact:delete")
          ...ThirdPartyContactRow_contact
          ...EditContactDialog_contact
        }
      }
    }
  }
`;

export const thirdPartyContactsPageQuery = graphql`
  query ThirdPartyContactsPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        ...ThirdPartyContactsPageFragment
      }
    }
  }
`;

interface ThirdPartyContactsPageProps {
  queryRef: PreloadedQuery<ThirdPartyContactsPageQuery>;
}

export default function ThirdPartyContactsPage(props: ThirdPartyContactsPageProps) {
  const { t } = useTranslation();
  const queryData = usePreloadedQuery<ThirdPartyContactsPageQuery>(thirdPartyContactsPageQuery, props.queryRef);
  if (queryData.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }

  const { data, ...pagination } = usePaginationFragment<
    ThirdPartyContactsPageRefetchQuery,
    ThirdPartyContactsPageFragment$key
  >(thirdPartyContactsFragment, queryData.node);

  const refetch = ({
    order,
  }: {
    order: { direction: string; field: string };
  }) => {
    pagination.refetch(
      {
        order: {
          direction: order.direction as "ASC" | "DESC",
          field: order.field as "FULL_NAME" | "EMAIL" | "CREATED_AT",
        },
      },
      { fetchPolicy: "network-only" },
    );
  };

  const connectionId = data.contacts.__id;
  const contacts = data.contacts.edges.map(edge => edge.node);
  const [editingContact, setEditingContact]
    = useState<(typeof contacts)[number] | null>(null);
  const hasAnyAction = contacts.some(
    contact => contact.canUpdate || contact.canDelete,
  );

  usePageTitle(t("thirdPartyContactsPage.pageTitle", { name: data.name }));

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("thirdPartyContactsPage.title")}
        description={t("thirdPartyContactsPage.description")}
      >
        {data.canCreateContact && (
          <CreateContactDialog thirdPartyId={data.id} connectionId={connectionId}>
            <Button icon={IconPlusLarge}>{t("thirdPartyContactsPage.actions.add")}</Button>
          </CreateContactDialog>
        )}
      </PageHeader>

      <SortableTable {...pagination} refetch={refetch}>
        <Thead>
          <Tr>
            <SortableTh field="FULL_NAME">{t("thirdPartyContactsPage.columns.name")}</SortableTh>
            <SortableTh field="EMAIL">{t("thirdPartyContactsPage.columns.email")}</SortableTh>
            <Th>{t("thirdPartyContactsPage.columns.phone")}</Th>
            <Th>{t("thirdPartyContactsPage.columns.role")}</Th>
            {hasAnyAction && <Th>{t("thirdPartyContactsPage.columns.actions")}</Th>}
          </Tr>
        </Thead>
        <Tbody>
          {contacts.map(contact => (
            <ThirdPartyContactRow
              key={contact.id}
              contactKey={contact}
              connectionId={connectionId}
              onEdit={() => setEditingContact(contact)}
            />
          ))}
        </Tbody>
      </SortableTable>

      {editingContact && editingContact.canUpdate && (
        <EditContactDialog
          contactKey={editingContact}
          onClose={() => setEditingContact(null)}
        />
      )}
    </div>
  );
}
