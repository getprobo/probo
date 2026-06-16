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
  IconPlusLarge,
  PageHeader,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import type { ComponentProps } from "react";
import { useState } from "react";
import { graphql, type PreloadedQuery, useRefetchableFragment, usePreloadedQuery } from "react-relay";

import type { ThirdPartyContactRow_contact$data } from "#/__generated__/core/ThirdPartyContactRow_contact.graphql";
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
  const queryData = usePreloadedQuery(thirdPartyContactsPageQuery, props.queryRef);
  if (queryData.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }

  const [data, refetch] = useRefetchableFragment<
    ThirdPartyContactsPageRefetchQuery,
    ThirdPartyContactsPageFragment$key
  >(thirdPartyContactsFragment, queryData.node);

  const connectionId = data.contacts.__id;
  const contacts = data.contacts.edges.map(edge => edge.node);
  const { __ } = useTranslate();
  const [editingContact, setEditingContact]
    = useState<ThirdPartyContactRow_contact$data | null>(null);
  const hasAnyAction = contacts.some(
    contact => contact.canUpdate || contact.canDelete,
  );

  usePageTitle(data.name + " - " + __("Contacts"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Contacts")}
        description={__("Manage third party contacts and their information.")}
      >
        {data.canCreateContact && (
          <CreateContactDialog thirdPartyId={data.id} connectionId={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add contact")}</Button>
          </CreateContactDialog>
        )}
      </PageHeader>

      <SortableTable
        refetch={refetch as ComponentProps<typeof SortableTable>["refetch"]}
      >
        <Thead>
          <Tr>
            <SortableTh field="FULL_NAME">{__("Name")}</SortableTh>
            <SortableTh field="EMAIL">{__("Email")}</SortableTh>
            <Th>{__("Phone")}</Th>
            <Th>{__("Role")}</Th>
            {hasAnyAction && <Th>{__("Actions")}</Th>}
          </Tr>
        </Thead>
        <Tbody>
          {contacts.map(contact => (
            <ThirdPartyContactRow
              key={contact.id}
              contactKey={contact}
              connectionId={connectionId}
              onEdit={setEditingContact}
            />
          ))}
        </Tbody>
      </SortableTable>

      {editingContact && editingContact.canUpdate && (
        <EditContactDialog
          contactId={editingContact.id}
          contact={editingContact}
          onClose={() => setEditingContact(null)}
        />
      )}
    </div>
  );
}
