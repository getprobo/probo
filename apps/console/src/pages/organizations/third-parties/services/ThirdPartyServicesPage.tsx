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
import { useState } from "react";
import { graphql, type PreloadedQuery, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { ThirdPartyServicesPageFragment$key } from "#/__generated__/core/ThirdPartyServicesPageFragment.graphql";
import type { ThirdPartyServicesPageQuery } from "#/__generated__/core/ThirdPartyServicesPageQuery.graphql";
import type { ThirdPartyServicesPageRefetchQuery } from "#/__generated__/core/ThirdPartyServicesPageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { CreateServiceDialog } from "../_components/CreateServiceDialog";
import { EditServiceDialog } from "../_components/EditServiceDialog";

import { ThirdPartyServiceRow } from "./_components/ThirdPartyServiceRow";

const thirdPartyServicesFragment = graphql`
  fragment ThirdPartyServicesPageFragment on ThirdParty
  @refetchable(queryName: "ThirdPartyServicesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ThirdPartyServiceOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    id
    name
    canCreateService: permission(action: "core:thirdParty-service:create")
    services(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "ThirdPartyServicesPage_services") {
      __id
      edges {
        node {
          id
          canUpdate: permission(action: "core:thirdParty-service:update")
          canDelete: permission(action: "core:thirdParty-service:delete")
          ...ThirdPartyServiceRow_service
          ...EditServiceDialog_service
        }
      }
    }
  }
`;

export const thirdPartyServicesPageQuery = graphql`
  query ThirdPartyServicesPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        ...ThirdPartyServicesPageFragment
      }
    }
  }
`;

interface ThirdPartyServicesPageProps {
  queryRef: PreloadedQuery<ThirdPartyServicesPageQuery>;
}

export default function ThirdPartyServicesPage(props: ThirdPartyServicesPageProps) {
  const { __ } = useTranslate();
  const queryData = usePreloadedQuery<ThirdPartyServicesPageQuery>(thirdPartyServicesPageQuery, props.queryRef);
  if (queryData.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }

  const { data, ...pagination } = usePaginationFragment<
    ThirdPartyServicesPageRefetchQuery,
    ThirdPartyServicesPageFragment$key
  >(thirdPartyServicesFragment, queryData.node);

  const refetch = ({
    order,
  }: {
    order: { direction: string; field: string };
  }) => {
    pagination.refetch(
      {
        order: {
          direction: order.direction as "ASC" | "DESC",
          field: order.field as "NAME" | "CREATED_AT",
        },
      },
      { fetchPolicy: "network-only" },
    );
  };

  const connectionId = data.services.__id;
  const services = data.services.edges.map(edge => edge.node);
  const [editingService, setEditingService]
    = useState<(typeof services)[number] | null>(null);
  const hasAnyAction = services.some(
    service => service.canUpdate || service.canDelete,
  );

  usePageTitle(data.name + " - " + __("Services"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Services")}
        description={__("Manage services provided by this third party.")}
      >
        {data.canCreateService && (
          <CreateServiceDialog thirdPartyId={data.id} connectionId={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add service")}</Button>
          </CreateServiceDialog>
        )}
      </PageHeader>

      <SortableTable {...pagination} refetch={refetch}>
        <Thead>
          <Tr>
            <SortableTh field="NAME">{__("Name")}</SortableTh>
            <Th>{__("Description")}</Th>
            {hasAnyAction && <Th>{__("Actions")}</Th>}
          </Tr>
        </Thead>
        <Tbody>
          {services.map(service => (
            <ThirdPartyServiceRow
              key={service.id}
              serviceKey={service}
              connectionId={connectionId}
              onEdit={() => setEditingService(service)}
            />
          ))}
        </Tbody>
      </SortableTable>

      {editingService && editingService.canUpdate && (
        <EditServiceDialog
          serviceKey={editingService}
          onClose={() => setEditingService(null)}
        />
      )}
    </div>
  );
}
