import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { graphql, usePaginationFragment } from "react-relay";
import { useOutletContext } from "react-router";

import type { CookieBannerConsentRecordsTabFragment$key } from "#/__generated__/core/CookieBannerConsentRecordsTabFragment.graphql";
import type { CookieBannerGraphNodeQuery$data } from "#/__generated__/core/CookieBannerGraphNodeQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

const fragment = graphql`
  fragment CookieBannerConsentRecordsTabFragment on CookieBanner
  @refetchable(queryName: "CookieBannerConsentRecordsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ConsentRecordOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    consentRecords(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    )
      @connection(
        key: "CookieBannerConsentRecordsTab_consentRecords"
      ) {
      edges {
        node {
          # eslint-disable-next-line relay/unused-fields
          id
          # eslint-disable-next-line relay/unused-fields
          visitorId
          # eslint-disable-next-line relay/unused-fields
          action
          # eslint-disable-next-line relay/unused-fields
          bannerVersion
          # eslint-disable-next-line relay/unused-fields
          ipAddress
          # eslint-disable-next-line relay/unused-fields
          createdAt
        }
      }
    }
  }
`;

function ConsentActionBadge({ action }: { action: string }) {
  const { __ } = useTranslate();

  switch (action) {
    case "ACCEPT_ALL":
      return <Badge variant="success">{__("Accept All")}</Badge>;
    case "REJECT_ALL":
      return <Badge variant="danger">{__("Reject All")}</Badge>;
    case "CUSTOMIZE":
      return <Badge variant="warning">{__("Customize")}</Badge>;
    default:
      return <Badge>{action}</Badge>;
  }
}

export default function CookieBannerConsentRecordsTab() {
  const { banner } = useOutletContext<{
    banner: CookieBannerGraphNodeQuery$data["node"];
  }>();

  const { __ } = useTranslate();
  const pagination = usePaginationFragment(
    fragment,
    banner as CookieBannerConsentRecordsTabFragment$key,
  );

  const records
    = pagination.data.consentRecords?.edges.map(edge => edge.node) ?? [];

  return (
    <div className="space-y-6">
      {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
      <SortableTable {...(pagination as any)}>
        <Thead>
          <Tr>
            <Th>{__("Visitor ID")}</Th>
            <Th>{__("Action")}</Th>
            <Th>{__("Banner Version")}</Th>
            <Th>{__("IP Address")}</Th>
            <SortableTh field="CREATED_AT">{__("Date")}</SortableTh>
          </Tr>
        </Thead>
        <Tbody>
          {records.map(record => (
            <Tr key={record.id}>
              <Td>
                <span className="font-mono text-sm">
                  {record.visitorId.slice(0, 12)}...
                </span>
              </Td>
              <Td>
                <ConsentActionBadge action={record.action} />
              </Td>
              <Td>v{record.bannerVersion}</Td>
              <Td>{record.ipAddress ?? "-"}</Td>
              <Td>{formatDate(record.createdAt)}</Td>
            </Tr>
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
