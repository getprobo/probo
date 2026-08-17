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
  IconPlusLarge,
  PageHeader,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { ThirdPartyStakeholdersPageFragment$key } from "#/__generated__/core/ThirdPartyStakeholdersPageFragment.graphql";
import type { ThirdPartyStakeholdersPageQuery } from "#/__generated__/core/ThirdPartyStakeholdersPageQuery.graphql";
import type { ThirdPartyStakeholdersPageRefetchQuery } from "#/__generated__/core/ThirdPartyStakeholdersPageRefetchQuery.graphql";
import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CreateContactDialog } from "../_components/CreateContactDialog";
import { EditContactDialog } from "../_components/EditContactDialog";
import { useUpdateThirdParty } from "../_lib/useUpdateThirdParty";

import { ThirdPartyContactRow } from "./_components/ThirdPartyContactRow";

const thirdPartyStakeholdersFragment = graphql`
  fragment ThirdPartyStakeholdersPageFragment on ThirdParty
  @refetchable(queryName: "ThirdPartyStakeholdersPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ThirdPartyContactOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    id
    canCreateContact: permission(action: "core:thirdParty-contact:create")
    contacts(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "ThirdPartyStakeholdersPage_contacts") {
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

export const thirdPartyStakeholdersPageQuery = graphql`
  query ThirdPartyStakeholdersPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        id
        name
        canUpdate: permission(action: "core:thirdParty:update")
        administrators {
          id
          fullName
          emailAddress
        }
        ...ThirdPartyStakeholdersPageFragment
      }
    }
  }
`;

interface ThirdPartyStakeholdersPageProps {
  queryRef: PreloadedQuery<ThirdPartyStakeholdersPageQuery>;
}

interface AdministratorsFormValues {
  administratorIds: string[];
}

export function ThirdPartyStakeholdersPage({ queryRef }: ThirdPartyStakeholdersPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const queryData = usePreloadedQuery<ThirdPartyStakeholdersPageQuery>(
    thirdPartyStakeholdersPageQuery,
    queryRef,
  );
  if (queryData.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }
  const thirdParty = queryData.node;

  const { data, ...pagination } = usePaginationFragment<
    ThirdPartyStakeholdersPageRefetchQuery,
    ThirdPartyStakeholdersPageFragment$key
  >(thirdPartyStakeholdersFragment, thirdParty);

  const [update, isUpdating] = useUpdateThirdParty();
  const administratorIds = useMemo(
    () => thirdParty.administrators.map(administrator => administrator.id),
    [thirdParty.administrators],
  );
  const { control, reset } = useForm<AdministratorsFormValues>({
    values: { administratorIds },
  });

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

  usePageTitle(t("thirdPartyStakeholdersPage.pageTitle", { name: thirdParty.name }));

  return (
    <div className="space-y-12">
      <PageHeader
        title={t("thirdPartyStakeholdersPage.title")}
        description={t("thirdPartyStakeholdersPage.description")}
      />

      <div className="space-y-4">
        <h2 className="text-base font-medium">
          {t("thirdPartyStakeholdersPage.sections.administrators")}
        </h2>
        <Card className="space-y-4" padded>
          <PeopleMultiSelectField
            organizationId={organizationId}
            control={control}
            name="administratorIds"
            label={t("thirdPartyStakeholdersPage.fields.administrators")}
            disabled={isUpdating || !thirdParty.canUpdate}
            selectedPeople={thirdParty.administrators.map(administrator => ({
              id: administrator.id,
              fullName: administrator.fullName,
              emailAddress: administrator.emailAddress,
            }))}
            placeholder={t("thirdPartyStakeholdersPage.placeholders.administrators")}
            onIdsChange={(ids) => {
              void update(thirdParty.id, "administratorIds", ids).catch(() => {
                // The mutation already toasts; roll the selector back so it
                // keeps showing what the server actually has.
                reset({ administratorIds });
              });
            }}
          />
        </Card>
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between gap-4">
          <h2 className="text-base font-medium">
            {t("thirdPartyStakeholdersPage.sections.contacts")}
          </h2>
          {data.canCreateContact && (
            <CreateContactDialog thirdPartyId={data.id} connectionId={connectionId}>
              <Button icon={IconPlusLarge}>{t("thirdPartyStakeholdersPage.actions.addContact")}</Button>
            </CreateContactDialog>
          )}
        </div>

        <SortableTable {...pagination} refetch={refetch}>
          <Thead>
            <Tr>
              <SortableTh field="FULL_NAME">{t("thirdPartyStakeholdersPage.columns.name")}</SortableTh>
              <SortableTh field="EMAIL">{t("thirdPartyStakeholdersPage.columns.email")}</SortableTh>
              <Th>{t("thirdPartyStakeholdersPage.columns.phone")}</Th>
              <Th>{t("thirdPartyStakeholdersPage.columns.role")}</Th>
              {hasAnyAction && <Th>{t("thirdPartyStakeholdersPage.columns.actions")}</Th>}
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
      </div>

      {editingContact && editingContact.canUpdate && (
        <EditContactDialog
          contactKey={editingContact}
          onClose={() => setEditingContact(null)}
        />
      )}
    </div>
  );
}
